package watched

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/sbondCo/Watcharr/activity"
	"github.com/sbondCo/Watcharr/database/entity"
	"gorm.io/gorm"
)

// AutoFinishCompletedMovies searches for movies currently in WATCHING status
// whose playback started more than (runtime + 15 minutes) ago, and marks them
// as FINISHED automatically with zero user interaction.
func AutoFinishCompletedMovies(db *gorm.DB) (int64, error) {
	if db == nil {
		return 0, nil
	}

	var watchingEntries []entity.Watched
	err := db.Model(&entity.Watched{}).
		Preload("Content").
		Where("status = ?", entity.WATCHING).
		Find(&watchingEntries).Error
	if err != nil {
		slog.Error("AutoFinishCompletedMovies: Failed to fetch watching entries", "error", err)
		return 0, err
	}

	if len(watchingEntries) == 0 {
		return 0, nil
	}

	now := time.Now().UTC()
	var finishedCount int64

	for _, w := range watchingEntries {
		if w.Content == nil || w.Content.Type != entity.MOVIE {
			continue
		}

		runtimeMinutes := w.Content.Runtime
		if runtimeMinutes == 0 {
			runtimeMinutes = 120 // Fallback to standard 2-hour duration if runtime unknown
		}

		// Total threshold: Runtime + 15 minutes of margin for pauses / post-credits
		durationThreshold := time.Duration(runtimeMinutes+15) * time.Minute
		if now.Sub(w.UpdatedAt) < durationThreshold {
			continue
		}

		// Movie has fully completed its duration!
		err := db.Transaction(func(tx *gorm.DB) error {
			res := tx.Model(&entity.Watched{}).
				Where("id = ? AND status = ?", w.ID, entity.WATCHING).
				Updates(map[string]any{
					"status":     entity.FINISHED,
					"updated_at": now,
				})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return nil
			}

			_, actErr := activity.NewCreator(
				tx,
				w.UserID,
				w.ID,
				entity.STATUS_CHANGED,
				false,
				entity.ActivityCreatedByWatcharr,
			).
				SetData(string(entity.FINISHED)).
				SetReason(fmt.Sprintf("Filme concluído automaticamente após %d min de reprodução.", runtimeMinutes)).
				SetCountAsPlay(true).
				Create()
			if actErr != nil {
				slog.Warn("AutoFinishCompletedMovies: Failed to create completion activity", "error", actErr)
			}

			InvalidateUserWatchedCache(w.UserID)
			return nil
		})

		if err == nil {
			finishedCount++
			slog.Info("AutoFinishCompletedMovies: Movie marked as FINISHED automatically",
				"userId", w.UserID,
				"watchedId", w.ID,
				"title", w.Content.Title,
				"runtimeMinutes", runtimeMinutes)
		} else {
			slog.Error("AutoFinishCompletedMovies: Failed to complete movie", "error", err, "watchedId", w.ID)
		}
	}

	return finishedCount, nil
}
