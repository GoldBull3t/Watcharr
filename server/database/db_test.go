package database

import (
	"os"
	"strings"
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

func TestNormalizeMySQLDSN(t *testing.T) {
	got := normalizeMySQLDSN("mysql://user:pass@localhost:3306/watcharr")
	want := "user:pass@tcp(localhost:3306)/watcharr?parseTime=True"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}

	got2 := normalizeMySQLDSN("user:pass@tcp(127.0.0.1:3306)/watcharr?charset=utf8mb4")
	want2 := "user:pass@tcp(127.0.0.1:3306)/watcharr?charset=utf8mb4&parseTime=True"
	if got2 != want2 {
		t.Fatalf("expected %q, got %q", want2, got2)
	}

	got3 := normalizeMySQLDSN("user:pass@tcp(127.0.0.1:3306)/watcharr?parseTime=true")
	want3 := "user:pass@tcp(127.0.0.1:3306)/watcharr?parseTime=true"
	if got3 != want3 {
		t.Fatalf("expected %q, got %q", want3, got3)
	}
}

func TestNormalizeSQLiteDSN(t *testing.T) {
	got := normalizeSQLiteDSN("/custom/path/watcharr.db")
	if !strings.Contains(got, "_pragma=foreign_keys(1)") || !strings.Contains(got, "_pragma=busy_timeout(10000)") {
		t.Fatalf("expected default pragmas appended to custom sqlite path, got %q", got)
	}

	gotWithQuery := normalizeSQLiteDSN("/custom/path/watcharr.db?cache=shared")
	if !strings.HasPrefix(gotWithQuery, "/custom/path/watcharr.db?cache=shared&_pragma=") {
		t.Fatalf("expected pragmas appended with &, got %q", gotWithQuery)
	}

	existing := "/custom/path/watcharr.db?_pragma=busy_timeout(5000)"
	if normalizeSQLiteDSN(existing) != existing {
		t.Fatalf("expected existing pragma DSN untouched, got %q", normalizeSQLiteDSN(existing))
	}
}
