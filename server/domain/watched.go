package domain

import (
	"errors"
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/feature/watched/watchedutil"
	"github.com/sbondCo/Watcharr/util"
)

var (
	// Watched entry already exists.
	ErrWatchedExists = errors.New("watched entry exists")
	// Watched entry exists as a soft deleted record (..and restore wasn't attempted by choice).
	ErrWatchedExistsSoftDeleted = errors.New("watched entry exists soft deleted")
)

type WatchedSort string

const (
	WatchedSortDateAdded    WatchedSort = "DATEADDED"
	WatchedSortLastChanged  WatchedSort = "LASTCHANGED"
	WatchedSortLastFinished WatchedSort = "LASTFIN"
	WatchedSortRating       WatchedSort = "RATING"
	WatchedSortAlphabetical WatchedSort = "ALPHA"
	WatchedSortDateReleased WatchedSort = "DATERELEASED"
)

type SortDirection string

const (
	WatchedSortDirAsc  SortDirection = "asc"
	WatchedSortDirDesc SortDirection = "desc"
)

// Get watched page request extra (GET) options.
// Since this is user input, validity of string types cannot be guaranteed.
type WatchedGetPageRequest struct {
	// Sorting type.
	Sort WatchedSort `form:"sort"`
	// Sorting direction (asc or desc).
	SortDir SortDirection `form:"sortDir,default=desc"`
	// Filtering options.
	FilterType   []util.SupportedMedia  `form:"type" collection_format:"csv"`
	FilterStatus []entity.WatchedStatus `form:"status" collection_format:"csv"`
	Year         int                    `form:"year"`
	MinRating    float64                `form:"minRating"`
	MaxRating    float64                `form:"maxRating"`
	Genre            string                 `form:"genre"`
	Query            string                 `form:"query"`
	ProductionStatus string                 `form:"productionStatus"`
}

type WatchedGetPageExtraProps struct {
	// Only get these watched ids.
	WatchedIds []int
	// Only get watched items where content matches this query.
	Query string
}

type WatchedDto struct {
	// Properties that always exist in every watched dto below.

	ID        uint                 `json:"id"`
	CreatedAt time.Time            `json:"createdAt"`
	UpdatedAt time.Time            `json:"updatedAt"`
	Status    entity.WatchedStatus `json:"status"`
	Rating    float64              `json:"rating"`
	Pinned    bool                 `json:"pinned"`

	// Properties that may not be included in all watched dtos
	// (depending on where we are making the dto for)

	// Thoughts aren't always included (not needed on watched list pages
	// & on public pages because they could be private).
	Thoughts string `json:"thoughts,omitempty"`
	// Watching Season extra detail for list.
	WatchingSeason   string                  `json:"watchingSeason,omitempty"`
	Activity         []entity.Activity       `json:"activity,omitempty"`
	WatchedSeasons   []entity.WatchedSeason  `json:"watchedSeasons,omitempty"`
	WatchedEpisodes  []entity.WatchedEpisode `json:"watchedEpisodes,omitempty"`
	Tags             []entity.Tag            `json:"tags,omitempty"`
	LastViewedSeason *int                    `json:"lastViewedSeason,omitempty"`
	// Amount of plays this media has, calculated from activity.
	Plays int `json:"plays,omitempty"`

	// Gaming 2.0 tracker fields
	HoursPlayed          *float64 `json:"hoursPlayed,omitempty"`
	Platform             string   `json:"platform,omitempty"`
	GameCompletionStatus string   `json:"gameCompletionStatus,omitempty"`
}

// New dto with base properties that we have for all WatchedDtos.
// Note: If this is updated, ensure whatever uses this still makes sense.
func NewWatchedDtoWithBaseProps(w *entity.Watched) WatchedDto {
	return WatchedDto{
		ID:        w.ID,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
		Status:    w.Status,
		Rating:    w.Rating,
		Pinned:               w.Pinned,
		HoursPlayed:          w.HoursPlayed,
		Platform:             w.Platform,
		GameCompletionStatus: w.GameCompletionStatus,
	}
}

func NewWatchedDtoForLists(w *entity.Watched) WatchedDto {
	dto := NewWatchedDtoWithBaseProps(w)

	if w.Content != nil && w.Content.Type == entity.SHOW {
		dto.WatchingSeason = watchedutil.GetLatestWatchedInTv(
			w.WatchedSeasons, w.WatchedEpisodes)
	}

	return dto
}

// For public lists showing other users watched data.
func NewWatchedDtoForPublicLists(w *entity.Watched) WatchedDto {
	dto := NewWatchedDtoWithBaseProps(w)

	if w.Content != nil && w.Content.Type == entity.SHOW {
		dto.WatchingSeason = watchedutil.GetLatestWatchedInTv(
			w.WatchedSeasons, w.WatchedEpisodes)
	}

	return dto
}

// A fuller dto with all details needed for a content details page.
func NewWatchedDtoForContentPage(w *entity.Watched) WatchedDto {
	dto := NewWatchedDtoWithBaseProps(w)

	dto.Thoughts = w.Thoughts
	dto.Activity = w.Activity
	dto.WatchedSeasons = w.WatchedSeasons
	dto.WatchedEpisodes = w.WatchedEpisodes
	dto.Tags = w.Tags
	dto.LastViewedSeason = w.LastViewedSeason
	dto.Plays = getPlaysFromActivity(w.Activity)

	if w.Content != nil && w.Content.Type == entity.SHOW {
		dto.WatchingSeason = watchedutil.GetLatestWatchedInTv(
			w.WatchedSeasons, w.WatchedEpisodes)
	}

	return dto
}

// Get our watched page response.
type WatchedGetPageResponse []Media

// Used for GetWatchedPage (and GetWatchedPage for our search).
func NewWatchedGetPageResponse(w []entity.Watched) WatchedGetPageResponse {
	r := WatchedGetPageResponse{}
	for i := range w {
		v := &w[i]
		d := NewWatchedDtoForLists(v)
		r = append(r, NewMediaFromWatched(v, &d))
	}
	return r
}

// Get a public users list response.
type WatchedPublicGetPageResponse []Media

func NewWatchedPublicGetPageResponse(w []entity.Watched) WatchedPublicGetPageResponse {
	r := WatchedPublicGetPageResponse{}
	for i := range w {
		v := &w[i]
		d := NewWatchedDtoForPublicLists(v)
		r = append(r, NewMediaFromWatched(v, &d))
	}
	return r
}

// Add a watched entry request
type WatchedAddRequest struct {
	// Type of content we are adding to watched.
	ContentType util.SupportedMedia `json:"contentType" binding:"required,oneof=movie tv game"`
	// ID of content from tmdb (if ContentType is movie or tv).
	TMDBID int `json:"tmdbId"`
	// DEPRECATED!! This will be removed soon, I've left it in only so any third
	// party scripts can migrate pain-free. This will work as if you have passed
	// the id for `tmdbId`, please replace 'contentId' properties in requests
	// with `tmdbId`.
	Deprecated_ContentID int `json:"contentId"`
	// ID of content from igdb (if ContentType is game).
	IGDBID int `json:"igdbId"`

	Status   entity.WatchedStatus `json:"status"`
	Rating   float64              `json:"rating" binding:"max=10"`
	Thoughts string               `json:"thoughts"`
	// Gaming 2.0 tracker fields
	HoursPlayed          *float64 `json:"hoursPlayed,omitempty"`
	Platform             string   `json:"platform,omitempty"`
	GameCompletionStatus string   `json:"gameCompletionStatus,omitempty"`
	// Pass a watched date and we will set the CreatedAt (and initial UpdatedAt)
	// properties for this watched entry to this specific date.
	WatchedDate time.Time `json:"watchedDate,omitempty"`
}

type WatchedAddExtraProps struct {
	// Set the CreatedBy value on the created activity.
	ActivityCreatedBy entity.ActivityCreatedBy
	// When watched entry to db, if a unique constraint is hit, should
	// we skip the restoration logic (bring back soft deleted record)?
	// - Syncing logic may use this field as it might not make sense to restore
	// a record when doing a task like a regular sync, otherwise the user could
	// soft delte a record and then have it come back when they next sync.
	// - Importers however probably shouldn't use this flag as true because
	// importing will always be a manual decision.
	DontRestore bool
}

// Update watched entry request
type WatchedUpdateRequest struct {
	Status         entity.WatchedStatus `json:"status"`
	Rating         float64              `json:"rating"`
	RemoveRating   bool                 `json:"removeRating"`
	Thoughts       string               `json:"thoughts" `
	RemoveThoughts bool                 `json:"removeThoughts"`
	Pinned         *bool                `json:"pinned" `
	// Allow the added activity count as play?
	// If the activity was going to count, this can stop it.
	LetCountAsPlay *bool `json:"letCountAsPlay"`

	// Gaming 2.0 tracker fields
	HoursPlayed          *float64 `json:"hoursPlayed,omitempty"`
	RemoveHoursPlayed    bool     `json:"removeHoursPlayed,omitempty"`
	Platform             *string  `json:"platform,omitempty"`
	GameCompletionStatus *string  `json:"gameCompletionStatus,omitempty"`

	// Set the CreatedBy value on the created activity.
	// Internal Only.
	ActivityCreatedBy entity.ActivityCreatedBy `json:"-"`
}

// If the struct is valid for the Update Request.
// - has atleast one property set.
// - If rating is set, can't be out of bounds.
func (w WatchedUpdateRequest) Valid() error {
	if w.Status == "" &&
		w.Rating == 0 &&
		!w.RemoveRating &&
		(w.Thoughts == "" && !w.RemoveThoughts) &&
		w.Pinned == nil &&
		w.LetCountAsPlay == nil &&
		w.HoursPlayed == nil &&
		!w.RemoveHoursPlayed &&
		w.Platform == nil &&
		w.GameCompletionStatus == nil {
		// No properties are set, so this struct is not valid.
		return errors.New("no properties provided")
	}
	if w.Status != "" && !w.Status.IsValid() {
		return errors.New("status is not set to a supported status")
	}
	if w.Rating < 0 || w.Rating > 10 {
		return errors.New("rating can only be a value from 0-10")
	}
	if w.HoursPlayed != nil && *w.HoursPlayed < 0 {
		return errors.New("hours played cannot be negative")
	}
	return nil
}

// Update response.
type WatchedUpdateResponse struct {
	NewActivities []entity.Activity `json:"newActivities"`
}

// Removal response.
type WatchedRemoveResponse struct {
	NewActivity entity.Activity `json:"newActivity"`
}

// Bulk removal request.
type WatchedBulkRemoveRequest struct {
	IDs  []uint `json:"ids"`
	All  bool   `json:"all"`
	Hard bool   `json:"hard"`
}

// Bulk removal response.
type WatchedBulkRemoveResponse struct {
	Count int64 `json:"count"`
}

// Bulk update request.
type WatchedBulkUpdateRequest struct {
	IDs    []uint               `json:"ids"`
	All    bool                 `json:"all"`
	Status entity.WatchedStatus `json:"status,omitempty"`
	Rating *float64             `json:"rating,omitempty"`
}

// Bulk update response.
type WatchedBulkUpdateResponse struct {
	Count int64 `json:"count"`
}

// UpNextItem represents the immediate next episode to watch for an active TV show
type UpNextItem struct {
	WatchedID             uint      `json:"watchedId"`
	ShowID                int       `json:"showId"`
	ShowTMDBID            int       `json:"showTmdbId"`
	ShowTitle             string    `json:"showTitle"`
	ShowPoster            string    `json:"showPoster"`
	ShowBackdrop          string    `json:"showBackdrop"`
	SeasonNumber          int       `json:"seasonNumber"`
	EpisodeNumber         int       `json:"episodeNumber"`
	EpisodeName           string    `json:"episodeName"`
	EpisodeOverview       string    `json:"episodeOverview"`
	EpisodeStill          string    `json:"episodeStill"`
	AirDate               string    `json:"airDate"`
	IsAired               bool      `json:"isAired"`
	TotalEpisodesInSeason int       `json:"totalEpisodesInSeason"`
	EpisodesRemaining     int       `json:"episodesRemaining"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

// RoulettePickRequest specifies filters for the smart roulette picker
type RoulettePickRequest struct {
	Type       string  `json:"type" form:"type"`
	Status     string  `json:"status" form:"status"`
	MaxRuntime int     `json:"maxRuntime" form:"maxRuntime"`
	Genre      string  `json:"genre" form:"genre"`
	MinRating  float64 `json:"minRating" form:"minRating"`
	Count      int     `json:"count" form:"count"`
}

// RoulettePickResponse contains the randomly picked items and the total eligible pool count
type RoulettePickResponse struct {
	Items     []Media `json:"items"`
	TotalPool int     `json:"totalPool"`
}
