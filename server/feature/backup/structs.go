package backup

import (
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
)

// BackupInfo represents metadata about an individual backup archive
type BackupInfo struct {
	Filename      string    `json:"filename"`
	Size          int64     `json:"size"`
	FormattedSize string    `json:"formattedSize"`
	CreatedAt     time.Time `json:"createdAt"`
	IsAuto        bool      `json:"isAuto"`
	Type          string    `json:"type"` // "sqlite" or "json"
}

// BackupConfigResponse returns current disaster recovery configuration and stats
type BackupConfigResponse struct {
	Enabled            bool       `json:"enabled"`
	Retention          int        `json:"retention"`
	ScheduleSeconds    int        `json:"scheduleSeconds"`
	NextRun            *time.Time `json:"nextRun,omitempty"`
	TotalBackups       int        `json:"totalBackups"`
	TotalSize          int64      `json:"totalSize"`
	FormattedTotalSize string     `json:"formattedTotalSize"`
	Driver             string     `json:"driver"`
	LastBackup         *time.Time `json:"lastBackup,omitempty"`
}

// UpdateBackupConfigRequest allows updating backup schedule and retention limits
type UpdateBackupConfigRequest struct {
	Enabled         *bool `json:"enabled"`
	Retention       *int  `json:"retention"`
	ScheduleSeconds *int  `json:"scheduleSeconds"`
}

// CreateBackupRequest allows triggering a manual backup with an optional format override
type CreateBackupRequest struct {
	Format string `json:"format"` // "auto", "sqlite", "json"
}

// UniversalBackupDump provides a portable cross-database snapshot of all entities
type UniversalBackupDump struct {
	Version      string                  `json:"version"`
	ExportedAt   time.Time               `json:"exportedAt"`
	Driver       string                  `json:"driver"`
	Users        []entity.User           `json:"users"`
	UserServices []entity.UserServices   `json:"userServices"`
	Content      []entity.Content        `json:"content"`
	Watched      []entity.Watched        `json:"watched"`
	Seasons      []entity.WatchedSeason  `json:"seasons"`
	Episodes     []entity.WatchedEpisode `json:"episodes"`
	Activity     []entity.Activity       `json:"activity"`
	Follow       []entity.Follow         `json:"follow"`
	Images       []entity.Image          `json:"images"`
	Games        []entity.Game           `json:"games"`
	ArrRequests  []entity.ArrRequest     `json:"arrRequests"`
	Tags         []entity.Tag            `json:"tags"`
}

