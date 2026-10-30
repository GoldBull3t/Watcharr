package watched

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"log/slog"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sbondCo/Watcharr/activity"
	"github.com/sbondCo/Watcharr/database/dbmodel"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/integration"
	"github.com/sbondCo/Watcharr/feature/watched/addedtocontent"
	"github.com/sbondCo/Watcharr/util"
	"github.com/sbondCo/Watcharr/media/tmdb"
	"gorm.io/gorm"
)

type ContentProvider interface {
	GetOrCacheContent(contentType entity.ContentType, tmdbId int) (entity.Content, error)
}

type GameProvider interface {
	GetOrCache(igdbID int) (entity.Game, error)
}

type UserProvider interface {
	UserGetSettings(userId uint) (entity.UserSettings, error)
}

type TMDBSeasonProvider interface {
	SeasonDetails(showId string, seasonNumber string, optParams ...map[string]string) (tmdb.SeasonDetails, error)
}

type Service struct {
	db           *gorm.DB
	cp           ContentProvider
	gameProvider GameProvider
	userProvider UserProvider
	tmdb         TMDBSeasonProvider
}

func NewService(
	db *gorm.DB,
	cp ContentProvider,
	gameProvider GameProvider,
	userProvider UserProvider,
	tmdbProvider ...TMDBSeasonProvider,
) *Service {
	var tp TMDBSeasonProvider
	if len(tmdbProvider) > 0 {
		tp = tmdbProvider[0]
	}
	return &Service{
		db,
		cp,
		gameProvider,
		userProvider,
		tp,
	}
}

// Get entire watched list
func (s *Service) getWatched(userId uint) ([]entity.Watched, error) {
	watched := new([]entity.Watched)
	res := s.db.Model(&entity.Watched{}).
		Preload("Content").
		Preload("Game").
		Preload("Game.Poster").
		Preload("Activity").
		Preload("WatchedSeasons").
		Preload("WatchedEpisodes").
		Preload("Tags").
		Where("user_id = ?", userId).
		Find(&watched)
	if res.Error != nil {
		slog.Error("getWatched: Failed!", "error", res.Error)
		return []entity.Watched{}, res.Error
	}
	return *watched, nil
}

// Returns a page of users watched list.
func (s *Service) GetWatchedPage(
	userId uint,
	pp util.PaginationParams,
	wr domain.WatchedGetPageRequest,
	extraProps *domain.WatchedGetPageExtraProps,
) (util.PaginationResponse[entity.Watched, util.None], error) {
	slog.Debug("GetWatchedPage: A page was requested.",
		"user_id", userId,
		"pagination_params", pp,
		"wr", wr)

	pRes := &util.PaginationResponse[entity.Watched, util.None]{}

	// Get user settings.
	userSettings, err := s.userProvider.UserGetSettings(userId)
	if err != nil {
		return *pRes, errors.New("failed to get user settings")
	}

	watched := new([]entity.Watched)
	res := s.db.
		Model(&entity.Watched{}).
		Where(&entity.Watched{UserID: userId})

	if extraProps != nil {
		// Process `WatchedIds` extra prop.
		if len(extraProps.WatchedIds) > 0 {
			res = res.Where("watcheds.id IN ?", extraProps.WatchedIds)
		}

		// Search query
		if extraProps.Query != "" {
			q := "%" + strings.ToLower(extraProps.Query) + "%"
			res = res.Where("LOWER(Content.title) LIKE ? OR LOWER(Game.name) LIKE ?", q, q)
		}
	}

	res = res.
		Joins("Content").
		Joins("Game").
		Preload("Game.Poster").
		Preload("Tags").
		// Apply filters first.
		Scopes(watchedRefineFilter(wr, &userSettings)).
		// Then count results (after filter);
		Count(&pRes.TotalResults).
		// Now calculate pagination properties with a TotalResults
		// that takes filtered out items into account.
		Scopes(util.Paginate(pp, pRes)).
		// Last we can apply our sorting.
		// Note: We must sort *after* we count, because currently
		// our 'Last Finished' sort, causes the COUNT query to also
		// be given extra JOINS that we don't want (because it slows it down).
		Scopes(watchedRefineSort(wr, userId)).
		Find(&watched)
	if res.Error != nil {
		slog.Error("GetWatchedPage: Failed!", "error", res.Error)
		return util.PaginationResponse[entity.Watched, util.None]{}, res.Error
	}
	pRes.Results = *watched
	pRes.Finished(pp)
	return *pRes, nil
}

// Get a users **public** watchlist.
func (s *Service) getPublicWatched(
	userId uint,
	username string,
	pp util.PaginationParams,
	wr domain.WatchedGetPageRequest,
) (util.PaginationResponse[entity.Watched, util.None], error) {
	slog.Debug("getPublicWatched: running",
		"user_id", userId, "username", username)

	// First we need to make sure the users list is public
	user := new(entity.User)
	// I figure we require knowlege of the users id and name to make it
	// harder to just type in random ids and see someones list.. dunno
	// if this is a thing we need but its here.. for now at least.
	res := s.db.
		Where("id = ? AND username = ?", userId, username).
		Take(&user)
	if res.Error != nil {
		slog.Error("getPublicWatched: Failed to get user.",
			"user_id", userId)
		return util.PaginationResponse[entity.Watched, util.None]{}, errors.New("failed to check privacy settings")
	}
	if user.Private != nil && *user.Private {
		slog.Error("getPublicWatched: This users list is private.",
			"user_id", userId)
		return util.PaginationResponse[entity.Watched, util.None]{}, errors.New("this watched list is private")
	}

	// Now we know the user is public, return their list
	watched := new([]entity.Watched)
	pRes := &util.PaginationResponse[entity.Watched, util.None]{}
	res = s.db.
		Model(&entity.Watched{}).
		Where(&entity.Watched{UserID: userId}).
		Joins("Content").
		Joins("Game").
		Preload("Game.Poster").
		Preload("Tags").
		// Apply filters first.
		Scopes(watchedRefineFilter(wr, nil)).
		// Then count results (after filter);
		Count(&pRes.TotalResults).
		// Now calculate pagination properties with a TotalResults
		// that takes filtered out items into account.
		Scopes(util.Paginate(pp, pRes)).
		// Sort options.
		// Note: See note above in GetWatchedPage.
		Scopes(watchedRefineSort(wr, userId)).
		Find(&watched)
	if res.Error != nil {
		slog.Error("getPublicWatched: Failed!", "error", res.Error)
		return util.PaginationResponse[entity.Watched, util.None]{}, errors.New("failed fetching the list")
	}
	pRes.Results = *watched
	pRes.Finished(pp)
	return *pRes, nil
}

// Get a watched list item by id (must be for `userId`).
func (s *Service) GetWatchedItemById(userId uint, id uint) (entity.Watched, error) {
	watched := new(entity.Watched)
	res := s.db.Model(&entity.Watched{}).
		Preload("Content").
		Where("user_id = ? AND id = ?", userId, id).
		Find(&watched)
	if res.Error != nil {
		slog.Error("GetWatchedItemById: Failed!", "error", res.Error)
		return entity.Watched{}, res.Error
	}
	return *watched, nil
}

// Get a watched list item ID by tmdb id for userID.
func (s *Service) GetWatchedItemIDByTmdbID(
	userId uint,
	tmdbId uint,
	contentType entity.ContentType,
) (uint, error) {
	slog.Debug("GetWatchedItemIDByTmdbID: Running.",
		"userId", userId, "tmdbId", tmdbId)
	var wID uint
	err := s.db.
		Model(&entity.Watched{}).
		Joins("INNER JOIN contents Content ON watcheds.content_id = Content.id").
		Select("watcheds.id").
		Where("user_id = ? AND Content.tmdb_id = ? AND Content.type = ?",
			userId, tmdbId, contentType).
		Scan(&wID).
		Error
	if err != nil {
		slog.Error("GetWatchedItemIDByTmdbID: Failed!", "error", err)
		return 0, err
	}
	slog.Debug("GetWatchedItemIDByTmdbID: Done.",
		"userId", userId, "tmdbId", tmdbId, "wID", wID)
	return wID, nil
}

// Get a watched list item by content (tmdb) id (must be for `userId`).
func (s *Service) GetWatchedItemByTmdbId(
	userId uint,
	tmdbId uint,
	contentType entity.ContentType,
) (entity.Watched, error) {
	slog.Debug("GetWatchedItemByTmdbId: Running.",
		"userId", userId, "tmdbId", tmdbId)
	watched := new(entity.Watched)
	res := s.db.Model(&entity.Watched{}).
		Joins("Content").
		Preload("Activity").
		Preload("WatchedSeasons").
		Preload("WatchedEpisodes").
		Preload("Tags").
		Where("user_id = ? AND Content.tmdb_id = ? AND Content.type = ?",
			userId, tmdbId, contentType).
		Take(&watched)
	if res.Error != nil {
		slog.Error("GetWatchedItemByTmdbId: Failed!", "error", res.Error)
		return entity.Watched{}, res.Error
	}
	slog.Debug("GetWatchedItemByTmdbId: Done.",
		"userId", userId, "tmdbId", tmdbId, "watched_item", watched)
	return *watched, nil
}

// Same as `getWatchedItemByTmdbId` except for getting in bulk (multiple content ids).
// `c` entries should be in format: [tmdb_id, ContentType] (Note: Couldn't figure out
// if it's possible to type this to enforce [int, ContentType] type for entries)
func (s *Service) GetWatchedItemsByTmdbIds(userId uint, c [][]any) ([]entity.Watched, error) {
	slog.Debug("GetWatchedItemsByTmdbIds: Running.", "userId", userId, "c", c)
	watched := new([]entity.Watched)
	res := s.db.Model(&entity.Watched{}).
		Joins("Content").
		Preload("Activity").
		Preload("WatchedSeasons").
		Preload("WatchedEpisodes").
		Preload("Tags").
		Where("user_id = ?", userId).
		Where("(Content.tmdb_id, Content.type) IN ?", c).
		Find(&watched)
	if res.Error != nil {
		slog.Error("GetWatchedItemsByTmdbIds: Failed!", "error", res.Error)
		return []entity.Watched{}, res.Error
	}
	slog.Debug(
		"GetWatchedItemsByTmdbIds: Done.",
		"userId", userId,
		"watcheds_found", len(*watched),
		// "wdev", *watched,
	)
	return *watched, nil
}

// Checks if Watched entry is of `ct`.
//
// This method is useful when we don't already have/need the whole Watched
// object fetched from the db and only need to validate it exists and is linked
// to a `Content` with a certain type.
//
// Returns error if query failed or type doesn't match.
func (s *Service) IsWatchedItemContentType(userId uint, id uint, ct entity.ContentType) error {
	var wc int64
	err := s.db.
		Model(&entity.Watched{}).
		Joins(`INNER JOIN contents ON contents.id = watcheds.content_id
			AND contents.type = ?`, ct).
		Where("watcheds.id = ? AND watcheds.user_id = ?", id, userId).
		Count(&wc).
		Error
	if err != nil {
		slog.Error("IsWatchedItemContentType: Query failed!", "error", err)
		return errors.New("failed when retrieving watched item")
	}
	if wc <= 0 {
		slog.Debug("IsWatchedItemContentType: NO.", "wc", wc)
		return errors.New("no show watched entry exists")
	}
	return nil
}

// Get a watched list item by game (igdb) id (must be for `userId`).
func (s *Service) GetWatchedItemByIgdbId(userId uint, igdbId uint) (entity.Watched, error) {
	slog.Debug("getWatchedItemByIgdbId: Running.", "userId", userId, "igdbId", igdbId)
	watched := new(entity.Watched)
	res := s.db.Model(&entity.Watched{}).
		Joins("Game").
		Preload("Game.Poster").
		Preload("Activity").
		Preload("Tags").
		Where("user_id = ? AND Game.igdb_id = ?", userId, igdbId).
		Take(&watched)
	if res.Error != nil {
		slog.Error("getWatchedItemByIgdbId: Failed!", "error", res.Error)
		return entity.Watched{}, res.Error
	}
	slog.Debug("getWatchedItemByIgdbId: Done.", "userId", userId, "igdbId", igdbId, "watched_item", watched)
	return *watched, nil
}

// Same as `getWatchedItemByIgdbId` except for getting in bulk (multiple content ids).
// `c` should be a slice of igdb ids.
func (s *Service) GetWatchedItemsByIgdbIds(userId uint, c []int) ([]entity.Watched, error) {
	slog.Debug("getWatchedItemsByIgdbIds: Running.", "userId", userId, "c", c)
	watched := new([]entity.Watched)
	res := s.db.Model(&entity.Watched{}).
		Joins("Game").
		Preload("Game.Poster").
		Preload("Activity").
		Preload("Tags").
		Where("user_id = ?", userId).
		Where("(Game.igdb_id) IN ?", c).
		Find(&watched)
	if res.Error != nil {
		slog.Error("getWatchedItemsByIgdbIds: Failed!", "error", res.Error)
		return []entity.Watched{}, res.Error
	}
	slog.Debug(
		"getWatchedItemsByIgdbIds: Done.",
		"userId", userId,
		"watcheds_found", len(*watched),
		// "wdev", *watched,
	)
	return *watched, nil
}

// Get watched item by an id and SupportedMedia type.
func (s *Service) GetWatchedItemBySupportedMediaId(userId uint, id uint, t util.SupportedMedia) (entity.Watched, error) {
	switch t {
	case util.SupportedMediaGame:
		return s.GetWatchedItemByIgdbId(userId, id)
	case util.SupportedMediaMovie:
		return s.GetWatchedItemByTmdbId(userId, id, entity.MOVIE)
	case util.SupportedMediaShow:
		return s.GetWatchedItemByTmdbId(userId, id, entity.SHOW)
	}
	slog.Error("GetWatchedItemBySupportedMediaId: Unsupported supportedmedia type",
		"type", t)
	return entity.Watched{}, errors.New("unsupported supportedmedia type")
}

// Get a list of watched items by a slice of ids and SupportedMedia types.
func (s *Service) GetWatchedItemsBySupportedMediaIds(userId uint, c []addedtocontent.IdToTypePair) ([]entity.Watched, error) {
	slog.Debug("GetWatchedItemsBySupportedMediaIds: Running.", "userId", userId, "c", c)
	// First we want to separate `c` into slices we can pass to the respective functions.
	tmdbIds := [][]any{}
	igdbIds := []int{}
	for _, v := range c {
		switch v.Type {
		case util.SupportedMediaMovie:
			tmdbIds = append(tmdbIds, []any{v.Id, entity.MOVIE})
		case util.SupportedMediaShow:
			tmdbIds = append(tmdbIds, []any{v.Id, entity.SHOW})
		case util.SupportedMediaGame:
			igdbIds = append(igdbIds, v.Id)
		}
	}
	// Now call each function relating to an overarching type.
	watcheds := []entity.Watched{}
	if len(tmdbIds) > 0 {
		if w, err := s.GetWatchedItemsByTmdbIds(userId, tmdbIds); err == nil {
			watcheds = append(watcheds, w...)
		} else {
			slog.Error("GetWatchedItemsBySupportedMediaIds: Failed to get items by tmdb ids.", "error", err)
			return []entity.Watched{}, err
		}
	}
	if len(igdbIds) > 0 {
		if w, err := s.GetWatchedItemsByIgdbIds(userId, igdbIds); err == nil {
			watcheds = append(watcheds, w...)
		} else {
			slog.Error("GetWatchedItemsBySupportedMediaIds: Failed to get items by igdb ids.", "error", err)
			return []entity.Watched{}, err
		}
	}
	return watcheds, nil
}

func (s *Service) AddWatched(
	userId uint,
	ar domain.WatchedAddRequest,
	extraProps domain.WatchedAddExtraProps,
) (entity.Watched, error) {
	slog.Debug("Adding watched item",
		"user_id", userId,
		"add_request", ar,
		"extra_props", extraProps)

	watched := entity.Watched{
		UserID: userId,
	}

	// Based on the ContentType given in the request, find the correct
	// content to attach to the Watched entry we are creating.
	switch ar.ContentType {
	case "movie", "tv":
		tmdbId := ar.TMDBID
		if tmdbId == 0 && ar.Deprecated_ContentID != 0 {
			// This if provides support for a now deprecated field.
			// When the property is removed, this if will be gone and we can
			// use `ar.TMDBID` directly for below.
			tmdbId = ar.Deprecated_ContentID
			slog.Error("AddWatched: You are using a deprecated field 'contentId', please replace it with 'tmdbId' to keep the same functionality for future releases of Watcharr.")
		}
		if tmdbId == 0 {
			return entity.Watched{}, errors.New("missing tmdb id")
		}
		// Get content cache (or cache it if we don't have it locally)
		content, err := s.cp.GetOrCacheContent(
			entity.ContentType(ar.ContentType), tmdbId)
		if err != nil {
			return entity.Watched{}, err
		}
		// Error if content has no id
		if content.ID == 0 {
			return entity.Watched{}, errors.New("failed to find content id")
		}
		// Add content to watched entry
		watched.ContentID = &content.ID
	case "game":
		if ar.IGDBID == 0 {
			return entity.Watched{}, errors.New("missing igdb id")
		}
		game, err := s.gameProvider.GetOrCache(ar.IGDBID)
		if err != nil {
			return entity.Watched{}, err
		}
		// Error if content has no id
		if game.ID == 0 {
			return entity.Watched{}, errors.New("failed to find game by id")
		}
		watched.GameID = &game.ID
	default:
		return entity.Watched{}, errors.New("invalid content type provided")
	}

	// Set default status for when content is added by
	// rating it instead of giving status first.
	if ar.Status == "" {
		if ar.ContentType == "movie" || ar.ContentType == "game" {
			ar.Status = entity.FINISHED
		} else {
			ar.Status = entity.WATCHING
		}
	}

	// Add watched data to struct
	watched.Status = ar.Status
	watched.Rating = ar.Rating
	watched.Thoughts = ar.Thoughts
	watched.HoursPlayed = ar.HoursPlayed
	watched.Platform = ar.Platform
	watched.GameCompletionStatus = ar.GameCompletionStatus

	// If custom WatchedDate passed, set CreatedAt and UpdatedAt fields to it.
	if !ar.WatchedDate.IsZero() {
		slog.Debug("Adding watched item: The provided WatchedDate is valid.",
			"userId", userId,
			"request", ar)
		watched.CreatedAt = ar.WatchedDate
		watched.UpdatedAt = ar.WatchedDate
	}

	// Create the entry in db.
	res := s.db.Create(&watched)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrDuplicatedKey) {
			// The record already exists.
			// It could be soft deleted or not.
			slog.Info("AddWatched: Creating row failed because of a unique key constraint violation.")
			if err := s.restoreWatchedAfterDuplicatedKeyErr(
				userId,
				ar,
				extraProps.DontRestore,
				&watched,
			); err != nil {
				// Try to restore the entry if unique contraint hit.
				slog.Error("AddWatched: Failed to restore existing watched entry.",
					"error", err)
				// Returns watched too because handlers of certain errors
				// may need it (and it's ID since we could have fetched it here)
				return watched, err
			}
		} else {
			slog.Error("AddWatched: Error adding watched to database", "error", res.Error.Error())
			return entity.Watched{}, errors.New("failed adding watched entry to database")
		}
	}
	slog.Debug("AddWatched: Added watched list item", "item", watched)

	// Finally add activity
	activityAddReq := activity.NewCreator(
		s.db,
		userId,
		watched.ID,
		entity.ADDED_WATCHED,
		false,
		extraProps.ActivityCreatedBy,
	)
	if activityJson, err := json.Marshal(map[string]any{
		"status": ar.Status,
		"rating": ar.Rating,
	}); err != nil {
		slog.Error("AddWatched: Failed to marshal json for add activity request, adding without data",
			"error", err)
	} else {
		activityAddReq.SetData(string(activityJson))
	}
	if ar.Status == entity.FINISHED {
		activityAddReq.SetCountAsPlay(true)
	}
	act, _ := activityAddReq.Create()
	watched.Activity = append(watched.Activity, act)

	return watched, nil
}

// Restore a watched entry that was soft deleted.
// Currently used for AddWatched, when it realizes the entry may exist already
// as a soft deleted record.
//
// NOTE: This function assumes it is being ran after a unique constraint
// violation (like in AddWatched).
func (s *Service) restoreWatchedAfterDuplicatedKeyErr(
	userId uint,
	ar domain.WatchedAddRequest,
	dontRestore bool,
	watchedOut *entity.Watched,
) error {
	slog.Info("restoreWatchedAfterDuplicatedKeyErr: Attempting to restore.",
		"user_id", userId)

	// Our base where statement to find the row we want to restore by unique
	// indexes (user_id AND (content_id or game_id)).
	whereStmt := entity.Watched{
		UserID: userId,
	}

	// A check to ensure safety of the following db query:
	// We MUST have a UserID
	if whereStmt.UserID == 0 {
		return errors.New("no userid in provided watched struct")
	}
	// ...AND a (contentId || gameId).
	if watchedOut.ContentID != nil && *watchedOut.ContentID != 0 {
		whereStmt.ContentID = watchedOut.ContentID
	} else if watchedOut.GameID != nil && *watchedOut.GameID != 0 {
		whereStmt.GameID = watchedOut.GameID
	} else {
		return errors.New("no supported media ids in provided watched struct")
	}

	// First we will try to get the record.
	res := s.db.
		Model(&entity.Watched{}).
		Unscoped().
		// Just loading activity now too so if we restore record,
		// we can give the full activity in response.
		Preload("Activity").
		Where(&whereStmt).
		Take(&watchedOut)
	if res.Error != nil {
		slog.Error("restoreWatchedAfterDuplicatedKeyErr: Select query failed!",
			"error", res.Error)
		return errors.New("errored checking for potential soft deleted record")
	}
	if watchedOut.ID == 0 {
		return errors.New("didn't find an existing record")
	}
	if watchedOut.DeletedAt.Time.IsZero() {
		// If it's not deleted, then it just exists..
		return domain.ErrWatchedExists
	} else if dontRestore {
		// If we are told not to attempt a restore (eg by a sync-er).
		return domain.ErrWatchedExistsSoftDeleted
	}

	// If we are here, we have a soft deleted record to restore:
	res = s.db.
		// Passing `&watchedOut` makes gorm do another query for inserting
		// all existing rows into db, im not sure why... but just going to
		// manually pass primary key in.
		Model(&entity.Watched{
			GormModel: dbmodel.GormModel{ID: watchedOut.ID},
		}).
		Unscoped().
		Updates(map[string]any{
			"status":     ar.Status,
			"rating":     ar.Rating,
			"thoughts":   ar.Thoughts,
			"deleted_at": nil,
		})
	if res.Error != nil || res.RowsAffected == 0 {
		slog.Error("restoreWatchedAfterDuplicatedKeyErr: Updating record failed!",
			"error", res.Error,
			"rows_affected", res.RowsAffected)
		return errors.New("failed to update record")
	}
	// We have updated the row, update our struct now.
	watchedOut.Status = ar.Status
	watchedOut.Rating = ar.Rating
	watchedOut.Thoughts = ar.Thoughts
	watchedOut.DeletedAt = gorm.DeletedAt{}

	slog.Info("restoreWatchedAfterDuplicatedKeyErr: Updated record.",
		"user_id", userId)

	return nil
}

// Update Watched entry.
func (s *Service) UpdateWatched(
	userId uint,
	id uint,
	ar domain.WatchedUpdateRequest,
) (domain.WatchedUpdateResponse, error) {
	slog.Debug("UpdateWatched", "request_data", ar)

	if err := ar.Valid(); err != nil {
		slog.Error("UpdateWatched: UpdateRequest struct is invalid.",
			"id", id, "error", err)
		return domain.WatchedUpdateResponse{}, err
	}

	upwat := entity.Watched{}
	err := s.db.
		Model(&entity.Watched{}).
		Where("id = ? AND user_id = ?", id, userId).
		Take(&upwat).
		Error
	if err != nil {
		slog.Error("UpdateWatched: Failed to get existing record.",
			"id", id, "error", err)
		return domain.WatchedUpdateResponse{},
			errors.New("failed to get watched entry")
	}

	// If nothing changes, we won't save to db.
	anythingChanged := false
	// Activities are built and added to this list, then if saving the Watched
	// changes succeeds, these activities are then saved too.
	activityMC := activity.NewMultiCreator()

	if ar.RemoveRating && upwat.Rating != 0 {
		anythingChanged = true
		upwat.Rating = 0

		activity.
			NewCreator(s.db, userId, id, entity.RATING_CHANGED, false, ar.ActivityCreatedBy).
			SetData("0").
			AddToMultiCreator(activityMC)
	} else if ar.Rating != 0 && ar.Rating != upwat.Rating {
		anythingChanged = true
		upwat.Rating = ar.Rating

		activity.
			NewCreator(s.db, userId, id, entity.RATING_CHANGED, false, ar.ActivityCreatedBy).
			SetData(strconv.FormatFloat(ar.Rating, 'f', -1, 64)).
			AddToMultiCreator(activityMC)
	}
	if ar.Status != "" && ar.Status != upwat.Status {
		anythingChanged = true
		upwat.Status = ar.Status

		countAsPlay := false
		if ar.Status == entity.FINISHED &&
			util.Deref(ar.LetCountAsPlay, true) != false {
			countAsPlay = true
		}
		activity.
			NewCreator(s.db, userId, id, entity.STATUS_CHANGED, countAsPlay, ar.ActivityCreatedBy).
			SetData(string(ar.Status)).
			AddToMultiCreator(activityMC)

		if ar.Status == entity.FINISHED && upwat.ContentID != nil {
			var cnt entity.Content
			if err := s.db.Where("id = ?", *upwat.ContentID).First(&cnt).Error; err == nil && cnt.Type == entity.SHOW {
				s.db.Model(&entity.WatchedSeason{}).Where("watched_id = ? AND user_id = ?", id, userId).Update("status", entity.FINISHED)
				s.db.Model(&entity.WatchedEpisode{}).Where("watched_id = ? AND user_id = ?", id, userId).Update("status", entity.FINISHED)
			}
		}
	}
	if ar.Thoughts != "" && ar.Thoughts != upwat.Thoughts {
		anythingChanged = true
		upwat.Thoughts = ar.Thoughts

		activity.
			NewCreator(s.db, userId, id, entity.THOUGHTS_CHANGED, false, ar.ActivityCreatedBy).
			AddToMultiCreator(activityMC)
	}
	if ar.RemoveThoughts {
		anythingChanged = true
		originalThoughts := upwat.Thoughts
		upwat.Thoughts = ""

		activity.
			NewCreator(s.db, userId, id, entity.THOUGHTS_REMOVED, false, ar.ActivityCreatedBy).
			SetData(originalThoughts).
			AddToMultiCreator(activityMC)
	}
	if ar.Pinned != nil {
		anythingChanged = true
		upwat.Pinned = *ar.Pinned
	}
	if ar.RemoveHoursPlayed && upwat.HoursPlayed != nil {
		anythingChanged = true
		upwat.HoursPlayed = nil
	} else if ar.HoursPlayed != nil && (upwat.HoursPlayed == nil || *ar.HoursPlayed != *upwat.HoursPlayed) {
		anythingChanged = true
		upwat.HoursPlayed = ar.HoursPlayed
	}
	if ar.Platform != nil && *ar.Platform != upwat.Platform {
		anythingChanged = true
		upwat.Platform = *ar.Platform
	}
	if ar.GameCompletionStatus != nil && *ar.GameCompletionStatus != upwat.GameCompletionStatus {
		anythingChanged = true
		upwat.GameCompletionStatus = *ar.GameCompletionStatus
	}

	if !anythingChanged {
		slog.Info("UpdateWatched: Nothing changed.. skipping save.")
		// NOTE: This is never an error (because idempotency but also it would
		// likely break any existing callers around project, if a caller needs
		// to know changed vs unchanged, make a bool in the Response struct).
		return domain.WatchedUpdateResponse{}, nil
	}

	res := s.db.Save(upwat)
	if res.Error != nil {
		slog.Error("UpdateWatched: Update query failed!", "error", err)
		return domain.WatchedUpdateResponse{}, errors.New("update failed")
	} else if res.RowsAffected <= 0 {
		slog.Error("UpdateWatched: Update query failed! No rows affected.")
		return domain.WatchedUpdateResponse{}, errors.New("no watched entry found")
	}

	// Trigger real-time activity webhook
	go func() {
		title := "Conteúdo"
		poster := ""
		mType := "content"
		if upwat.Content != nil {
			title = upwat.Content.Title
			poster = integration.BuildPosterURL(upwat.Content.PosterPath)
			mType = string(upwat.Content.Type)
		} else if upwat.Game != nil {
			title = upwat.Game.Name
			mType = "game"
		}
		evt := "media.status_changed"
		if ar.Status == entity.FINISHED {
			evt = "media.finished"
		} else if ar.Status == entity.WATCHING {
			evt = "media.watching"
		} else if ar.Rating > 0 && ar.Status == "" {
			evt = "media.rated"
		}
		integration.DispatchActivityWebhook(s.db, userId, integration.ActivityWebhookPayload{
			Event:     evt,
			Status:    string(upwat.Status),
			MediaType: mType,
			Title:     title,
			Rating:    upwat.Rating,
			Poster:    poster,
			Reason:    strconv.Itoa(int(ar.ActivityCreatedBy)),
		})
	}()

	newActivities := activityMC.CreateAll()

	return domain.WatchedUpdateResponse{NewActivities: newActivities}, nil
}

func (s *Service) UpdateWatchedLastViewedSeason(
	userId uint,
	id uint,
	seasonNum int,
) error {
	slog.Debug("UpdateWatchedLastViewedSeason",
		"user_id", userId,
		"id", id,
		"season_num", seasonNum)
	res := s.db.
		Model(&entity.Watched{}).
		Where("id = ? AND user_id = ?", id, userId).
		Update("last_viewed_season", seasonNum)
	if res.Error != nil {
		slog.Error("updateWatchedLastViewedSeason: Failed when updating.",
			"error", res.Error)
		return errors.New("failed to update db")
	}
	if res.RowsAffected == 0 {
		// likely the watched entry does not exist or is not owned by this `userId`.
		slog.Error("updateWatchedLastViewedSeason: Watched entry does not exist.")
		return errors.New("watched entry does not exist")
	}
	return nil
}

func (s *Service) removeWatched(
	userId uint,
	id uint,
) (domain.WatchedRemoveResponse, error) {
	slog.Debug("Removing watched item:", "id", id, "user_id", userId)
	// Our model has a deleted_at field, which will make gorm do a soft delete.
	// Since other tables (eg activities) will link their rows to a watched_id, it's best to soft
	// delete, so if user restores watched item they still have activity for example (also so
	// someone else wont get other users activity if auto increment gives them the same watched id).
	res := s.db.
		Model(&entity.Watched{}).
		Where("id = ? AND user_id = ?", id, userId).
		Delete(&entity.Watched{})
	if res.Error != nil {
		slog.Error("Removing watched entry failed",
			"id", id,
			"error", res.Error.Error())
		return domain.WatchedRemoveResponse{}, errors.New("failed to remove watched entry")
	}
	if res.RowsAffected <= 0 {
		return domain.WatchedRemoveResponse{}, errors.New("no watched entry found")
	}
	addedActivity, _ := activity.
		NewCreator(s.db, userId, id, entity.REMOVED_WATCHED, false, 0).
		Create()
	return domain.WatchedRemoveResponse{NewActivity: addedActivity}, nil
}

func (s *Service) bulkRemoveWatched(
	userId uint,
	req domain.WatchedBulkRemoveRequest,
) (domain.WatchedBulkRemoveResponse, error) {
	slog.Info("bulkRemoveWatched:", "userId", userId, "all", req.All, "hard", req.Hard, "idsCount", len(req.IDs))

	var deletedCount int64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if req.Hard {
			if req.All {
				res := tx.Unscoped().Where("user_id = ?", userId).Delete(&entity.Watched{})
				if res.Error != nil {
					return res.Error
				}
				deletedCount = res.RowsAffected
				if err := tx.Unscoped().Where("user_id = ?", userId).Delete(&entity.WatchedSeason{}).Error; err != nil {
					return err
				}
				if err := tx.Unscoped().Where("user_id = ?", userId).Delete(&entity.WatchedEpisode{}).Error; err != nil {
					return err
				}
				if err := tx.Unscoped().Where("user_id = ?", userId).Delete(&entity.Activity{}).Error; err != nil {
					return err
				}
			} else if len(req.IDs) > 0 {
				res := tx.Unscoped().Where("id IN ? AND user_id = ?", req.IDs, userId).Delete(&entity.Watched{})
				if res.Error != nil {
					return res.Error
				}
				deletedCount = res.RowsAffected
				if err := tx.Unscoped().Where("watched_id IN ? AND user_id = ?", req.IDs, userId).Delete(&entity.WatchedSeason{}).Error; err != nil {
					return err
				}
				if err := tx.Unscoped().Where("watched_id IN ? AND user_id = ?", req.IDs, userId).Delete(&entity.WatchedEpisode{}).Error; err != nil {
					return err
				}
				if err := tx.Unscoped().Where("watched_id IN ? AND user_id = ?", req.IDs, userId).Delete(&entity.Activity{}).Error; err != nil {
					return err
				}
			}
		} else {
			if req.All {
				res := tx.Model(&entity.Watched{}).Where("user_id = ?", userId).Delete(&entity.Watched{})
				if res.Error != nil {
					return res.Error
				}
				deletedCount = res.RowsAffected
			} else if len(req.IDs) > 0 {
				res := tx.Model(&entity.Watched{}).Where("id IN ? AND user_id = ?", req.IDs, userId).Delete(&entity.Watched{})
				if res.Error != nil {
					return res.Error
				}
				deletedCount = res.RowsAffected
			}
		}
		return nil
	})
	if err != nil {
		return domain.WatchedBulkRemoveResponse{}, err
	}

	// Reclaim memory back to OS in background after batch operation
	go debug.FreeOSMemory()

	return domain.WatchedBulkRemoveResponse{Count: deletedCount}, nil
}

func (s *Service) bulkUpdateWatched(
	userId uint,
	req domain.WatchedBulkUpdateRequest,
) (domain.WatchedBulkUpdateResponse, error) {
	slog.Info("bulkUpdateWatched:", "userId", userId, "all", req.All, "idsCount", len(req.IDs), "status", req.Status, "rating", req.Rating)

	if !req.All && len(req.IDs) == 0 {
		return domain.WatchedBulkUpdateResponse{}, errors.New("no items selected for bulk update")
	}

	updates := make(map[string]any)
	if req.Status != "" {
		if !req.Status.IsValid() {
			return domain.WatchedBulkUpdateResponse{}, errors.New("invalid status provided")
		}
		updates["status"] = req.Status
	}
	if req.Rating != nil {
		if *req.Rating < 0 || *req.Rating > 10 {
			return domain.WatchedBulkUpdateResponse{}, errors.New("rating must be between 0 and 10")
		}
		updates["rating"] = *req.Rating
	}
	if len(updates) == 0 {
		return domain.WatchedBulkUpdateResponse{}, errors.New("nothing to update")
	}

	var updatedCount int64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		q := tx.Model(&entity.Watched{}).Where("user_id = ?", userId)
		if !req.All {
			q = q.Where("id IN ?", req.IDs)
		}

		// Find target IDs to record activity entries in bulk
		var targetIDs []uint
		if err := q.Pluck("id", &targetIDs).Error; err != nil {
			return err
		}
		if len(targetIDs) == 0 {
			return nil
		}

		res := tx.Model(&entity.Watched{}).Where("id IN ? AND user_id = ?", targetIDs, userId).Updates(updates)
		if res.Error != nil {
			return res.Error
		}
		updatedCount = res.RowsAffected

		var activities []entity.Activity
		if req.Status != "" {
			countAsPlay := req.Status == entity.FINISHED
			statusStr := string(req.Status)
			for _, wid := range targetIDs {
				activities = append(activities, entity.Activity{
					UserID:      userId,
					WatchedID:   wid,
					Type:        entity.STATUS_CHANGED,
					Data:        statusStr,
					CountAsPlay: countAsPlay,
				})
			}
		}
		if req.Rating != nil {
			ratingStr := strconv.FormatFloat(*req.Rating, 'f', -1, 64)
			for _, wid := range targetIDs {
				activities = append(activities, entity.Activity{
					UserID:    userId,
					WatchedID: wid,
					Type:      entity.RATING_CHANGED,
					Data:      ratingStr,
				})
			}
		}
		if len(activities) > 0 {
			if err := tx.CreateInBatches(&activities, 200).Error; err != nil {
				slog.Warn("bulkUpdateWatched: Failed to insert batch activities", "error", err)
			}
		}
		return nil
	})
	if err != nil {
		return domain.WatchedBulkUpdateResponse{}, err
	}

	return domain.WatchedBulkUpdateResponse{Count: updatedCount}, nil
}

// Get all currently watching items across the entire collection for the user.
func (s *Service) GetContinueWatching(userId uint) ([]entity.Watched, error) {
	var watched []entity.Watched
	res := s.db.
		Model(&entity.Watched{}).
		Where(&entity.Watched{UserID: userId, Status: entity.WATCHING}).
		Joins("Content").
		Joins("Game").
		Preload("Game.Poster").
		Preload("Tags").
		Preload("WatchedSeasons").
		Preload("WatchedEpisodes").
		Order("watcheds.updated_at DESC").
		Limit(30).
		Find(&watched)
	if res.Error != nil {
		slog.Error("GetContinueWatching: Failed!", "error", res.Error)
		return nil, res.Error
	}
	return watched, nil
}

type seasonCacheEntry struct {
	details   tmdb.SeasonDetails
	expiresAt time.Time
}

var (
	seasonCache    sync.Map
	seasonCacheTTL = 2 * time.Hour
)

// getSeasonDetailsCached retrieves TMDB season details with an in-memory TTL cache to eliminate redundant HTTP calls
func (s *Service) getSeasonDetailsCached(showId string, seasonNum string) (tmdb.SeasonDetails, error) {
	if s.tmdb == nil {
		return tmdb.SeasonDetails{}, errors.New("tmdb provider not available")
	}

	key := showId + ":" + seasonNum
	if val, ok := seasonCache.Load(key); ok {
		entry := val.(seasonCacheEntry)
		if time.Now().Before(entry.expiresAt) {
			return entry.details, nil
		}
		seasonCache.Delete(key)
	}

	sd, err := s.tmdb.SeasonDetails(showId, seasonNum)
	if err != nil {
		return tmdb.SeasonDetails{}, err
	}

	seasonCache.Store(key, seasonCacheEntry{
		details:   sd,
		expiresAt: time.Now().Add(seasonCacheTTL),
	})
	return sd, nil
}

// GetUpNext computes the exact next episode to watch for all active TV shows using parallel worker pool and caching
func (s *Service) GetUpNext(userId uint) ([]domain.UpNextItem, error) {
	var watched []entity.Watched
	res := s.db.
		Model(&entity.Watched{}).
		Where(&entity.Watched{UserID: userId, Status: entity.WATCHING}).
		Joins("Content").
		Preload("WatchedSeasons").
		Preload("WatchedEpisodes").
		Order("watcheds.updated_at DESC").
		Limit(30).
		Find(&watched)
	if res.Error != nil {
		slog.Error("GetUpNext: Query failed!", "error", res.Error)
		return nil, res.Error
	}

	type showTask struct {
		index int
		w     entity.Watched
	}

	var validShows []showTask
	for i := range watched {
		w := &watched[i]
		if w.Content != nil && w.Content.Type == entity.SHOW {
			validShows = append(validShows, showTask{index: len(validShows), w: watched[i]})
		}
	}

	if len(validShows) == 0 {
		return []domain.UpNextItem{}, nil
	}

	results := make([]*domain.UpNextItem, len(validShows))
	var wg sync.WaitGroup

	// Bounded concurrency (up to 6 parallel workers) to avoid TMDB rate limits while cutting latency from seconds to <300ms
	workers := 6
	if len(validShows) < workers {
		workers = len(validShows)
	}
	sem := make(chan struct{}, workers)

	for i := range validShows {
		wg.Add(1)
		go func(task showTask) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			w := &task.w
			maxFinishedSeason := 0
			maxFinishedEpisode := 0

			for _, ep := range w.WatchedEpisodes {
				if ep.Status == entity.FINISHED {
					if ep.SeasonNumber > maxFinishedSeason || (ep.SeasonNumber == maxFinishedSeason && ep.EpisodeNumber > maxFinishedEpisode) {
						maxFinishedSeason = ep.SeasonNumber
						maxFinishedEpisode = ep.EpisodeNumber
					}
				}
			}

			for _, ws := range w.WatchedSeasons {
				if ws.Status == entity.FINISHED {
					if ws.SeasonNumber > maxFinishedSeason {
						maxFinishedSeason = ws.SeasonNumber
						maxFinishedEpisode = 0
					}
				}
			}

			candidateSeason := 1
			candidateEpisode := 1

			if maxFinishedSeason > 0 {
				if maxFinishedEpisode == 0 {
					candidateSeason = maxFinishedSeason + 1
					candidateEpisode = 1
				} else {
					candidateSeason = maxFinishedSeason
					candidateEpisode = maxFinishedEpisode + 1
				}
			}

			item := domain.UpNextItem{
				WatchedID:     w.ID,
				ShowID:        *w.ContentID,
				ShowTMDBID:    w.Content.TmdbID,
				ShowTitle:     w.Content.Title,
				ShowPoster:    w.Content.PosterPath,
				ShowBackdrop:  w.Content.BackdropPath,
				SeasonNumber:  candidateSeason,
				EpisodeNumber: candidateEpisode,
				EpisodeName:   fmt.Sprintf("Episódio %d", candidateEpisode),
				UpdatedAt:     w.UpdatedAt,
			}

			if s.tmdb != nil && w.Content.TmdbID > 0 {
				sd, err := s.getSeasonDetailsCached(strconv.Itoa(w.Content.TmdbID), strconv.Itoa(candidateSeason))
				if err == nil && len(sd.Episodes) > 0 {
					if candidateEpisode <= len(sd.Episodes) {
						ep := sd.Episodes[candidateEpisode-1]
						item.EpisodeName = ep.Name
						item.EpisodeOverview = ep.Overview
						item.EpisodeStill = ep.StillPath
						item.AirDate = ep.AirDate
						item.TotalEpisodesInSeason = len(sd.Episodes)
						item.EpisodesRemaining = len(sd.Episodes) - candidateEpisode + 1
						if ep.AirDate != "" {
							if t, err := time.Parse("2006-01-02", ep.AirDate); err == nil {
								item.IsAired = time.Now().After(t)
							}
						}
						results[task.index] = &item
						return
					} else {
						// Advance to next season
						nextSd, err := s.getSeasonDetailsCached(strconv.Itoa(w.Content.TmdbID), strconv.Itoa(candidateSeason+1))
						if err == nil && len(nextSd.Episodes) > 0 {
							ep := nextSd.Episodes[0]
							item.SeasonNumber = candidateSeason + 1
							item.EpisodeNumber = 1
							item.EpisodeName = ep.Name
							item.EpisodeOverview = ep.Overview
							item.EpisodeStill = ep.StillPath
							item.AirDate = ep.AirDate
							item.TotalEpisodesInSeason = len(nextSd.Episodes)
							item.EpisodesRemaining = len(nextSd.Episodes)
							if ep.AirDate != "" {
								if t, err := time.Parse("2006-01-02", ep.AirDate); err == nil {
									item.IsAired = time.Now().After(t)
								}
							}
							results[task.index] = &item
							return
						}
					}
				}
			}

			results[task.index] = &item
		}(validShows[i])
	}
	wg.Wait()

	upNext := make([]domain.UpNextItem, 0, len(results))
	for _, it := range results {
		if it != nil {
			upNext = append(upNext, *it)
		}
	}

	return upNext, nil
}

// PickRoulette selects random items from user's watchlist/library matching specific mood/runtime criteria
func (s *Service) PickRoulette(userId uint, req domain.RoulettePickRequest) (domain.RoulettePickResponse, error) {
	query := s.db.Model(&entity.Watched{}).Where("watcheds.user_id = ?", userId)

	status := strings.ToLower(strings.TrimSpace(req.Status))
	switch status {
	case "planned":
		query = query.Where("watcheds.status = ?", entity.PLANNED)
	case "watching":
		query = query.Where("watcheds.status = ?", entity.WATCHING)
	case "finished":
		query = query.Where("watcheds.status = ?", entity.FINISHED)
	default:
		if status == "" {
			query = query.Where("watcheds.status = ?", entity.PLANNED)
		}
	}

	var allWatched []entity.Watched
	err := query.
		Joins("Content").
		Joins("Game").
		Preload("Game.Poster").
		Preload("Tags").
		Find(&allWatched).Error
	if err != nil {
		slog.Error("PickRoulette: Failed to fetch candidate pool", "error", err)
		return domain.RoulettePickResponse{}, err
	}

	reqType := strings.ToLower(strings.TrimSpace(req.Type))
	if reqType == "" {
		reqType = "movie"
	}

	genreFilter := strings.ToLower(strings.TrimSpace(req.Genre))

	filtered := make([]entity.Watched, 0, len(allWatched))
	for _, w := range allWatched {
		// 1. Type filter
		if reqType == "movie" {
			if w.Content == nil || w.Content.Type != entity.MOVIE {
				continue
			}
		} else if reqType == "tv" {
			if w.Content == nil || w.Content.Type != entity.SHOW {
				continue
			}
		} else if reqType == "game" {
			if w.Game == nil {
				continue
			}
		}

		// 2. MaxRuntime filter
		if req.MaxRuntime > 0 {
			if w.Content != nil {
				if w.Content.Runtime > 0 && int(w.Content.Runtime) > req.MaxRuntime {
					continue
				}
			}
		}

		// 3. MinRating filter
		if req.MinRating > 0 {
			var score float64
			if w.Rating > 0 {
				score = w.Rating
			} else if w.Content != nil {
				score = float64(w.Content.VoteAverage)
			} else if w.Game != nil {
				score = w.Game.Rating / 10.0
			}
			if score < req.MinRating {
				continue
			}
		}

		// 4. Genre filter
		if genreFilter != "" {
			matched := false
			genreAliases := []string{genreFilter}
			switch genreFilter {
			case "action", "ação", "acao":
				genreAliases = []string{"action", "ação", "acao"}
			case "comedy", "comédia", "comedia":
				genreAliases = []string{"comedy", "comédia", "comedia"}
			case "science fiction", "sci-fi", "ficção científica", "ficcao cientifica":
				genreAliases = []string{"science fiction", "sci-fi", "ficção científica", "ficcao cientifica"}
			case "horror", "terror":
				genreAliases = []string{"horror", "terror"}
			case "animation", "animação", "animacao":
				genreAliases = []string{"animation", "animação", "animacao"}
			case "thriller", "suspense":
				genreAliases = []string{"thriller", "suspense"}
			case "fantasy", "fantasia":
				genreAliases = []string{"fantasy", "fantasia"}
			case "adventure", "aventura":
				genreAliases = []string{"adventure", "aventura"}
			}

			for _, alias := range genreAliases {
				if w.Content != nil && strings.Contains(strings.ToLower(w.Content.Genres), alias) {
					matched = true
					break
				}
				if w.Game != nil && strings.Contains(strings.ToLower(w.Game.Genres), alias) {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		filtered = append(filtered, w)
	}

	totalPool := len(filtered)
	if totalPool == 0 {
		return domain.RoulettePickResponse{
			Items:     []domain.Media{},
			TotalPool: 0,
		}, nil
	}

	count := req.Count
	if count <= 0 {
		count = 1
	}
	if count > 10 {
		count = 10
	}
	if count > totalPool {
		count = totalPool
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(filtered), func(i, j int) {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	})

	picked := filtered[:count]
	items := make([]domain.Media, 0, len(picked))
	for i := range picked {
		v := &picked[i]
		dto := domain.NewWatchedDtoForLists(v)
		items = append(items, domain.NewMediaFromWatched(v, &dto))
	}

	return domain.RoulettePickResponse{
		Items:     items,
		TotalPool: totalPool,
	}, nil
}
