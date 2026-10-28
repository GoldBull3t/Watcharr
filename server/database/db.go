package database

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/database/migrate"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DBType represents the supported database types
type DBType string

const (
	DBTypeSQLite   DBType = "sqlite"
	DBTypePostgres DBType = "postgres"
	DBTypeMySQL    DBType = "mysql"
)

// GetConfiguredDBType returns the active database engine type determined from environment variables
func GetConfiguredDBType() DBType {
	t := strings.ToLower(strings.TrimSpace(os.Getenv("DB_TYPE")))
	if t == "postgresql" || t == "postgres" || t == "pg" {
		return DBTypePostgres
	}
	if t == "mysql" || t == "mariadb" {
		return DBTypeMySQL
	}
	if t == "sqlite" || t == "sqlite3" {
		return DBTypeSQLite
	}

	// Auto-detect from connection string if DB_TYPE is not set
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		return DBTypePostgres
	}
	if strings.HasPrefix(dsn, "mysql://") {
		return DBTypeMySQL
	}

	return DBTypeSQLite
}

// GetSQLiteDBPath returns the filesystem path to the primary SQLite database file
func GetSQLiteDBPath() string {
	customDSN := os.Getenv("DB_DSN")
	if customDSN == "" {
		customDSN = os.Getenv("DATABASE_URL")
	}
	if customDSN != "" && !strings.HasPrefix(customDSN, "postgres") && !strings.HasPrefix(customDSN, "mysql") {
		parts := strings.Split(customDSN, "?")
		return parts[0]
	}
	return path.Join(config.DataPath, "watcharr.db")
}

// Create a new database connection.
// Supports SQLite (default), PostgreSQL, and MySQL based on DB_TYPE or connection string.
// For initial database connection in `main()`, also call `Setup()` on the
// returned `db` to run migrations, etc!
//
// NOTE: Our mock db used in tests mimics this function, so if this func is
// changed, you should look at the mock db to make it match if it makes
// sense, so our tests stay accurate to prod.
func Open() (*gorm.DB, error) {
	dbType := GetConfiguredDBType()
	slog.Info("Open: Opening database connection", "type", dbType)

	var dialector gorm.Dialector
	customDSN := os.Getenv("DB_DSN")
	if customDSN == "" {
		customDSN = os.Getenv("DATABASE_URL")
	}

	switch dbType {
	case DBTypePostgres:
		if customDSN == "" {
			return nil, fmt.Errorf("DB_DSN or DATABASE_URL must be specified when using DB_TYPE=postgres")
		}
		dialector = postgres.Open(customDSN)

	case DBTypeMySQL:
		if customDSN == "" {
			return nil, fmt.Errorf("DB_DSN or DATABASE_URL must be specified when using DB_TYPE=mysql")
		}
		dialector = mysql.Open(customDSN)

	case DBTypeSQLite:
		fallthrough
	default:
		dsn := customDSN
		if dsn == "" {
			dsn = fmt.Sprintf(
				"%s?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=cache_size(-2000)&_pragma=temp_store(FILE)&_pragma=mmap_size(0)",
				path.Join(config.DataPath, "watcharr.db"),
			)
		}
		dialector = sqlite.Open(dsn)
	}

	db, err := gorm.Open(
		dialector,
		&gorm.Config{
			TranslateError: true,
			PrepareStmt:    true,
		},
	)
	if err != nil {
		slog.Error("Open: Opening database failed.", "type", dbType, "error", err)
		return nil, err
	}
	return db, nil
}

// Setup configures our db connection and applies migrations.
//
// NOTE: This exists as a separate function so it can be reused by our testutil
// package that we want to have configured in the same way as the main db so
// that tests reflect real life.
func Setup(db *gorm.DB) error {
	if err := configure(db); err != nil {
		slog.Error("Setup: Configuring connection failed!", "error", err)
		return err
	}
	// Perform auto migration.
	slog.Info("Setup: AutoMigrating")
	err := db.AutoMigrate(
		&migrate.MigrationRecord{},
		&entity.User{},
		&entity.UserServices{},
		&entity.Content{},
		&entity.Watched{},
		&entity.WatchedSeason{},
		&entity.WatchedEpisode{},
		&entity.Activity{},
		&entity.Token{},
		&entity.Follow{},
		&entity.Image{},
		&entity.Game{},
		&entity.ArrRequest{},
		&entity.Tag{},
		&entity.PushSubscription{},
	)
	if err != nil {
		slog.Error("Setup: Auto migration failed.")
		return err
	}
	slog.Info("Setup: AutoMigrated")
	// Perform our manual migrations.
	if err := migrate.Now(db); err != nil {
		slog.Error("Setup: Manual migrations failed.", "error", err)
		return err
	}
	// Optimize database.
	if err := optimize(db); err != nil {
		slog.Error("Setup: Optimizing database failed.", "error", err)
		return err
	}
	return nil
}

// Configure our database connection and pool.
// For SQLite, PRAGMAs need to be defined per-connection, so we do that here.
// For PostgreSQL and MySQL, connection pooling parameters are tuned.
func configure(db *gorm.DB) error {
	driverName := db.Dialector.Name()
	slog.Info("configure: Configuring connection.", "driver", driverName)

	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("configure: Failed to get underlying sql.DB!", "error", err)
		return err
	}

	if driverName == "sqlite" {
		// Busy timeout: allow concurrent writers to wait up to 5000ms instead of immediately failing with SQLITE_BUSY
		if res := db.Exec("PRAGMA busy_timeout = 5000"); res.Error != nil {
			slog.Error("configure: Configuring busy_timeout failed!", "error", res.Error)
			return res.Error
		}
		slog.Info("configure: Configured busy_timeout to 5000ms.")

		// Foreign Keys: SQLite disables foreign keys by default per connection. Enable them to enforce integrity and cascading.
		if res := db.Exec("PRAGMA foreign_keys = ON"); res.Error != nil {
			slog.Error("configure: Configuring foreign_keys failed!", "error", res.Error)
			return res.Error
		}
		slog.Info("configure: Configured foreign_keys to ON.")

		// Journal Mode: Ensure SQLite WAL mode is active for concurrency between readers and writer
		if res := db.Exec("PRAGMA journal_mode = WAL"); res.Error != nil {
			slog.Warn("configure: Configuring journal_mode to WAL failed!", "error", res.Error)
		} else {
			slog.Info("configure: Configured journal_mode to WAL.")
		}

		// Synchronous: https://sqlite.org/pragma.html#pragma_synchronous
		// In WAL mode, `NORMAL` (1) is the official recommended setting.
		if res := db.Exec("PRAGMA synchronous=1"); res.Error != nil {
			slog.Error("configure: Configuring synchronous failed!")
			return res.Error
		}
		slog.Info("configure: Configured synchronous to NORMAL (1).")

		// Page cache size: allocate up to 2MB (-2000 KiB) in RAM for query caching
		if res := db.Exec("PRAGMA cache_size = -2000"); res.Error != nil {
			slog.Warn("configure: Configuring cache_size failed!", "error", res.Error)
		} else {
			slog.Info("configure: Configured cache_size to 2MB.")
		}

		// Temp store: store temp tables, sorting indices and views on disk to conserve container RAM
		if res := db.Exec("PRAGMA temp_store = FILE"); res.Error != nil {
			slog.Warn("configure: Configuring temp_store failed!", "error", res.Error)
		} else {
			slog.Info("configure: Configured temp_store to FILE.")
		}

		// Disable mmap to avoid holding mapped pages in container RSS
		if res := db.Exec("PRAGMA mmap_size = 0"); res.Error != nil {
			slog.Warn("configure: Configuring mmap_size failed!", "error", res.Error)
		} else {
			slog.Info("configure: Configured mmap_size to 0.")
		}

		// Pool configuration: SQLite works best with lean pool sizing to avoid idle buffer overhead
		sqlDB.SetMaxOpenConns(2)
		sqlDB.SetMaxIdleConns(1)
		sqlDB.SetConnMaxLifetime(10 * time.Minute)
	} else {
		// External database engine (PostgreSQL, MySQL / MariaDB)
		maxOpen := 25
		maxIdle := 5
		if envOpen := os.Getenv("DB_MAX_OPEN_CONNS"); envOpen != "" {
			if v, err := strconv.Atoi(envOpen); err == nil && v > 0 {
				maxOpen = v
			}
		}
		if envIdle := os.Getenv("DB_MAX_IDLE_CONNS"); envIdle != "" {
			if v, err := strconv.Atoi(envIdle); err == nil && v > 0 {
				maxIdle = v
			}
		}
		sqlDB.SetMaxOpenConns(maxOpen)
		sqlDB.SetMaxIdleConns(maxIdle)
		sqlDB.SetConnMaxLifetime(1 * time.Hour)
		slog.Info("configure: Configured connection pool for external database", "driver", driverName, "max_open", maxOpen, "max_idle", maxIdle)
	}

	return nil
}

// Optimize the database.
func optimize(db *gorm.DB) error {
	if db.Dialector.Name() != "sqlite" {
		slog.Debug("optimize: Skipping SQLite-specific PRAGMAs for non-SQLite database", "driver", db.Dialector.Name())
		return nil
	}

	slog.Info("optimize: Running optimizations.")

	// WAL Checkpoint (aka commit anything in the WAL to the main db).
	// https://sqlite.org/pragma.html#pragma_wal_checkpoint
	timeBeforeQuery := time.Now()
	if res := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); res.Error != nil {
		slog.Error("optimize: Checkpoint failed!")
		return res.Error
	}
	slog.Info("optimize: Checkpointed.", "took", time.Since(timeBeforeQuery))

	// Optimize pragma.
	// https://sqlite.org/pragma.html#pragma_optimize
	timeBeforeQuery = time.Now()
	if res := db.Exec("PRAGMA optimize=0x10002"); res.Error != nil {
		slog.Error("optimize: Optimize pragma failed!")
		return res.Error
	}
	slog.Info("optimize: Optimize pragma succeeded.",
		"took", time.Since(timeBeforeQuery))

	// Shrink memory: return unused page cache to the operating system
	if res := db.Exec("PRAGMA shrink_memory"); res.Error != nil {
		slog.Warn("optimize: shrink_memory failed!", "error", res.Error)
	} else {
		slog.Info("optimize: Shrunk SQLite in-memory cache.")
	}

	slog.Info("optimize: Done.")
	return nil
}

// Optimize task that is scheduled and ran every whenever.
// So this func is for optimzations that we want to re-run every time the
// task is scheduled for.
func TaskOptimize(db *gorm.DB) error {
	if db.Dialector.Name() != "sqlite" {
		slog.Debug("TaskOptimize: Skipping SQLite-specific PRAGMAs for non-SQLite database", "driver", db.Dialector.Name())
		return nil
	}

	slog.Info("TaskOptimize: Running optimizations.")

	// WAL Checkpoint (aka commit anything in the WAL to the main db).
	// https://sqlite.org/pragma.html#pragma_wal_checkpoint
	timeBeforeQuery := time.Now()
	if res := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); res.Error != nil {
		slog.Error("TaskOptimize: Checkpoint failed!")
		return res.Error
	}
	slog.Info("TaskOptimize: Checkpointed.",
		"took", time.Since(timeBeforeQuery))

	// Optimize pragma.
	// https://sqlite.org/pragma.html#pragma_optimize
	timeBeforeQuery = time.Now()
	if res := db.Exec("PRAGMA optimize"); res.Error != nil {
		slog.Error("TaskOptimize: Optimize pragma failed!")
		return res.Error
	}
	slog.Info("TaskOptimize: Optimize pragma succeeded.",
		"took", time.Since(timeBeforeQuery))

	// Shrink memory: return unused page cache to the operating system
	if res := db.Exec("PRAGMA shrink_memory"); res.Error != nil {
		slog.Warn("TaskOptimize: shrink_memory failed!", "error", res.Error)
	} else {
		slog.Info("TaskOptimize: Shrunk SQLite in-memory cache.")
	}

	slog.Info("TaskOptimize: Done.")
	return nil
}
