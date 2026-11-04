package season

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/sbondCo/Watcharr/activity"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/media/tmdb"
	"github.com/sbondCo/Watcharr/util"
)

// Called after season status is changed.
// Automates setting other statuses.
func (s *Service) hookStatusChanged(
	userID uint,
	watchedID uint,
	seasonNum int,
	newSeasonStatus entity.WatchedStatus,
) domain.SeasonStatusChangedHookResponse {
	userSettings, err := s.userProvider.UserGetSettings(userID)
	if err != nil {
		slog.Error("hookStatusChanged: Failed to get user settings! Continuing.",
			"error", err)
	} else {
		if !util.Deref(userSettings.AutomateShowStatuses, true) {
			slog.Debug("hookStatusChanged: User has AutomateShowStatuses disabled. Skipping hook.",
				"user_id", userID)
			return domain.SeasonStatusChangedHookResponse{}
		}
	}

	hookResponse := domain.SeasonStatusChangedHookResponse{}

	// Get watched entry.
	// NOTE: Content type is validated from caller (SetWatchedSeason), so its
	// safe to not validate that here.
	watchedEntry, err := s.wp.GetWatchedItemById(userID, watchedID)
	if err != nil {
		slog.Error("hookStatusChanged: Failed to get watched entry!",
			"error", err)
		hookResponse.Errors = append(hookResponse.Errors, "failed to get watched entry")
		return hookResponse
	}

	// When a season is marked as FINISHED, automatically mark all its episodes as FINISHED in db first
	if newSeasonStatus == entity.FINISHED {
		eps, err := s.completeAllEpisodesForSeason(userID, watchedID, watchedEntry.Content.TmdbID, seasonNum)
		if err != nil {
			slog.Error("hookStatusChanged: Failed to complete episodes for season", "error", err)
			hookResponse.Errors = append(hookResponse.Errors, "failed to complete episodes for season")
		} else {
			hookResponse.WatchedEpisodes = eps
		}
	}

	newShowStatus, nextSeason, addedActivities, err := s.hookStatusChangedSetShowStatus(
		userID,
		watchedID,
		watchedEntry.Content.TmdbID,
		watchedEntry.Status,
		seasonNum,
		newSeasonStatus,
	)
	if err != nil {
		slog.Error("hookStatusChanged: Setting show status Failed!",
			"error", err)
		hookResponse.Errors = append(hookResponse.Errors, "failed to update show")
	} else {
		hookResponse.NewShowStatus = newShowStatus
		hookResponse.NextSeason = nextSeason
		if len(addedActivities) > 0 {
			hookResponse.AddedActivities = append(
				hookResponse.AddedActivities, addedActivities...)
		}
	}

	return hookResponse
}

// Marks all episodes of a season as FINISHED in the database.
func (s *Service) completeAllEpisodesForSeason(
	userID uint,
	watchedID uint,
	tmdbID int,
	seasonNum int,
) ([]entity.WatchedEpisode, error) {
	sDetails, err := s.tmdb.SeasonDetails(strconv.Itoa(tmdbID), strconv.Itoa(seasonNum))
	if err != nil {
		slog.Error("completeAllEpisodesForSeason: Failed to get SeasonDetails", "error", err)
		return nil, err
	}
	if len(sDetails.Episodes) == 0 {
		return nil, nil
	}

	var updatedEpisodes []entity.WatchedEpisode
	for _, ep := range sDetails.Episodes {
		var we entity.WatchedEpisode
		err := s.db.Where(
			"user_id = ? AND watched_id = ? AND season_number = ? AND episode_number = ?",
			userID, watchedID, seasonNum, ep.EpisodeNumber,
		).First(&we).Error

		if err != nil {
			// Record doesn't exist, create it as FINISHED
			newEp := entity.WatchedEpisode{
				UserID:        userID,
				WatchedID:     watchedID,
				SeasonNumber:  seasonNum,
				EpisodeNumber: ep.EpisodeNumber,
				Status:        entity.FINISHED,
			}
			if createErr := s.db.Create(&newEp).Error; createErr == nil {
				updatedEpisodes = append(updatedEpisodes, newEp)
			} else {
				slog.Error("completeAllEpisodesForSeason: Failed to create episode", "error", createErr)
			}
		} else if we.Status != entity.FINISHED {
			we.Status = entity.FINISHED
			if saveErr := s.db.Save(&we).Error; saveErr == nil {
				updatedEpisodes = append(updatedEpisodes, we)
			} else {
				slog.Error("completeAllEpisodesForSeason: Failed to save episode", "error", saveErr)
			}
		} else {
			updatedEpisodes = append(updatedEpisodes, we)
		}
	}

	return updatedEpisodes, nil
}

// Set top level watched entry status.
// Returns newShowStatus, nextSeason, addedActivities, error.
func (s *Service) hookStatusChangedSetShowStatus(
	userID uint,
	watchedID uint,
	tmdbID int,
	showWatchedStatus entity.WatchedStatus,
	seasonNum int,
	newSeasonStatus entity.WatchedStatus,
) (entity.WatchedStatus, *entity.WatchedSeason, []entity.Activity, error) {
	var newShowStatus entity.WatchedStatus
	var nextSeasonPtr *entity.WatchedSeason
	var addedActivities []entity.Activity
	var reason string

	// 1. Season was set to WATCHING:
	// Transition show status to WATCHING if not already WATCHING.
	if newSeasonStatus == entity.WATCHING {
		if showWatchedStatus != entity.WATCHING {
			newShowStatus = entity.WATCHING
			reason = fmt.Sprintf("Season %d was set to %s.", seasonNum, newSeasonStatus)
		}
	} else if newSeasonStatus == entity.PLANNED {
		// 2. Season was set to PLANNED:
		// Check if any other season is currently WATCHING.
		watchingCount, err := s.getNumberOfWatchedSeasonsInShow(userID, watchedID, []entity.WatchedStatus{entity.WATCHING})
		if err != nil {
			slog.Error("hookStatusChangedSetShowStatus: Failed to get watching count", "error", err)
		}
		if watchingCount == 0 && showWatchedStatus != entity.PLANNED {
			newShowStatus = entity.PLANNED
			reason = fmt.Sprintf("Season %d was set to %s.", seasonNum, newSeasonStatus)
		}
	} else if newSeasonStatus == entity.FINISHED {
		// 3. Season was set to FINISHED:
		// Check if any other season is currently WATCHING.
		watchingCount, err := s.getNumberOfWatchedSeasonsInShow(userID, watchedID, []entity.WatchedStatus{entity.WATCHING})
		if err != nil {
			slog.Error("hookStatusChangedSetShowStatus: Failed to get watching count", "error", err)
		}

		if watchingCount > 0 {
			// Another season is still actively being watched; show stays WATCHING
			slog.Debug("hookStatusChangedSetShowStatus: Another season is WATCHING; keeping show status WATCHING.")
		} else {
			// No season is actively WATCHING. Check TMDB for next seasons and overall show status.
			content, err := s.tmdb.ShowDetails(tmdb.ShowDetailsOptions{
				ID: strconv.Itoa(tmdbID),
			})
			if err != nil {
				slog.Error("hookStatusChangedSetShowStatus: Couldn't get ShowDetails", "error", err)
				return "", nil, nil, err
			}

			// Find next uncompleted regular season
			var nextUncompletedSeason int = 0
			if seasonNum > 0 && content.NumberOfSeasons > 0 {
				for nextS := seasonNum + 1; nextS <= int(content.NumberOfSeasons); nextS++ {
					var ws entity.WatchedSeason
					err := s.db.Where("user_id = ? AND watched_id = ? AND season_number = ?", userID, watchedID, nextS).First(&ws).Error
					if err != nil {
						// Does not exist in WatchedSeason yet -> this is the next season to plan
						nextUncompletedSeason = nextS
						break
					}
					if ws.Status != entity.FINISHED && ws.Status != entity.DROPPED {
						// Exists and not finished/dropped -> this is the next season to plan
						nextUncompletedSeason = nextS
						break
					}
				}
			}

			if nextUncompletedSeason > 0 {
				// We found an upcoming season: plan it automatically!
				var nextSeason entity.WatchedSeason
				err := s.db.Where("user_id = ? AND watched_id = ? AND season_number = ?", userID, watchedID, nextUncompletedSeason).First(&nextSeason).Error
				if err != nil {
					// Create new WatchedSeason as PLANNED
					nextSeason = entity.WatchedSeason{
						UserID:       userID,
						WatchedID:    watchedID,
						SeasonNumber: nextUncompletedSeason,
						Status:       entity.PLANNED,
					}
					if createErr := s.db.Create(&nextSeason).Error; createErr == nil {
						nextSeasonPtr = &nextSeason
						act, actErr := activity.NewCreator(
							s.db,
							userID,
							watchedID,
							entity.SEASON_ADDED,
							false,
							entity.ActivityCreatedByWatcharr,
						).SetDataJSON(map[string]any{
							"season": nextUncompletedSeason,
							"status": entity.PLANNED,
						}).SetReason(fmt.Sprintf("Automatically planned after Season %d completed.", seasonNum)).Create()
						if actErr == nil && act.ID != 0 {
							addedActivities = append(addedActivities, act)
						}
					} else {
						slog.Error("hookStatusChangedSetShowStatus: Failed to create next planned season", "error", createErr)
					}
				} else if nextSeason.Status != entity.PLANNED {
					nextSeason.Status = entity.PLANNED
					if saveErr := s.db.Save(&nextSeason).Error; saveErr == nil {
						nextSeasonPtr = &nextSeason
						act, actErr := activity.NewCreator(
							s.db,
							userID,
							watchedID,
							entity.SEASON_STATUS_CHANGED,
							false,
							entity.ActivityCreatedByWatcharr,
						).SetDataJSON(map[string]any{
							"season": nextUncompletedSeason,
							"status": entity.PLANNED,
						}).SetReason(fmt.Sprintf("Automatically planned after Season %d completed.", seasonNum)).Create()
						if actErr == nil && act.ID != 0 {
							addedActivities = append(addedActivities, act)
						}
					} else {
						slog.Error("hookStatusChangedSetShowStatus: Failed to save next planned season", "error", saveErr)
					}
				} else {
					nextSeasonPtr = &nextSeason
				}

				newShowStatus = entity.PLANNED
				reason = fmt.Sprintf("Season %d was completed. Next season (Season %d) set to Planned.", seasonNum, nextUncompletedSeason)
			} else {
				// No more uncompleted seasons found up to NumberOfSeasons
				allSeasonsFinished, err := s.allSeasonsCompletedForShow(userID, watchedID, tmdbID)
				if err != nil {
					slog.Error("hookStatusChangedSetShowStatus: Seasons completion check failed.", "error", err)
				}
				if allSeasonsFinished {
					if content.Status == tmdb.ShowStatusCanceled || content.Status == tmdb.ShowStatusEnded {
						newShowStatus = entity.FINISHED
						reason = "All seasons were completed and the show is not continuing."
					} else {
						newShowStatus = entity.PLANNED
						reason = "All available seasons were completed and the show is continuing."
					}
				} else {
					// In case some previous season was skipped
					newShowStatus = entity.PLANNED
					reason = fmt.Sprintf("Season %d was completed.", seasonNum)
				}
			}
		}
	}

	// Save newShowStatus if it changed and reason is set
	if newShowStatus != "" && reason != "" && newShowStatus != showWatchedStatus {
		res := s.db.
			Model(&entity.Watched{}).
			Where("user_id = ? AND id = ?", userID, watchedID).
			Updates(map[string]any{
				"status":     newShowStatus,
				"updated_at": time.Now(),
			})
		if res.Error != nil {
			slog.Error("hookStatusChanged: Failed to update show status!",
				"error", res.Error)
			return "", nil, nil, errors.New("failed to update db")
		}
		if res.RowsAffected == 0 {
			slog.Error("hookStatusChanged: Nothing matched to update.")
			return "", nil, nil, errors.New("watched entry does not exist")
		}

		addedActivity, err := activity.
			NewCreator(
				s.db,
				userID,
				watchedID,
				entity.STATUS_CHANGED,
				false,
				entity.ActivityCreatedByWatcharr,
			).
			SetDataJSON(map[string]any{
				"status": newShowStatus,
			}).
			SetReason(reason).
			Create()
		if err != nil {
			slog.Error("hookStatusChangedSetShowStatus: Failed to create activity", "error", err)
		} else if addedActivity.ID != 0 {
			addedActivities = append(addedActivities, addedActivity)
		}
	} else if newShowStatus == showWatchedStatus {
		newShowStatus = ""
	}

	return newShowStatus, nextSeasonPtr, addedActivities, nil
}
