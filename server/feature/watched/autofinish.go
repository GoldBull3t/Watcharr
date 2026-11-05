package watched

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/sbondCo/Watcharr/activity"
	"github.com/sbondCo/Watcharr/database/entity"
	"gorm.io/gorm"
)

// AutoFinishCompletedMovies is an alias for AutoFinishWatcharrMedia for backwards compatibility.
func AutoFinishCompletedMovies(db *gorm.DB) (int64, error) {
	return AutoFinishWatcharrMedia(db)
}

// AutoFinishWatcharrMedia searches for movies and TV episodes currently in WATCHING status
// whose playback started more than (runtime + margin) ago, and marks them
// as FINISHED automatically with zero user interaction, auto-advancing TV episodes.
func AutoFinishWatcharrMedia(db *gorm.DB) (int64, error) {
	if db == nil {
		return 0, nil
	}

	now := time.Now().UTC()
	var finishedCount int64

	// 1. Process Movies in WATCHING status
	var watchingEntries []entity.Watched
	err := db.Model(&entity.Watched{}).
		Preload("Content").
		Where("status = ?", entity.WATCHING).
		Find(&watchingEntries).Error
	if err != nil {
		slog.Error("AutoFinishWatcharrMedia: Failed to fetch watching entries", "error", err)
	} else {
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
			txErr := db.Transaction(func(tx *gorm.DB) error {
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
					slog.Warn("AutoFinishWatcharrMedia: Failed to create completion activity", "error", actErr)
				}

				InvalidateUserWatchedCache(w.UserID)
				return nil
			})

			if txErr == nil {
				finishedCount++
				slog.Info("AutoFinishWatcharrMedia: Movie marked as FINISHED automatically",
					"userId", w.UserID,
					"watchedId", w.ID,
					"title", w.Content.Title,
					"runtimeMinutes", runtimeMinutes)
			}
		}
	}

	// 2. Process TV Show Episodes in WATCHING status (Auto-Advance Series on Couch)
	var watchingEpisodes []entity.WatchedEpisode
	err = db.Model(&entity.WatchedEpisode{}).
		Where("status = ?", entity.WATCHING).
		Find(&watchingEpisodes).Error
	if err != nil {
		slog.Error("AutoFinishWatcharrMedia: Failed to fetch watching episodes", "error", err)
	} else {
		for _, ep := range watchingEpisodes {
			var w entity.Watched
			wErr := db.Model(&entity.Watched{}).
				Preload("Content").
				Where("id = ?", ep.WatchedID).
				First(&w).Error
			if wErr != nil || w.Content == nil || w.Content.Type != entity.SHOW {
				continue
			}

			epRuntime := w.Content.Runtime
			if epRuntime == 0 {
				epRuntime = 45 // Default 45 minutes for TV episode if unknown
			}

			// Threshold: Episode runtime + 10 minutes of margin
			threshold := time.Duration(epRuntime+10) * time.Minute
			if now.Sub(ep.UpdatedAt) < threshold {
				continue
			}

			// Episode has finished! Complete it and advance to next episode
			txErr := db.Transaction(func(tx *gorm.DB) error {
				// Mark current episode as FINISHED
				upRes := tx.Model(&entity.WatchedEpisode{}).
					Where("id = ? AND status = ?", ep.ID, entity.WATCHING).
					Updates(map[string]any{
						"status":     entity.FINISHED,
						"updated_at": now,
					})
				if upRes.Error != nil {
					return upRes.Error
				}
				if upRes.RowsAffected == 0 {
					return nil
				}

				// Record episode completion activity
				_, actErr := activity.NewCreator(
					tx,
					ep.UserID,
					ep.WatchedID,
					entity.STATUS_CHANGED,
					false,
					entity.ActivityCreatedByWatcharr,
				).
					SetData(string(entity.FINISHED)).
					SetReason(fmt.Sprintf("S%d:E%d concluído automaticamente após %d min.", ep.SeasonNumber, ep.EpisodeNumber, epRuntime)).
					SetCountAsPlay(true).
					Create()
				if actErr != nil {
					slog.Warn("AutoFinishWatcharrMedia: Failed to create episode activity", "error", actErr)
				}

				// Auto-advance to next episode in season
				nextEpNum := ep.EpisodeNumber + 1
				var nextEp entity.WatchedEpisode
				nextFindErr := tx.Where(
					"user_id = ? AND watched_id = ? AND season_number = ? AND episode_number = ?",
					ep.UserID, ep.WatchedID, ep.SeasonNumber, nextEpNum,
				).First(&nextEp).Error

				if nextFindErr != nil {
					// Create next episode as WATCHING
					newNext := entity.WatchedEpisode{
						UserID:        ep.UserID,
						WatchedID:     ep.WatchedID,
						SeasonNumber:  ep.SeasonNumber,
						EpisodeNumber: nextEpNum,
						Status:        entity.WATCHING,
					}
					_ = tx.Create(&newNext).Error
				} else if nextEp.Status != entity.FINISHED {
					// Update next episode to WATCHING
					_ = tx.Model(&entity.WatchedEpisode{}).Where("id = ?", nextEp.ID).Updates(map[string]any{
						"status":     entity.WATCHING,
						"updated_at": now,
					}).Error
				}

				// Touch parent show to keep it fresh at the top of Watching list
				_ = tx.Model(&entity.Watched{}).Where("id = ?", ep.WatchedID).Updates(map[string]any{
					"status":             entity.WATCHING,
					"last_viewed_season": ep.SeasonNumber,
					"updated_at":         now,
				}).Error

				InvalidateUserWatchedCache(ep.UserID)
				return nil
			})

			if txErr == nil {
				finishedCount++
				slog.Info("AutoFinishWatcharrMedia: Episode marked FINISHED & advanced to next episode",
					"userId", ep.UserID,
					"watchedId", ep.WatchedID,
					"season", ep.SeasonNumber,
					"episode", ep.EpisodeNumber)
			}
		}
	}

	return finishedCount, nil
}
