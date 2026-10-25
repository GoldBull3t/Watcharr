package season

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"

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

	newShowStatus, addedActivity, err := s.hookStatusChangedSetShowStatus(
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
		if addedActivity.ID != 0 {
			hookResponse.AddedActivities = append(
				hookResponse.AddedActivities, addedActivity)
		}
	}

	// When a season is marked as FINISHED, automatically mark all its episodes as FINISHED in db
	if newSeasonStatus == entity.FINISHED {
		eps, err := s.completeAllEpisodesForSeason(userID, watchedID, watchedEntry.Content.TmdbID, seasonNum)
		if err != nil {
			slog.Error("hookStatusChanged: Failed to complete episodes for season", "error", err)
			hookResponse.Errors = append(hookResponse.Errors, "failed to complete episodes for season")
		} else {
			hookResponse.WatchedEpisodes = eps
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
// Returns newShowStatus, addedActivity.
func (s *Service) hookStatusChangedSetShowStatus(
	userID uint,
	watchedID uint,
	tmdbID int,
	showWatchedStatus entity.WatchedStatus,
	seasonNum int,
	newSeasonStatus entity.WatchedStatus,
) (entity.WatchedStatus, entity.Activity, error) {
	var newShowStatus entity.WatchedStatus
	var reason string

	// If newSeasonStatus is WATCHING or FINISHED & the show is PLANNED, then
	// show status should be set to WATCHING (unless checks below change that).
	if (newSeasonStatus == entity.WATCHING || newSeasonStatus == entity.FINISHED) &&
		(showWatchedStatus == "" || showWatchedStatus == entity.PLANNED) {
		newShowStatus = entity.WATCHING
		reason = fmt.Sprintf(
			"Season %d was set to %s.",
			seasonNum,
			newSeasonStatus,
		)
	}

	// If all seasons are finished, the show status needs to be set to:
	// - PLANNED (if show is continuing)
	// - FINISHED (if show is ended/cancelled)
	if newSeasonStatus == entity.FINISHED {
		// NOTE: We only need to do this check if newSeasonStatus is FINISHED, since
		// if it was anything else, not all seasons can be FINISHED.
		allSeasonsFinished, err := s.allSeasonsCompletedForShow(
			userID, watchedID, tmdbID)
		if err != nil {
			slog.Error("hookStatusChangedSetShowStatus: Seasons completion check failed.",
				"error", err)
			return "", entity.Activity{}, err
		}
		if allSeasonsFinished {
			content, err := s.tmdb.ShowDetails(tmdb.ShowDetailsOptions{
				ID: strconv.Itoa(tmdbID),
			})
			if err != nil {
				slog.Error("hookStatusChangedSetShowStatus: Couldn't get ShowDetails",
					"error", err)
				return "", entity.Activity{}, err
			}

			if content.Status == tmdb.ShowStatusCanceled ||
				content.Status == tmdb.ShowStatusEnded {

				newShowStatus = entity.FINISHED
				reason = "All seasons were completed and the show is not continuing."
			} else {
				newShowStatus = entity.FINISHED
				reason = "All available seasons were completed."
			}
		}
	}

	// Finally save the newShowStatus

	if newShowStatus == "" || reason == "" {
		slog.Info("hookStatusChangedSetShowStatus: No new status or reason set. Not saving.")
		return "", entity.Activity{}, nil
	}

	res := s.db.
		Model(&entity.Watched{}).
		Where("user_id = ? AND id = ?", userID, watchedID).
		Update("status", newShowStatus)
	if res.Error != nil {
		slog.Error("hookStatusChanged: Failed to update show status!",
			"error", res.Error)
		return "", entity.Activity{}, errors.New("failed to update db")
	}
	if res.RowsAffected == 0 {
		slog.Error("hookStatusChanged: Nothing matched to update.")
		return "", entity.Activity{}, errors.New("watched entry does not exist")
	}

	addedActivity, _ := activity.
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

	return newShowStatus, addedActivity, nil
}
