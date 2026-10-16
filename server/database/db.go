package database

import (
	"log/slog"
	"path"
	"time"

	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/database/migrate"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Create a new database connection.
// For initial database connection in `main()`, also call `Setup()` on the
// returned `db` to run migrations, etc!
//
// NOTE: Our mock db used in tests mimics this function, so if this func is
// changed, you should look at the mock db to make it match if it makes
// sense, so our tests stay accurate to prod.
func Open() (*gorm.DB, error) {
	slog.Info("Open: Opening new database connection")
	db, err := gorm.Open(
		sqlite.Open(path.Join(config.DataPath, "watcharr.db")),
		&gorm.Config{TranslateError: true},
	)
	if err != nil {
		slog.Error("Open: Opening database failed.")
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

// Configure our SQLite database connection.
// Some PRAGMAs need to be defined per-connection, so we do that here.
func configure(db *gorm.DB) error {
	slog.Info("configure: Configuring connection.")

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
	// It guarantees full database integrity against OS and app crashes, while
	// avoiding redundant and blocking fsync disk operations on every single transaction.
	if res := db.Exec("PRAGMA synchronous=1"); res.Error != nil {
		slog.Error("configure: Configuring synchronous failed!")
		return res.Error
	}
	slog.Info("configure: Configured synchronous to NORMAL (1).")

	// Page cache size: allocate up to 64MB (-64000 KiB) in RAM for query caching
	if res := db.Exec("PRAGMA cache_size = -64000"); res.Error != nil {
		slog.Warn("configure: Configuring cache_size failed!", "error", res.Error)
	} else {
		slog.Info("configure: Configured cache_size to 64MB.")
	}

	// Temp store: store temp tables, sorting indices and views in RAM rather than disk
	if res := db.Exec("PRAGMA temp_store = MEMORY"); res.Error != nil {
		slog.Warn("configure: Configuring temp_store failed!", "error", res.Error)
	} else {
		slog.Info("configure: Configured temp_store to MEMORY.")
	}

	// Pool configuration: SQLite works best with controlled pool sizing to avoid lock contention
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("configure: Failed to get underlying sql.DB!", "error", err)
		return err
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return nil
}

// Optimize the database.
func optimize(db *gorm.DB) error {
	slog.Info("optimize: Running optimizations.")

	// WAL Checkpoint (aka commit anything in the WAL to the main db).
	// Seems best to do this to make sure it has happened, especially since
	// we are running vacuum next.
	// https://sqlite.org/pragma.html#pragma_wal_checkpoint
	timeBeforeQuery := time.Now()
	if res := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); res.Error != nil {
		slog.Error("optimize: Checkpoint failed!")
		return res.Error
	}
	slog.Info("optimize: Checkpointed.", "took", time.Since(timeBeforeQuery))

	// Optimize pragma.
	// Running with recommended argument for our new long-living connection.
	// https://sqlite.org/pragma.html#pragma_optimize
	timeBeforeQuery = time.Now()
	if res := db.Exec("PRAGMA optimize=0x10002"); res.Error != nil {
		slog.Error("optimize: Optimize pragma failed!")
		return res.Error
	}
	slog.Info("optimize: Optimize pragma succeeded.",
		"took", time.Since(timeBeforeQuery))

	// Vacuum.
	// > VACUUM rebuilds the database file, repacking it into a minimal amount
	// > of disk space.
	// https://sqlite.org/lang_vacuum.html
	timeBeforeQuery = time.Now()
	if res := db.Exec("VACUUM"); res.Error != nil {
		slog.Error("optimize: Vacuum failed!")
		return res.Error
	}
	slog.Info("optimize: Vacuumed successfully.",
		"took", time.Since(timeBeforeQuery))

	// WAL Checkpoint (aka commit anything in the WAL to the main db).
	// Do this after vacuum too to ensure we have a clean slate for this
	// startup.
	// https://sqlite.org/pragma.html#pragma_wal_checkpoint
	timeBeforeQuery = time.Now()
	if res := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); res.Error != nil {
		slog.Error("optimize: Checkpoint failed!")
		return res.Error
	}
	slog.Info("optimize: Checkpointed.", "took", time.Since(timeBeforeQuery))

	slog.Info("optimize: Done.")
	return nil
}

// Optimize task that is scheduled and ran every whenever.
// So this func is for optimzations that we want to re-run every time the
// task is scheduled for.
func TaskOptimize(db *gorm.DB) error {
	slog.Info("TaskOptimize: Running optimizations.")

	// WAL Checkpoint (aka commit anything in the WAL to the main db).
	// To avoid our WAL file becoming huge, we checkpoint regularly.
	// https://sqlite.org/pragma.html#pragma_wal_checkpoint
	timeBeforeQuery := time.Now()
	if res := db.Exec("PRAGMA wal_checkpoint(TRUNCATE)"); res.Error != nil {
		slog.Error("TaskOptimize: Checkpoint failed!")
		return res.Error
	}
	slog.Info("TaskOptimize: Checkpointed.",
		"took", time.Since(timeBeforeQuery))

	// Optimize pragma.
	// No args for our task as recommended.
	// https://sqlite.org/pragma.html#pragma_optimize
	timeBeforeQuery = time.Now()
	if res := db.Exec("PRAGMA optimize"); res.Error != nil {
		slog.Error("TaskOptimize: Optimize pragma failed!")
		return res.Error
	}
	slog.Info("TaskOptimize: Optimize pragma succeeded.",
		"took", time.Since(timeBeforeQuery))

	slog.Info("TaskOptimize: Done.")
	return nil
}
