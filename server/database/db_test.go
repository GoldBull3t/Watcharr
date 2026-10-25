package database

import (
	"os"
	"testing"
)

func TestGetConfiguredDBType(t *testing.T) {
	// Test default is SQLite
	os.Unsetenv("DB_TYPE")
	os.Unsetenv("DB_DSN")
	os.Unsetenv("DATABASE_URL")
	if GetConfiguredDBType() != DBTypeSQLite {
		t.Fatalf("expected default to be DBTypeSQLite, got %s", GetConfiguredDBType())
	}

	// Test explicit DB_TYPE postgres
	os.Setenv("DB_TYPE", "postgres")
	if GetConfiguredDBType() != DBTypePostgres {
		t.Fatalf("expected DBTypePostgres, got %s", GetConfiguredDBType())
	}

	os.Setenv("DB_TYPE", "postgresql")
	if GetConfiguredDBType() != DBTypePostgres {
		t.Fatalf("expected DBTypePostgres, got %s", GetConfiguredDBType())
	}

	// Test explicit DB_TYPE mysql
	os.Setenv("DB_TYPE", "mysql")
	if GetConfiguredDBType() != DBTypeMySQL {
		t.Fatalf("expected DBTypeMySQL, got %s", GetConfiguredDBType())
	}

	os.Setenv("DB_TYPE", "mariadb")
	if GetConfiguredDBType() != DBTypeMySQL {
		t.Fatalf("expected DBTypeMySQL, got %s", GetConfiguredDBType())
	}

	// Test auto-detect from DATABASE_URL
	os.Unsetenv("DB_TYPE")
	os.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/watcharr")
	if GetConfiguredDBType() != DBTypePostgres {
		t.Fatalf("expected auto-detected DBTypePostgres, got %s", GetConfiguredDBType())
	}

	os.Setenv("DATABASE_URL", "mysql://user:pass@localhost:3306/watcharr")
	if GetConfiguredDBType() != DBTypeMySQL {
		t.Fatalf("expected auto-detected DBTypeMySQL, got %s", GetConfiguredDBType())
	}

	// Reset
	os.Unsetenv("DB_TYPE")
	os.Unsetenv("DATABASE_URL")
}

func TestOpenMissingDSN(t *testing.T) {
	os.Setenv("DB_TYPE", "postgres")
	os.Unsetenv("DB_DSN")
	os.Unsetenv("DATABASE_URL")
	_, err := Open()
	if err == nil {
		t.Fatalf("expected error when DB_TYPE=postgres without DSN, got nil")
	}

	os.Setenv("DB_TYPE", "mysql")
	_, err = Open()
	if err == nil {
		t.Fatalf("expected error when DB_TYPE=mysql without DSN, got nil")
	}

	os.Unsetenv("DB_TYPE")
}

