package backup

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/internal/testutil"
)

func TestCreateAndListBackup(t *testing.T) {
	testutil.SetupLogging()
	db := testutil.SetupDB(t)

	// Insert dummy user
	db.Create(&entity.User{
		Username: "backup_test_user",
		Type:     entity.WATCHARR_USER,
	})

	cfg := &config.ServerConfig{}
	cfg.SetBackupEnabled(true)
	cfg.SetBackupRetention(3)

	s := NewService(db, cfg)
	s.backupDir = filepath.Join(t.TempDir(), "backups")
	_ = s.EnsureBackupDir()

	// 1. Create a manual backup (JSON format in test environment)
	info, err := s.CreateBackup(false, "json")
	if err != nil {
		t.Fatalf("CreateBackup(false, json) failed: %v", err)
	}

	if info == nil || info.Filename == "" {
		t.Fatalf("expected non-empty BackupInfo, got nil or empty")
	}

	if info.IsAuto {
		t.Errorf("expected IsAuto to be false for manual backup")
	}

	// 2. List backups
	list, err := s.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups failed: %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("expected 1 backup, got %d", len(list))
	}

	if list[0].Filename != info.Filename {
		t.Errorf("expected filename %s, got %s", info.Filename, list[0].Filename)
	}
}

func TestPruneBackups(t *testing.T) {
	testutil.SetupLogging()
	db := testutil.SetupDB(t)

	cfg := &config.ServerConfig{}
	cfg.SetBackupEnabled(true)
	cfg.SetBackupRetention(2)

	s := NewService(db, cfg)
	s.backupDir = filepath.Join(t.TempDir(), "backups")
	_ = s.EnsureBackupDir()

	// Create 4 auto backups
	for i := 0; i < 4; i++ {
		// Mock small time difference for file timestamps
		time.Sleep(10 * time.Millisecond)
		_, err := s.CreateBackup(true, "json")
		if err != nil {
			t.Fatalf("CreateBackup(true) failed: %v", err)
		}
	}

	// Also create 1 manual backup (which should never be pruned by retention)
	_, err := s.CreateBackup(false, "json")
	if err != nil {
		t.Fatalf("CreateBackup(false) failed: %v", err)
	}

	list, err := s.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups failed: %v", err)
	}

	var autoCount, manualCount int
	for _, b := range list {
		if b.IsAuto {
			autoCount++
		} else {
			manualCount++
		}
	}

	// Retention is 2, so at most 2 auto backups should remain
	if autoCount > 2 {
		t.Errorf("expected at most 2 auto backups after pruning, got %d", autoCount)
	}

	if manualCount != 1 {
		t.Errorf("expected 1 manual backup preserved, got %d", manualCount)
	}
}

func TestPathTraversalProtection(t *testing.T) {
	testutil.SetupLogging()
	db := testutil.SetupDB(t)

	cfg := &config.ServerConfig{}
	s := NewService(db, cfg)
	s.backupDir = filepath.Join(t.TempDir(), "backups")
	_ = s.EnsureBackupDir()

	maliciousFilenames := []string{
		"../secret.txt",
		"..\\windows\\system32",
		"/etc/passwd",
		"sub/folder/file.db.gz",
		"file.exe",
		"malicious.sh",
	}

	for _, name := range maliciousFilenames {
		if err := s.DeleteBackup(name); err == nil {
			t.Errorf("expected DeleteBackup(%q) to fail with path traversal error, got nil", name)
		}
		if _, err := s.DownloadBackup(name); err == nil {
			t.Errorf("expected DownloadBackup(%q) to fail with path traversal error, got nil", name)
		}
	}
}

func TestRestoreSafetySnapshot(t *testing.T) {
	testutil.SetupLogging()
	db := testutil.SetupDB(t)

	cfg := &config.ServerConfig{}
	s := NewService(db, cfg)
	s.backupDir = filepath.Join(t.TempDir(), "backups")
	_ = s.EnsureBackupDir()

	// Create a backup
	info, err := s.CreateBackup(false, "json")
	if err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Restore from it
	if err := s.RestoreBackup(info.Filename); err != nil {
		t.Fatalf("RestoreBackup failed: %v", err)
	}

	// Verify that a pre-restore safety snapshot was generated
	list, err := s.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups failed: %v", err)
	}

	hasPreRestore := false
	for _, b := range list {
		if b.Filename != info.Filename && filepath.Ext(b.Filename) == ".gz" {
			hasPreRestore = true
			break
		}
	}

	if !hasPreRestore {
		t.Errorf("expected a pre-restore safety snapshot to be created in backups folder")
	}
}

func TestJSONBackupRestorePreservesAllEntities(t *testing.T) {
	testutil.SetupLogging()
	db := testutil.SetupDB(t)

	// 1. Seed images, users (with passwords), follows, games, tags, watched, watched_tags, push_subscriptions
	img1 := entity.Image{Hash: "hash_avatar_1", Path: "img/avatar/1.jpg", BlurHash: "bh1"}
	img2 := entity.Image{Hash: "hash_poster_2", Path: "img/poster/2.jpg", BlurHash: "bh2"}
	if err := db.Create(&img1).Error; err != nil {
		t.Fatalf("failed to create img1: %v", err)
	}
	if err := db.Create(&img2).Error; err != nil {
		t.Fatalf("failed to create img2: %v", err)
	}

	u1 := entity.User{
		Username: "alice",
		Password: "$2a$10$hashedpasswordalice1234567890",
		Type:     entity.WATCHARR_USER,
		AvatarID: &img1.ID,
	}
	u2 := entity.User{
		Username: "bob",
		Password: "$2a$10$hashedpasswordbob1234567890",
		Type:     entity.WATCHARR_USER,
	}
	if err := db.Create(&u1).Error; err != nil {
		t.Fatalf("failed to create u1: %v", err)
	}
	if err := db.Create(&u2).Error; err != nil {
		t.Fatalf("failed to create u2: %v", err)
	}

	follow := entity.Follow{
		UserID:         u1.ID,
		FollowedUserID: u2.ID,
	}
	if err := db.Create(&follow).Error; err != nil {
		t.Fatalf("failed to create follow: %v", err)
	}

	game := entity.Game{
		IgdbID:   4242,
		Name:     "Elden Ring",
		PosterID: &img2.ID,
	}
	if err := db.Create(&game).Error; err != nil {
		t.Fatalf("failed to create game: %v", err)
	}

	tag := entity.Tag{
		UserID: u1.ID,
		Name:   "Favorites",
	}
	if err := db.Create(&tag).Error; err != nil {
		t.Fatalf("failed to create tag: %v", err)
	}

	watchedItem := entity.Watched{
		UserID: u1.ID,
		GameID: &game.ID,
		Status: entity.FINISHED,
		Rating: 10.0,
		Tags:   []entity.Tag{tag},
	}
	if err := db.Create(&watchedItem).Error; err != nil {
		t.Fatalf("failed to create watchedItem: %v", err)
	}

	ps := entity.PushSubscription{
		UserID:   u1.ID,
		Endpoint: "https://fcm.googleapis.com/fcm/send/test-sub-1",
		P256dh:   "p256dh-key",
		Auth:     "auth-key",
	}
	if err := db.Create(&ps).Error; err != nil {
		t.Fatalf("failed to create push subscription: %v", err)
	}

	cfg := &config.ServerConfig{}
	s := NewService(db, cfg)
	s.backupDir = filepath.Join(t.TempDir(), "backups")
	_ = s.EnsureBackupDir()

	// 2. Create JSON backup
	info, err := s.CreateBackup(false, "json")
	if err != nil {
		t.Fatalf("CreateBackup(json) failed: %v", err)
	}

	// 3. Mutate/wipe database state before restoring
	db.Exec("DELETE FROM watched_tags")
	db.Unscoped().Where("1 = 1").Delete(&entity.Watched{})
	db.Unscoped().Where("1 = 1").Delete(&entity.Tag{})
	db.Unscoped().Where("1 = 1").Delete(&entity.Game{})
	db.Unscoped().Where("1 = 1").Delete(&entity.Follow{})
	db.Unscoped().Where("1 = 1").Delete(&entity.PushSubscription{})
	db.Table("users").Where("id = ?", u1.ID).Update("password", "corrupted")

	// 4. Restore from JSON backup
	if err := s.RestoreBackup(info.Filename); err != nil {
		t.Fatalf("RestoreBackup(json) failed: %v", err)
	}

	// 5. Verify password hash was preserved
	var restoredPass string
	if err := db.Table("users").Select("password").Where("id = ?", u1.ID).Scan(&restoredPass).Error; err != nil {
		t.Fatalf("failed to query restored password: %v", err)
	}
	if restoredPass != "$2a$10$hashedpasswordalice1234567890" {
		t.Errorf("expected password hash to be preserved, got %q", restoredPass)
	}

	// 6. Verify images, games, follows, watched_tags, and push_subscriptions
	var imgCount int64
	db.Model(&entity.Image{}).Count(&imgCount)
	if imgCount != 2 {
		t.Errorf("expected 2 images restored, got %d", imgCount)
	}

	var restoredGame entity.Game
	if err := db.First(&restoredGame, game.ID).Error; err != nil {
		t.Fatalf("expected game restored: %v", err)
	}
	if restoredGame.PosterID == nil || *restoredGame.PosterID != img2.ID {
		t.Errorf("expected game PosterID=%d preserved, got %v", img2.ID, restoredGame.PosterID)
	}

	var followCount int64
	db.Model(&entity.Follow{}).Where("user_id = ? AND followed_user_id = ?", u1.ID, u2.ID).Count(&followCount)
	if followCount != 1 {
		t.Errorf("expected follow relationship preserved, got count=%d", followCount)
	}

	var wtCount int64
	db.Table("watched_tags").Where("watched_id = ? AND tag_id = ?", watchedItem.ID, tag.ID).Count(&wtCount)
	if wtCount != 1 {
		t.Errorf("expected watched_tags join entry preserved, got count=%d", wtCount)
	}

	var psCount int64
	db.Model(&entity.PushSubscription{}).Where("endpoint = ?", ps.Endpoint).Count(&psCount)
	if psCount != 1 {
		t.Errorf("expected push subscription preserved, got count=%d", psCount)
	}
}
