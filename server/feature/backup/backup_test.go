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
