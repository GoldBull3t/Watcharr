package activity

import (
	"errors"
	"log/slog"
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"gorm.io/gorm"
)

func Get(
	db *gorm.DB,
	userId uint,
	watchedId uint,
) ([]entity.Activity, error) {
	activity := new([]entity.Activity)
	res := db.Model(&entity.Activity{}).
		Where("user_id = ? AND watched_id = ?", userId, watchedId).
		Find(&activity)
	if res.Error != nil {
		slog.Error("Get: Query failed!", "error", res.Error)
		return []entity.Activity{}, errors.New("failed getting activity")
	}
	return *activity, nil
}

// Update activity. Currently only allows updating custom_date (needs refactoring if other edits are allowed).
func Update(
	db *gorm.DB,
	userId uint,
	id uint,
	activityUpdateRequest domain.ActivityUpdateRequest,
) error {
	if id == 0 {
		return errors.New("id must be set to update an activity")
	}
	if activityUpdateRequest.CustomDate.IsZero() {
		return errors.New("customDate must be set to update an activity")
	}
	res := db.
		Model(&entity.Activity{}).
		Where("user_id = ? AND id = ?", userId, id).
		Update("custom_date", activityUpdateRequest.CustomDate)
	if res.Error != nil {
		slog.Error("Update: Query failed!", "error", res.Error)
		return errors.New("query failed")
	}
	if res.RowsAffected < 1 {
		slog.Error("Update: No activities were updated.")
		return errors.New("no rows affected")
	}
	slog.Debug("Update: Updated activity", "updated_activity", id)
	return nil
}

func Delete(db *gorm.DB, userId uint, id uint) error {
	if id == 0 {
		return errors.New("an id wasn't provided")
	}
	res := db.
		Unscoped().
		Where("user_id = ?", userId).
		Delete(&entity.Activity{}, id)
	if res.Error != nil {
		slog.Error("Delete: Query failed!", "error", res.Error)
		return errors.New("query failed")
	}
	if res.RowsAffected < 1 {
		slog.Error("Delete: No activities were deleted.")
		return errors.New("no rows affected")
	}
	return nil
}

// CleanupOldActivities permanently removes soft-deleted activities older than 90 days
// and legacy unsupported activity types (such as SEASON_ADDED_AUTO).
func CleanupOldActivities(db *gorm.DB) {
	if db == nil {
		return
	}
	cutoff := time.Now().AddDate(0, -3, 0)
	res := db.Unscoped().
		Where("(deleted_at IS NOT NULL AND deleted_at < ?) OR type = ?", cutoff, "SEASON_ADDED_AUTO").
		Delete(&entity.Activity{})
	if res.Error != nil {
		slog.Error("CleanupOldActivities: Failed to hard-delete old activities!", "error", res.Error)
		return
	}
	if res.RowsAffected > 0 {
		slog.Info("CleanupOldActivities: Purged old soft-deleted and legacy activities", "rows", res.RowsAffected)
	}
}
