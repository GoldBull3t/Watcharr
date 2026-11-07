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

// WatchedTagBackup represents a row in the watched_tags many-to-many join table
type WatchedTagBackup struct {
	WatchedID uint `json:"watchedId" gorm:"column:watched_id"`
	TagID     uint `json:"tagId" gorm:"column:tag_id"`
}

// UserBackup wraps entity.User so fields with json:"-" or gorm:"->:false" are preserved in JSON backups
type UserBackup struct {
	entity.User
	PasswordHash string `json:"passwordHash,omitempty"`
	AvatarID     *uint  `json:"avatarId,omitempty"`
	Permissions  int    `json:"permissions"`
}

// UserServicesBackup wraps entity.UserServices so fields with json:"-" are preserved in JSON backups
type UserServicesBackup struct {
	entity.UserServices
	Name       entity.UserServiceName `json:"name"`
	ClientID   string                 `json:"clientId"`
	AuthToken  string                 `json:"authToken"`
	AuthToken2 string                 `json:"authToken2,omitempty"`
	UserID     uint                   `json:"userId"`
}

// WatchedBackup wraps entity.Watched so foreign key fields with json:"-" are preserved in JSON backups
type WatchedBackup struct {
	entity.Watched
	UserID    uint `json:"userId"`
	ContentID *int `json:"contentId,omitempty"`
	GameID    *int `json:"gameId,omitempty"`
}

// WatchedSeasonBackup wraps entity.WatchedSeason so fields with json:"-" are preserved in JSON backups
type WatchedSeasonBackup struct {
	entity.WatchedSeason
	UserID    uint `json:"userId"`
	WatchedID uint `json:"watchedId"`
}

// WatchedEpisodeBackup wraps entity.WatchedEpisode so fields with json:"-" are preserved in JSON backups
type WatchedEpisodeBackup struct {
	entity.WatchedEpisode
	UserID    uint `json:"userId"`
	WatchedID uint `json:"watchedId"`
}

// ActivityBackup wraps entity.Activity so fields with json:"-" are preserved in JSON backups
type ActivityBackup struct {
	entity.Activity
	UserID uint `json:"userId"`
}

// FollowBackup wraps entity.Follow so fields with json:"-" are preserved in JSON backups
type FollowBackup struct {
	entity.Follow
	UpdatedAt      time.Time `json:"updatedAt"`
	UserID         uint      `json:"userId"`
	FollowedUserID uint      `json:"followedUserId"`
}

// ImageBackup wraps entity.Image so ID and Hash (which have json:"-") are preserved in JSON backups
type ImageBackup struct {
	entity.Image
	ID   uint   `json:"id"`
	Hash string `json:"hash"`
}

// GameBackup wraps entity.Game so PosterID (which has json:"-") is preserved in JSON backups
type GameBackup struct {
	entity.Game
	PosterID *uint `json:"posterId,omitempty"`
}

// ArrRequestBackup wraps entity.ArrRequest so fields with json:"-" are preserved in JSON backups
type ArrRequestBackup struct {
	entity.ArrRequest
	UserID    uint `json:"userId"`
	ContentID *int `json:"contentId,omitempty"`
}

// TagBackup wraps entity.Tag so UserID (which has json:"-") is preserved in JSON backups
type TagBackup struct {
	entity.Tag
	UserID uint `json:"userId"`
}

// UniversalBackupDump provides a portable cross-database snapshot of all entities
type UniversalBackupDump struct {
	Version           string                    `json:"version"`
	ExportedAt        time.Time                 `json:"exportedAt"`
	Driver            string                    `json:"driver"`
	PasswordHashes    map[uint]string           `json:"passwordHashes,omitempty"`
	Users             []UserBackup              `json:"users"`
	UserServices      []UserServicesBackup      `json:"userServices"`
	Content           []entity.Content          `json:"content"`
	Watched           []WatchedBackup           `json:"watched"`
	WatchedTags       []WatchedTagBackup        `json:"watchedTags,omitempty"`
	Seasons           []WatchedSeasonBackup     `json:"seasons"`
	Episodes          []WatchedEpisodeBackup    `json:"episodes"`
	Activity          []ActivityBackup          `json:"activity"`
	Follow            []FollowBackup            `json:"follow"`
	Images            []ImageBackup             `json:"images"`
	Games             []GameBackup              `json:"games"`
	ArrRequests       []ArrRequestBackup        `json:"arrRequests"`
	Tags              []TagBackup               `json:"tags"`
	PushSubscriptions []entity.PushSubscription `json:"pushSubscriptions,omitempty"`
}
