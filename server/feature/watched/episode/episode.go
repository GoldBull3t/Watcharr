package episode

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/sbondCo/Watcharr/activity"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/integration"
	"github.com/sbondCo/Watcharr/media/tmdb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WatchedProvider interface {
	GetWatchedItemById(userId uint, id uint) (entity.Watched, error)
	IsWatchedItemContentType(userId uint, id uint, ct entity.ContentType) error
}

type WatchedSeasonProvider interface {
	GetWatchedSeason(userId uint, watchedId uint, seasonNumber int) (*entity.WatchedSeason, error)
	SetWatchedSeason(userId uint, ar domain.WatchedSeasonSetRequest) (domain.WatchedSeasonSetResponse, error)
}

type UserProvider interface {
	UserGetSettings(userId uint) (entity.UserSettings, error)
}

type Service struct {
	db           *gorm.DB
	wp           WatchedProvider
	wsp          WatchedSeasonProvider
	tmdb         *tmdb.TMDB
	userProvider UserProvider
}

func NewService(
	db *gorm.DB,
	wp WatchedProvider,
	wsp WatchedSeasonProvider,
	tmdb *tmdb.TMDB,
	userProvider UserProvider,
) *Service {
	return &Service{
		db,
		wp,
		wsp,
		tmdb,
		userProvider,
	}
}

// Get WatchedEpisode.
// Returns nil for entity.WatchedEpisode if it doesn't exist.
func (s *Service) GetWatchedEpisode(
	userId uint,
	watchedId uint,
	seasonNumber int,
	episodeNumber int,
) (*entity.WatchedEpisode, error) {
	var we *entity.WatchedEpisode
	err := s.db.
		Model(&entity.WatchedEpisode{}).
		Where("watched_id = ? AND user_id = ? AND season_number = ? AND episode_number = ?",
			watchedId, userId, seasonNumber, episodeNumber).
		Take(&we).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Debug("GetWatchedEpisode: Record not found.")
			return nil, nil
		}
		slog.Error("GetWatchedEpisode: Query failed!", "error", err)
		return &entity.WatchedEpisode{}, errors.New("failed to get watched episode")
	}
	return we, nil
}

// Add/edit a watched episode.
func (s *Service) SetWatchedEpisode(
	userId uint,
	ar domain.WatchedEpisodeSetRequest,
) (domain.WatchedEpisodeSetResponse, error) {
	slog.Debug("SetWatchedEpisode: Setting.",
		"userId", userId,
		"watchedID", ar.WatchedID,
		"season", ar.SeasonNumber,
		"episode", ar.EpisodeNumber)

	// Make sure watched item exists and is the correct type (TV)
	err := s.wp.IsWatchedItemContentType(userId, ar.WatchedID, entity.SHOW)
	if err != nil {
		slog.Error("SetWatchedEpisode: Failed!", "error", err)
		return domain.WatchedEpisodeSetResponse{},
			errors.New("failed or invalid watched entry targetted")
	}

	// Try to lookup existing Watched Episode.
	wEpisode, err := s.GetWatchedEpisode(
		userId, ar.WatchedID, ar.SeasonNumber, ar.EpisodeNumber)
	if err != nil {
		slog.Error("SetWatchedEpisode: WatchedEpisode query failed!",
			"error", err)
		return domain.WatchedEpisodeSetResponse{},
			errors.New("couldn't get watched episode")
	}

	resp := domain.WatchedEpisodeSetResponse{}
	activityMC := activity.NewMultiCreator()

	// Add or update Watched Episode.
	if wEpisode != nil {
		slog.Debug("SetWatchedEpisode: Updating existing.")

		if ar.Status != "" && ar.Status != wEpisode.Status {
			wEpisode.Status = ar.Status
			resp.Update = true
			activity.
				NewCreator(
					s.db,
					userId,
					ar.WatchedID,
					entity.EPISODE_STATUS_CHANGED,
					false,
					ar.ActivityCreatedBy,
				).
				SetDataJSON(map[string]any{
					"season":  ar.SeasonNumber,
					"episode": ar.EpisodeNumber,
					"status":  ar.Status,
				}).
				AddToMultiCreator(activityMC)
		}

		if ar.Rating != 0 && ar.Rating != wEpisode.Rating {
			wEpisode.Rating = ar.Rating
			resp.Update = true
			activity.
				NewCreator(
					s.db,
					userId,
					ar.WatchedID,
					entity.EPISODE_RATING_CHANGED,
					false,
					ar.ActivityCreatedBy,
				).
				SetDataJSON(map[string]any{
					"season":  ar.SeasonNumber,
					"episode": ar.EpisodeNumber,
					"rating":  ar.Rating,
				}).
				AddToMultiCreator(activityMC)
		}

		// If Update still = False, then no changes were made! Exit early.
		if !resp.Update {
			slog.Warn("SetWatchedEpisode: No changes were necessary.")
			// We don't return an error so we retain idempotency for this method.
			// Update must be set to True before returning the response, otherwise
			// we are indicating that the WatchedEpisode was just Created, which
			// could cause bugs in the client.
			// Made a new Response object instead of returning `resp` to avoid
			// bugs where we alter something above in it and it gets returned
			// here mistakenly.
			return domain.WatchedEpisodeSetResponse{
				Update:         true,
				WatchedEpisode: *wEpisode,
			}, nil
		}
	} else {
		slog.Debug("SetWatchedEpisode: Creating new.")

		wEpisode = &entity.WatchedEpisode{
			UserID:        userId,
			WatchedID:     ar.WatchedID,
			SeasonNumber:  ar.SeasonNumber,
			EpisodeNumber: ar.EpisodeNumber,
			Status:        ar.Status,
			Rating:        ar.Rating,
		}

		activity.
			NewCreator(
				s.db,
				userId,
				ar.WatchedID,
				entity.EPISODE_ADDED,
				false,
				ar.ActivityCreatedBy,
			).
			SetDataJSON(map[string]any{
				"season":  ar.SeasonNumber,
				"episode": ar.EpisodeNumber,
				"status":  ar.Status,
				"rating":  ar.Rating,
			}).
			SetCustomDate(&ar.AddActivityDate).
			SetReason(ar.AddActivityReason).
			AddToMultiCreator(activityMC)
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(wEpisode).Error; err != nil {
			slog.Debug("SetWatchedEpisode: Save query failed.", "error", err)
			return errors.New("failed to save")
		}

		// Touch Watched entry UpdatedAt so Continue Watching and Up Next re-sort immediately
		if err := tx.Model(&entity.Watched{}).Where("user_id = ? AND id = ?", userId, ar.WatchedID).Update("updated_at", time.Now()).Error; err != nil {
			return err
		}

		addedActivities, err := activityMC.CreateAllTx(tx)
		if err != nil {
			return err
		}

		resp.WatchedEpisode = *wEpisode
		resp.AddedActivities = addedActivities
		return nil
	})
	if err != nil {
		return domain.WatchedEpisodeSetResponse{}, err
	}

	if ar.Status != "" {
		slog.Debug("SetWatchedEpisode: Episode status was changed, calling hook.")
		resp.StatusChangedHookResponse =
			s.hookStatusChanged(
				userId,
				ar.WatchedID,
				ar.SeasonNumber,
				ar.EpisodeNumber,
				ar.Status,
			)

		go func() {
			var w entity.Watched
			if s.db.Preload("Content").Where("id = ? AND user_id = ?", ar.WatchedID, userId).First(&w).Error == nil && w.Content != nil {
				evt := "media.episode_status_changed"
				if ar.Status == entity.FINISHED {
					evt = "media.finished"
				}
				integration.DispatchActivityWebhook(s.db, userId, integration.ActivityWebhookPayload{
					Event:     evt,
					Status:    string(ar.Status),
					MediaType: "tv",
					Title:     w.Content.Title,
					Season:    ar.SeasonNumber,
					Episode:   ar.EpisodeNumber,
					Poster:    integration.BuildPosterURL(w.Content.PosterPath),
				})
			}
		}()
	}

	return resp, nil
}

// Remove a watched episode
func (s *Service) rmWatchedEpisode(userId uint, id uint) (entity.Activity, error) {
	slog.Debug("rmWatchedEpisode called", "user_id", userId, "id", id)
	var (
		watchedEpisode entity.WatchedEpisode
		addedActivity  entity.Activity
	)
	err := s.db.Transaction(func(tx *gorm.DB) error {
		resp := tx.
			Clauses(clause.Returning{}).
			Model(&entity.WatchedEpisode{}).
			Unscoped().
			Where("id = ? AND user_id = ?", id, userId).
			Delete(&watchedEpisode)
		if resp.Error != nil {
			slog.Error("Failed when removing a watched episode", "error", resp.Error)
			return errors.New("failed when removing watched episode")
		}
		if resp.RowsAffected == 0 {
			slog.Error("Failed when removing a watched episode", "error", "zero rows affected")
			return errors.New("wasn't removed from db.. may not exist")
		}
		slog.Debug("rmWatchedEpisode, deleted row", "row", watchedEpisode)
		if watchedEpisode.ID != 0 {
			if err := tx.Model(&entity.Watched{}).Where("user_id = ? AND id = ?", userId, watchedEpisode.WatchedID).Update("updated_at", time.Now()).Error; err != nil {
				return err
			}
			jsonBytes, err := json.Marshal(map[string]interface{}{
				"season":  watchedEpisode.SeasonNumber,
				"episode": watchedEpisode.EpisodeNumber,
				"status":  watchedEpisode.Status,
				"rating":  watchedEpisode.Rating,
			})
			if err != nil {
				slog.Error("rmWatchedEpisode: failed to marshal json", "error", err)
			}
			act, err := activity.
				NewCreator(
					tx,
					userId,
					watchedEpisode.WatchedID,
					entity.EPISODE_REMOVED,
					false,
					0,
				).
				SetData(string(jsonBytes)).
				Create()
			if err != nil {
				slog.Error("rmWatchedEpisode: failed to create activity in tx", "error", err)
				return err
			}
			addedActivity = act
		}
		return nil
	})
	if err != nil {
		return entity.Activity{}, err
	}
	return addedActivity, nil
}

func (s *Service) getNumberOfWatchedEpisodesInSeason(
	userId uint,
	watchedId uint,
	seasonNumber int,
	acceptableStatus []entity.WatchedStatus,
) (int64, error) {
	var count int64
	err := s.db.
		Model(&entity.WatchedEpisode{}).
		Where("user_id = ? AND watched_id = ? AND season_number = ? AND status IN ?",
			userId, watchedId, seasonNumber, acceptableStatus).
		Count(&count).
		Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

// Checks if all the Episodes in a given Season are marked finished, returns
// true if so.
func (s *Service) allEpisodesCompletedForSeason(
	userId uint,
	watchedId uint,
	tmdbID int,
	seasonNum int,
) (bool, error) {
	tmdbIdStr := strconv.Itoa(tmdbID)
	seasonNumStr := strconv.Itoa(seasonNum)

	seasonDetails, err := s.tmdb.SeasonDetails(tmdbIdStr, seasonNumStr)
	if err != nil {
		slog.Error("allEpisodesCompletedForSeason: Failed to get season details!",
			"error", err)
		return false, err
	}
	allEpisodesCount := len(seasonDetails.Episodes)

	finishedEpisodesCount, err := s.getNumberOfWatchedEpisodesInSeason(
		userId, watchedId, seasonNum, []entity.WatchedStatus{entity.FINISHED, entity.DROPPED})
	if err != nil {
		slog.Error("allEpisodesCompletedForSeason: Failed to get number of watched episodes in this season!",
			"error", err)
		return false, err
	}
	slog.Debug("allEpisodesCompletedForSeason: Got episode counts.",
		"allEpisodesCount", allEpisodesCount,
		"finishedEpisodesCount", finishedEpisodesCount)

	if finishedEpisodesCount >= int64(allEpisodesCount) {
		return true, nil
	}
	return false, nil
}
