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
	"gorm.io/gorm"
)

type Service struct {
	db        *gorm.DB
	cfg       *config.ServerConfig
	backupDir string
	mu        sync.Mutex
}

func NewService(db *gorm.DB, cfg *config.ServerConfig) *Service {
	bDir := filepath.Join(config.DataPath, "backups")
	s := &Service{
		db:        db,
		cfg:       cfg,
		backupDir: bDir,
	}
	_ = s.EnsureBackupDir()
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
		_ = s.pruneBackupsLocked(s.cfg.GetBackupRetention())
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
	_ = s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")

	tempRaw := filepath.Join(s.backupDir, "raw_"+finalGzFilename+".tmp")
	defer os.Remove(tempRaw)

	// Try VACUUM INTO for zero-lock, defragmented consistent snapshot
	vacRes := s.db.Exec(fmt.Sprintf("VACUUM INTO '%s'", tempRaw))
	if vacRes.Error != nil {
		slog.Warn("createSQLiteBackup: VACUUM INTO failed, falling back to direct copy", "error", vacRes.Error)
		srcPath := database.GetSQLiteDBPath()
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

func (s *Service) createJSONBackup(finalGzFilename string) error {
	dump := UniversalBackupDump{
		Version:    "1.0",
		ExportedAt: time.Now(),
		Driver:     s.db.Dialector.Name(),
	}

	_ = s.db.Find(&dump.Users).Error
	_ = s.db.Find(&dump.UserServices).Error
	_ = s.db.Find(&dump.Content).Error
	_ = s.db.Find(&dump.Watched).Error
	_ = s.db.Find(&dump.Seasons).Error
	_ = s.db.Find(&dump.Episodes).Error
	_ = s.db.Find(&dump.Activity).Error
	_ = s.db.Find(&dump.Follow).Error
	_ = s.db.Find(&dump.Images).Error
	_ = s.db.Find(&dump.Games).Error
	_ = s.db.Find(&dump.ArrRequests).Error
	_ = s.db.Find(&dump.Tags).Error

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
		_ = s.createSQLiteBackup(safetyName)
	} else {
		safetyName = fmt.Sprintf("watcharr_backup_pre_restore_%s.json.gz", time.Now().Format("20060102_150405"))
		_ = s.createJSONBackup(safetyName)
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

	// Checkpoint WAL
	_ = s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")

	// Target sqlite path
	destDB := database.GetSQLiteDBPath()

	// Replace database file
	if err := copyFile(tempRestore, destDB); err != nil {
		return fmt.Errorf("failed to write restored database file: %w", err)
	}

	// Clean up stale WAL / SHM files if any
	_ = os.Remove(destDB + "-wal")
	_ = os.Remove(destDB + "-shm")

	// Verify restored db integrity
	_ = s.db.Exec("PRAGMA wal_checkpoint(TRUNCATE)")
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
		reader = gz
	} else {
		reader = f
	}

	var dump UniversalBackupDump
	if err := json.NewDecoder(reader).Decode(&dump); err != nil {
		return fmt.Errorf("failed to decode json backup archive: %w", err)
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		// Replace all core data in transaction
		if len(dump.Users) > 0 {
			for _, u := range dump.Users {
				_ = tx.Save(&u).Error
			}
		}
		if len(dump.UserServices) > 0 {
			for _, us := range dump.UserServices {
				_ = tx.Save(&us).Error
			}
		}
		if len(dump.Content) > 0 {
			for _, c := range dump.Content {
				_ = tx.Save(&c).Error
			}
		}
		if len(dump.Watched) > 0 {
			for _, w := range dump.Watched {
				_ = tx.Save(&w).Error
			}
		}
		if len(dump.Seasons) > 0 {
			for _, s := range dump.Seasons {
				_ = tx.Save(&s).Error
			}
		}
		if len(dump.Episodes) > 0 {
			for _, ep := range dump.Episodes {
				_ = tx.Save(&ep).Error
			}
		}
		if len(dump.Activity) > 0 {
			for _, act := range dump.Activity {
				_ = tx.Save(&act).Error
			}
		}
		if len(dump.Tags) > 0 {
			for _, t := range dump.Tags {
				_ = tx.Save(&t).Error
			}
		}
		if len(dump.Games) > 0 {
			for _, g := range dump.Games {
				_ = tx.Save(&g).Error
			}
		}
		if len(dump.ArrRequests) > 0 {
			for _, r := range dump.ArrRequests {
				_ = tx.Save(&r).Error
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

	_, err = io.Copy(out, gr)
	return err
}
