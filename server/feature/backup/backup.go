package backup

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database"
	"github.com/sbondCo/Watcharr/database/entity"
	"gorm.io/gorm"
)

const maxDecompressedBackupBytes = 500 << 20 // 500 MB max decompressed size

type Service struct {
	db             *gorm.DB
	cfg            *config.ServerConfig
	backupDir      string
	mu             sync.Mutex
	rescheduleTask func(seconds int) error
}

// SetTaskRescheduler wires task.RescheduleTask for "Automated Database Backup" without creating package cycles.
func (s *Service) SetTaskRescheduler(fn func(seconds int) error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rescheduleTask = fn
}

func NewService(db *gorm.DB, cfg *config.ServerConfig) *Service {
	bDir := filepath.Join(config.DataPath, "backups")
	s := &Service{
		db:        db,
		cfg:       cfg,
		backupDir: bDir,
	}
	if err := s.EnsureBackupDir(); err != nil {
		slog.Error("NewService: failed to ensure backup directory", "error", err)
	}
	return s
}

// EnsureBackupDir guarantees the backup directory exists
func (s *Service) EnsureBackupDir() error {
	return os.MkdirAll(s.backupDir, 0755)
}

// FormatBytes formats a byte count into a human-readable string (KB, MB, GB)
func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func (s *Service) validateFilename(filename string) error {
	cleaned := filepath.Base(filename)
	if cleaned != filename || strings.Contains(filename, "..") || strings.ContainsAny(filename, `/\`) {
		return errors.New("invalid backup filename")
	}
	if !strings.HasSuffix(filename, ".db.gz") &&
		!strings.HasSuffix(filename, ".db") &&
		!strings.HasSuffix(filename, ".json.gz") &&
		!strings.HasSuffix(filename, ".json") {
		return errors.New("unsupported backup file extension")
	}
	return nil
}

// ListBackups returns all available backups sorted by creation date (newest first)
func (s *Service) ListBackups() ([]BackupInfo, error) {
	if err := s.EnsureBackupDir(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(s.backupDir)
	if err != nil {
		return nil, err
	}

	backups := make([]BackupInfo, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if s.validateFilename(name) != nil {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		bType := "sqlite"
		if strings.Contains(name, ".json") {
			bType = "json"
		}

		isAuto := strings.Contains(name, "_auto_")

		backups = append(backups, BackupInfo{
			Filename:      name,
			Size:          info.Size(),
			FormattedSize: FormatBytes(info.Size()),
			CreatedAt:     info.ModTime(),
			IsAuto:        isAuto,
			Type:          bType,
		})
	}

	sort.Slice(backups, func(i, j int) bool {
		return backups[i].CreatedAt.After(backups[j].CreatedAt)
	})

	return backups, nil
}

// CreateBackup generates a clean, compressed snapshot of the database
func (s *Service) CreateBackup(isAuto bool, format string) (*BackupInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if isAuto && !s.cfg.GetBackupEnabled() {
		slog.Info("CreateBackup: Automated backup skipped because backup is disabled in config")
		return nil, nil
	}

	if err := s.EnsureBackupDir(); err != nil {
		return nil, fmt.Errorf("failed to ensure backup directory: %w", err)
	}

	tag := "manual"
	if isAuto {
		tag = "auto"
	}
	ts := time.Now().Format("20060102_150405")

	dbType := s.db.Dialector.Name()
	targetFormat := strings.ToLower(strings.TrimSpace(format))
	if targetFormat == "" || targetFormat == "auto" {
		if dbType == "sqlite" {
			targetFormat = "sqlite"
		} else {
			targetFormat = "json"
		}
	}

	var filename string
	var err error

	if targetFormat == "sqlite" && dbType == "sqlite" {
		filename = fmt.Sprintf("watcharr_backup_%s_%s.db.gz", tag, ts)
		err = s.createSQLiteBackup(filename)
	} else {
		filename = fmt.Sprintf("watcharr_backup_%s_%s.json.gz", tag, ts)
		err = s.createJSONBackup(filename)
	}

	if err != nil {
		slog.Error("CreateBackup: Backup generation failed", "filename", filename, "error", err)
		return nil, err
	}

	targetPath := filepath.Join(s.backupDir, filename)
	fi, err := os.Stat(targetPath)
	if err != nil {
		return nil, err
	}

	slog.Info("CreateBackup: Backup created successfully", "filename", filename, "size", FormatBytes(fi.Size()))

	if isAuto {
		if err := s.pruneBackupsLocked(s.cfg.GetBackupRetention()); err != nil {
			slog.Warn("CreateBackup: failed to prune old backups", "error", err)
		}
	}

	bType := "sqlite"
	if strings.Contains(filename, ".json") {
		bType = "json"
	}

	return &BackupInfo{
		Filename:      filename,
		Size:          fi.Size(),
		FormattedSize: FormatBytes(fi.Size()),
		CreatedAt:     fi.ModTime(),
		IsAuto:        isAuto,
		Type:          bType,
	}, nil
}

// RunAutomatedBackup fulfills the BackupProvider interface for task scheduling
func (s *Service) RunAutomatedBackup() error {
	_, err := s.CreateBackup(true, "auto")
	return err
}

func (s *Service) createSQLiteBackup(finalGzFilename string) error {
	// 1. Truncate WAL to write all active transactions to disk
	if res := s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); res.Error != nil {
		slog.Warn("createSQLiteBackup: PRAGMA wal_checkpoint failed", "error", res.Error)
	}

	tempRaw := filepath.Join(s.backupDir, "raw_"+finalGzFilename+".tmp")
	defer os.Remove(tempRaw)

	// Try VACUUM INTO for zero-lock, defragmented consistent snapshot
	escapedTempRaw := strings.ReplaceAll(tempRaw, "'", "''")
	vacRes := s.db.Exec(fmt.Sprintf("VACUUM INTO '%s'", escapedTempRaw))
	if vacRes.Error != nil {
		slog.Warn("createSQLiteBackup: VACUUM INTO failed, falling back to direct copy", "error", vacRes.Error)
		srcPath := s.getActiveSQLiteDBPath()
		if err := copyFile(srcPath, tempRaw); err != nil {
			return fmt.Errorf("failed to copy sqlite source db: %w", err)
		}
	}

	// Compress tempRaw into final gz
	finalPath := filepath.Join(s.backupDir, finalGzFilename)
	if err := gzipFile(tempRaw, finalPath); err != nil {
		_ = os.Remove(finalPath)
		return fmt.Errorf("failed to gzip backup: %w", err)
	}

	return nil
}

func (s *Service) getActiveSQLiteDBPath() string {
	type dbListEntry struct {
		Seq  int
		Name string
		File string
	}
	var entries []dbListEntry
	if err := s.db.Raw("PRAGMA database_list").Scan(&entries).Error; err == nil {
		for _, e := range entries {
			if e.Name == "main" && e.File != "" {
				return e.File
			}
		}
	}
	return database.GetSQLiteDBPath()
}

func (s *Service) createJSONBackup(finalGzFilename string) error {
	dump := UniversalBackupDump{
		Version:        "1.0",
		ExportedAt:     time.Now(),
		Driver:         s.db.Dialector.Name(),
		PasswordHashes: make(map[uint]string),
	}

	var users []entity.User
	if err := s.db.Find(&users).Error; err != nil {
		return fmt.Errorf("createJSONBackup: failed to dump users: %w", err)
	}
	type userPassRow struct {
		ID       uint
		Password string
	}
	var passRows []userPassRow
	if err := s.db.Table("users").Select("id, password").Scan(&passRows).Error; err != nil {
		return fmt.Errorf("createJSONBackup: failed to dump user password hashes: %w", err)
	}
	for _, pr := range passRows {
		dump.PasswordHashes[pr.ID] = pr.Password
	}
	dump.Users = make([]UserBackup, len(users))
	for i, u := range users {
		pwHash := dump.PasswordHashes[u.ID]
		u.Password = pwHash
		dump.Users[i] = UserBackup{
			User:         u,
			PasswordHash: pwHash,
			AvatarID:     u.AvatarID,
			Permissions:  u.Permissions,
		}
	}

	var userServices []entity.UserServices
	if err := s.db.Find(&userServices).Error; err != nil {
		return fmt.Errorf("createJSONBackup: failed to dump user services: %w", err)
	}
	dump.UserServices = make([]UserServicesBackup, len(userServices))
	for i, us := range userServices {
		dump.UserServices[i] = UserServicesBackup{
			UserServices: us,
			Name:         us.Name,
			ClientID:     us.ClientID,
			AuthToken:    us.AuthToken,
			AuthToken2:   us.AuthToken2,
			UserID:       us.UserID,
		}
	}

	if err := s.db.Find(&dump.Content).Error; err != nil {
		return fmt.Errorf("createJSONBackup: failed to dump content: %w", err)
	}

	var watchedList []entity.Watched
	if err := s.db.Find(&watchedList).Error; err != nil {
		return fmt.Errorf("createJSONBackup: failed to dump watched: %w", err)
	}
	dump.Watched = make([]WatchedBackup, len(watchedList))
	for i, w := range watchedList {
		dump.Watched[i] = WatchedBackup{
			Watched:   w,
			UserID:    w.UserID,
			ContentID: w.ContentID,
			GameID:    w.GameID,
		}
	}

	if err := s.db.Table("watched_tags").Select("watched_id, tag_id").Scan(&dump.WatchedTags).Error; err != nil {
		return fmt.Errorf("createJSONBackup: failed to dump watched_tags: %w", err)
	}

	var seasons []entity.WatchedSeason
	if err := s.db.Find(&seasons).Error; err != nil {
		return fmt.Errorf("createJSONBackup: failed to dump seasons: %w", err)
	}
	dump.Seasons = make([]WatchedSeasonBackup, len(seasons))
	for i, ws := range seasons {
		dump.Seasons[i] = WatchedSeasonBackup{
			WatchedSeason: ws,
			UserID:        ws.UserID,
			WatchedID:     ws.WatchedID,
		}
	}

	var episodes []entity.WatchedEpisode
	if err := s.db.Find(&episodes).Error; err != nil {
		return fmt.Errorf("createJSONBackup: failed to dump episodes: %w", err)
	}
	dump.Episodes = make([]WatchedEpisodeBackup, len(episodes))
	for i, we := range episodes {
		dump.Episodes[i] = WatchedEpisodeBackup{
			WatchedEpisode: we,
			UserID:         we.UserID,
			WatchedID:      we.WatchedID,
		}
	}

	var activities []entity.Activity
	if err := s.db.Find(&activities).Error; err != nil {
		return fmt.Errorf("createJSONBackup: failed to dump activity: %w", err)
	}
	dump.Activity = make([]ActivityBackup, len(activities))
	for i, a := range activities {
		dump.Activity[i] = ActivityBackup{
			Activity: a,
			UserID:   a.UserID,
		}
	}

	var follows []entity.Follow
	if err := s.db.Find(&follows).Error; err != nil {
		return fmt.Errorf("createJSONBackup: failed to dump follow: %w", err)
	}
	dump.Follow = make([]FollowBackup, len(follows))
	for i, f := range follows {
		dump.Follow[i] = FollowBackup{
			Follow:         f,
			UpdatedAt:      f.UpdatedAt,
			UserID:         f.UserID,
			FollowedUserID: f.FollowedUserID,
		}
	}

	var images []entity.Image
	if err := s.db.Find(&images).Error; err != nil {
		return fmt.Errorf("createJSONBackup: failed to dump images: %w", err)
	}
	dump.Images = make([]ImageBackup, len(images))
	for i, img := range images {
		dump.Images[i] = ImageBackup{
			Image: img,
			ID:    img.ID,
			Hash:  img.Hash,
		}
	}

	var games []entity.Game
	if err := s.db.Find(&games).Error; err != nil {
		return fmt.Errorf("createJSONBackup: failed to dump games: %w", err)
	}
	dump.Games = make([]GameBackup, len(games))
	for i, g := range games {
		dump.Games[i] = GameBackup{
			Game:     g,
			PosterID: g.PosterID,
		}
	}

	var arrRequests []entity.ArrRequest
	if err := s.db.Find(&arrRequests).Error; err != nil {
		return fmt.Errorf("createJSONBackup: failed to dump arr requests: %w", err)
	}
	dump.ArrRequests = make([]ArrRequestBackup, len(arrRequests))
	for i, ar := range arrRequests {
		dump.ArrRequests[i] = ArrRequestBackup{
			ArrRequest: ar,
			UserID:     ar.UserID,
			ContentID:  ar.ContentID,
		}
	}

	var tags []entity.Tag
	if err := s.db.Find(&tags).Error; err != nil {
		return fmt.Errorf("createJSONBackup: failed to dump tags: %w", err)
	}
	dump.Tags = make([]TagBackup, len(tags))
	for i, t := range tags {
		dump.Tags[i] = TagBackup{
			Tag:    t,
			UserID: t.UserID,
		}
	}

	if err := s.db.Find(&dump.PushSubscriptions).Error; err != nil {
		return fmt.Errorf("createJSONBackup: failed to dump push subscriptions: %w", err)
	}

	jsonData, err := json.Marshal(dump)
	if err != nil {
		return fmt.Errorf("failed to marshal universal backup: %w", err)
	}

	finalPath := filepath.Join(s.backupDir, finalGzFilename)
	out, err := os.Create(finalPath)
	if err != nil {
		return err
	}
	defer out.Close()

	gw := gzip.NewWriter(out)
	defer gw.Close()

	if _, err := gw.Write(jsonData); err != nil {
		return err
	}

	return nil
}

// PruneBackups deletes older automatic backups exceeding the retention threshold
func (s *Service) PruneBackups(retention int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pruneBackupsLocked(retention)
}

func (s *Service) pruneBackupsLocked(retention int) error {
	if retention <= 0 {
		retention = 7
	}

	entries, err := os.ReadDir(s.backupDir)
	if err != nil {
		return err
	}

	var autoBackups []os.FileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.Contains(name, "_auto_") {
			continue
		}
		if info, err := entry.Info(); err == nil {
			autoBackups = append(autoBackups, info)
		}
	}

	if len(autoBackups) <= retention {
		return nil
	}

	// Sort newest first
	sort.Slice(autoBackups, func(i, j int) bool {
		return autoBackups[i].ModTime().After(autoBackups[j].ModTime())
	})

	// Remove older ones
	for i := retention; i < len(autoBackups); i++ {
		p := filepath.Join(s.backupDir, autoBackups[i].Name())
		slog.Info("PruneBackups: Removing obsolete automated backup", "file", autoBackups[i].Name())
		_ = os.Remove(p)
	}

	return nil
}

// DeleteBackup permanently removes a backup file
func (s *Service) DeleteBackup(filename string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.validateFilename(filename); err != nil {
		return err
	}

	p := filepath.Join(s.backupDir, filename)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return errors.New("backup file not found")
	}

	return os.Remove(p)
}

// DownloadBackup validates and returns the full path of a backup file
func (s *Service) DownloadBackup(filename string) (string, error) {
	if err := s.validateFilename(filename); err != nil {
		return "", err
	}

	p := filepath.Join(s.backupDir, filename)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return "", errors.New("backup file not found")
	}

	return p, nil
}

// SaveUploadedBackup writes an uploaded backup archive to the backup directory
func (s *Service) SaveUploadedBackup(r io.Reader, filename string) (*BackupInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.validateFilename(filename); err != nil {
		return nil, err
	}

	if err := s.EnsureBackupDir(); err != nil {
		return nil, err
	}

	dest := filepath.Join(s.backupDir, filename)
	out, err := os.Create(dest)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	if _, err := io.Copy(out, r); err != nil {
		_ = os.Remove(dest)
		return nil, err
	}

	fi, err := out.Stat()
	if err != nil {
		return nil, err
	}

	bType := "sqlite"
	if strings.Contains(filename, ".json") {
		bType = "json"
	}

	return &BackupInfo{
		Filename:      filename,
		Size:          fi.Size(),
		FormattedSize: FormatBytes(fi.Size()),
		CreatedAt:     fi.ModTime(),
		IsAuto:        false,
		Type:          bType,
	}, nil
}

// RestoreBackup safely restores the database from a given backup archive
func (s *Service) RestoreBackup(filename string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.validateFilename(filename); err != nil {
		return err
	}

	srcArchive := filepath.Join(s.backupDir, filename)
	if _, err := os.Stat(srcArchive); os.IsNotExist(err) {
		return errors.New("backup file not found")
	}

	// 1. Mandatory Pre-Restore Safety Snapshot to guarantee ZERO data loss
	safetyName := fmt.Sprintf("watcharr_backup_pre_restore_%s.db.gz", time.Now().Format("20060102_150405"))
	if s.db.Dialector.Name() == "sqlite" {
		if err := s.createSQLiteBackup(safetyName); err != nil {
			slog.Error("RestoreBackup: Failed to create safety backup prior to restore", "error", err)
			return fmt.Errorf("pre-restore safety backup failed: %w", err)
		}
	} else {
		safetyName = fmt.Sprintf("watcharr_backup_pre_restore_%s.json.gz", time.Now().Format("20060102_150405"))
		if err := s.createJSONBackup(safetyName); err != nil {
			slog.Error("RestoreBackup: Failed to create safety backup prior to restore", "error", err)
			return fmt.Errorf("pre-restore safety backup failed: %w", err)
		}
	}
	slog.Info("RestoreBackup: Created pre-restore safety snapshot", "safetyFile", safetyName)

	// 2. Perform restoration according to archive type
	if strings.HasSuffix(filename, ".db.gz") || strings.HasSuffix(filename, ".db") {
		if s.db.Dialector.Name() != "sqlite" {
			return errors.New("cannot restore a sqlite binary backup on a non-sqlite database")
		}
		return s.restoreSQLite(srcArchive)
	}

	if strings.HasSuffix(filename, ".json.gz") || strings.HasSuffix(filename, ".json") {
		return s.restoreJSON(srcArchive)
	}

	return errors.New("unsupported backup format for restore")
}

func (s *Service) restoreSQLite(archivePath string) error {
	tempRestore := filepath.Join(s.backupDir, "temp_restore.db")
	defer os.Remove(tempRestore)

	if strings.HasSuffix(archivePath, ".gz") {
		if err := ungzipFile(archivePath, tempRestore); err != nil {
			return fmt.Errorf("failed to decompress sqlite backup: %w", err)
		}
	} else {
		if err := copyFile(archivePath, tempRestore); err != nil {
			return fmt.Errorf("failed to copy sqlite backup: %w", err)
		}
	}

	// Checkpoint WAL before live copy so all frames are flushed and WAL is truncated cleanly
	if res := s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); res.Error != nil {
		slog.Warn("restoreSQLite: PRAGMA wal_checkpoint failed", "error", res.Error)
	}

	// Target sqlite path
	destDB := s.getActiveSQLiteDBPath()

	// Replace database file without unlinking active -wal/-shm descriptors
	if err := copyFile(tempRestore, destDB); err != nil {
		return fmt.Errorf("failed to write restored database file: %w", err)
	}

	// Checkpoint WAL after copy to ensure shared-memory index is synchronized
	if res := s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); res.Error != nil {
		slog.Warn("restoreSQLite: Post-restore PRAGMA wal_checkpoint failed", "error", res.Error)
	}
	var schemaVer int
	if err := s.db.Raw("PRAGMA user_version").Scan(&schemaVer).Error; err != nil {
		slog.Warn("restoreSQLite: Post-restore schema verification warning", "error", err)
	}

	slog.Info("restoreSQLite: Database restored successfully from SQLite archive", "file", archivePath)
	return nil
}

func (s *Service) restoreJSON(archivePath string) error {
	var reader io.Reader
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	if strings.HasSuffix(archivePath, ".gz") {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return err
		}
		defer gz.Close()
		reader = io.LimitReader(gz, maxDecompressedBackupBytes)
	} else {
		reader = io.LimitReader(f, maxDecompressedBackupBytes)
	}

	var dump UniversalBackupDump
	if err := json.NewDecoder(reader).Decode(&dump); err != nil {
		return fmt.Errorf("failed to decode json backup archive: %w", err)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// 1. Delete existing tables in reverse child-first topological order
		if err := tx.Exec("DELETE FROM watched_tags").Error; err != nil {
			return fmt.Errorf("failed to clear watched_tags: %w", err)
		}
		modelsToClear := []any{
			&entity.Activity{},
			&entity.WatchedEpisode{},
			&entity.WatchedSeason{},
			&entity.Watched{},
			&entity.Tag{},
			&entity.ArrRequest{},
			&entity.PushSubscription{},
			&entity.Follow{},
			&entity.Token{},
			&entity.UserServices{},
			&entity.Game{},
			&entity.Content{},
			&entity.User{},
			&entity.Image{},
		}
		for _, m := range modelsToClear {
			if err := tx.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(m).Error; err != nil {
				return fmt.Errorf("failed to clear table for %T: %w", m, err)
			}
		}

		// 2. Insert in topological parent-first order:
		// Images -> Users -> UserServices -> Follow -> Content -> Games -> Tags -> Watched -> watched_tags -> WatchedSeasons -> WatchedEpisodes -> Activities -> ArrRequests -> PushSubscriptions
		for _, ib := range dump.Images {
			img := ib.Image
			if ib.ID != 0 {
				img.ID = ib.ID
			}
			if ib.Hash != "" {
				img.Hash = ib.Hash
			}
			if err := tx.Create(&img).Error; err != nil {
				return fmt.Errorf("failed to restore image %d: %w", img.ID, err)
			}
		}

		for _, ub := range dump.Users {
			u := ub.User
			if ub.AvatarID != nil {
				u.AvatarID = ub.AvatarID
			}
			if ub.Permissions != 0 {
				u.Permissions = ub.Permissions
			}
			pwHash := ub.PasswordHash
			if pwHash == "" && dump.PasswordHashes != nil {
				pwHash = dump.PasswordHashes[u.ID]
			}
			if pwHash == "" {
				pwHash = u.Password
			}
			u.Password = pwHash
			if err := tx.Create(&u).Error; err != nil {
				return fmt.Errorf("failed to restore user %s: %w", u.Username, err)
			}
			if pwHash != "" {
				if err := tx.Table("users").Where("id = ?", u.ID).Update("password", pwHash).Error; err != nil {
					return fmt.Errorf("failed to restore password hash for user %s: %w", u.Username, err)
				}
			}
		}

		for _, usb := range dump.UserServices {
			us := usb.UserServices
			if usb.Name != "" {
				us.Name = usb.Name
			}
			if usb.ClientID != "" {
				us.ClientID = usb.ClientID
			}
			if usb.AuthToken != "" {
				us.AuthToken = usb.AuthToken
			}
			if usb.AuthToken2 != "" {
				us.AuthToken2 = usb.AuthToken2
			}
			if usb.UserID != 0 {
				us.UserID = usb.UserID
			}
			if err := tx.Create(&us).Error; err != nil {
				return fmt.Errorf("failed to restore user service: %w", err)
			}
		}

		for _, fb := range dump.Follow {
			fl := fb.Follow
			if !fb.UpdatedAt.IsZero() {
				fl.UpdatedAt = fb.UpdatedAt
			}
			if fb.UserID != 0 {
				fl.UserID = fb.UserID
			}
			if fb.FollowedUserID != 0 {
				fl.FollowedUserID = fb.FollowedUserID
			}
			if err := tx.Create(&fl).Error; err != nil {
				return fmt.Errorf("failed to restore follow: %w", err)
			}
		}

		for _, c := range dump.Content {
			if err := tx.Create(&c).Error; err != nil {
				return fmt.Errorf("failed to restore content %s: %w", c.Title, err)
			}
		}

		for _, gb := range dump.Games {
			g := gb.Game
			if gb.PosterID != nil {
				g.PosterID = gb.PosterID
			}
			if err := tx.Create(&g).Error; err != nil {
				return fmt.Errorf("failed to restore game: %w", err)
			}
		}

		for _, tb := range dump.Tags {
			t := tb.Tag
			if tb.UserID != 0 {
				t.UserID = tb.UserID
			}
			if err := tx.Create(&t).Error; err != nil {
				return fmt.Errorf("failed to restore tag: %w", err)
			}
		}

		for _, wb := range dump.Watched {
			w := wb.Watched
			if wb.UserID != 0 {
				w.UserID = wb.UserID
			}
			if wb.ContentID != nil {
				w.ContentID = wb.ContentID
			}
			if wb.GameID != nil {
				w.GameID = wb.GameID
			}
			if err := tx.Create(&w).Error; err != nil {
				return fmt.Errorf("failed to restore watched item: %w", err)
			}
		}

		for _, wt := range dump.WatchedTags {
			if err := tx.Table("watched_tags").Create(map[string]any{
				"watched_id": wt.WatchedID,
				"tag_id":     wt.TagID,
			}).Error; err != nil {
				return fmt.Errorf("failed to restore watched_tag: %w", err)
			}
		}

		for _, sb := range dump.Seasons {
			ws := sb.WatchedSeason
			if sb.UserID != 0 {
				ws.UserID = sb.UserID
			}
			if sb.WatchedID != 0 {
				ws.WatchedID = sb.WatchedID
			}
			if err := tx.Create(&ws).Error; err != nil {
				return fmt.Errorf("failed to restore watched season: %w", err)
			}
		}

		for _, epb := range dump.Episodes {
			ep := epb.WatchedEpisode
			if epb.UserID != 0 {
				ep.UserID = epb.UserID
			}
			if epb.WatchedID != 0 {
				ep.WatchedID = epb.WatchedID
			}
			if err := tx.Create(&ep).Error; err != nil {
				return fmt.Errorf("failed to restore watched episode: %w", err)
			}
		}

		for _, actb := range dump.Activity {
			act := actb.Activity
			if actb.UserID != 0 {
				act.UserID = actb.UserID
			}
			if err := tx.Create(&act).Error; err != nil {
				return fmt.Errorf("failed to restore activity: %w", err)
			}
		}

		for _, arb := range dump.ArrRequests {
			r := arb.ArrRequest
			if arb.UserID != 0 {
				r.UserID = arb.UserID
			}
			if arb.ContentID != nil {
				r.ContentID = arb.ContentID
			}
			if err := tx.Create(&r).Error; err != nil {
				return fmt.Errorf("failed to restore arr request: %w", err)
			}
		}

		for _, ps := range dump.PushSubscriptions {
			if err := tx.Create(&ps).Error; err != nil {
				return fmt.Errorf("failed to restore push subscription: %w", err)
			}
		}

		return nil
	})
}

// GetConfig returns aggregated backup status and settings
func (s *Service) GetConfig() BackupConfigResponse {
	backups, _ := s.ListBackups()

	var totalSize int64
	var lastBackup *time.Time
	if len(backups) > 0 {
		t := backups[0].CreatedAt
		lastBackup = &t
		for _, b := range backups {
			totalSize += b.Size
		}
	}

	sched := s.cfg.GetTaskSchedule("Automated Database Backup")
	if sched <= 0 {
		sched = 86400 // Default 24 hours
	}

	return BackupConfigResponse{
		Enabled:            s.cfg.GetBackupEnabled(),
		Retention:          s.cfg.GetBackupRetention(),
		ScheduleSeconds:    sched,
		TotalBackups:       len(backups),
		TotalSize:          totalSize,
		FormattedTotalSize: FormatBytes(totalSize),
		Driver:             s.db.Dialector.Name(),
		LastBackup:         lastBackup,
	}
}

// UpdateConfig updates backup settings and retention policy
func (s *Service) UpdateConfig(req UpdateBackupConfigRequest) error {
	if req.Enabled != nil {
		s.cfg.SetBackupEnabled(*req.Enabled)
	}
	if req.Retention != nil {
		s.cfg.SetBackupRetention(*req.Retention)
	}
	if req.ScheduleSeconds != nil && *req.ScheduleSeconds > 0 {
		s.cfg.SetTaskSchedule("Automated Database Backup", *req.ScheduleSeconds)
		if s.rescheduleTask != nil {
			if err := s.rescheduleTask(*req.ScheduleSeconds); err != nil {
				slog.Warn("UpdateConfig: failed to reschedule Automated Database Backup task", "error", err)
			}
		}
	}
	return s.cfg.Write()
}

// Utility functions for file copying and compression
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func gzipFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	gw := gzip.NewWriter(out)
	defer gw.Close()

	_, err = io.Copy(gw, in)
	return err
}

func ungzipFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	gr, err := gzip.NewReader(in)
	if err != nil {
		return err
	}
	defer gr.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, io.LimitReader(gr, maxDecompressedBackupBytes))
	return err
}
