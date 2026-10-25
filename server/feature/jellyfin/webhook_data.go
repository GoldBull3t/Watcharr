package jellyfin

import "log/slog"

// The webhook data format expected from Jellyfin.
type WebhookData struct {
	//
	// Server
	//

	NotificationType WebhookType
	ServerID         string `json:"ServerId"`

	// UserDataSaved events include a save reason.
	SaveReason WebhookUserDataSaveReason

	//
	// User
	//

	NotificationUsername string
	UserID               string `json:"UserId"`

	//
	// BaseItem
	//

	Name           string
	ItemType       WebhookItemType
	ItemID         string `json:"ItemId"`
	ProviderIMDBID string `json:"Provider_imdb"`
	ProviderTVDBID string `json:"Provider_tvdb"`
	ProviderTMDBID string `json:"Provider_tmdb"`

	//
	// Season
	//

	SeriesName string
	SeriesID   string `json:"SeriesId"`
	// Is a ptr so we can tell when its `unset` vs `season 0`
	SeasonNumber *int

	//
	// Episode
	//

	EpisodeNumber int
	// For multi episode files, this will have the final episode number (IndexNumberEnd).
	// Will be unset (0) when the episode file is just one episode, so we can
	// ignore this when it is zero.
	EpisodeNumberEnd int

	// Only for `NotificationType == PlaybackStop`.
	// Tells us if this Playback resulted in a completed play.
	PlayedToCompletion bool
	// Other NotificationTypes.
	// I believe this comes from the existing value of if the item is played in
	// jellyfins database, so isn't specific to the session (so always prefer
	// PlayedToCompletion when we have it in PlaybackStop event).
	Played bool
}

// Webhook notification types that we use/support.
// https://github.com/jellyfin/jellyfin-plugin-webhook/blob/master/Jellyfin.Plugin.Webhook/Destinations/NotificationType.cs
type WebhookType string

const (
	// When user starts playback on jellyfin.
	WebhookTypePlaybackStart WebhookType = "PlaybackStart"
	// User stops watching (finished media / quit half way through).
	WebhookTypePlaybackStop WebhookType = "PlaybackStop"
	// User data is updated (eg: Marks movie finished in webui)
	WebhookTypeUserDataSaved WebhookType = "UserDataSaved"
)

// Webhook ItemTypes that we support.
type WebhookItemType string

const (
	WebhookItemTypeMovie WebhookItemType = "Movie"
	// WebhookItemTypeSeries  WebhookItemType = "Series"
	// WebhookItemTypeSeason  WebhookItemType = "Season"
	WebhookItemTypeEpisode WebhookItemType = "Episode"
)

type WebhookUserDataSaveReason string

const (
	// User toggled 'Played' on the item in jellyfin web ui.
	WebhookUserDataSaveReasonTogglePlayed WebhookUserDataSaveReason = "TogglePlayed"
)

// Emby Webhook payload format (supported by Emby Premiere Webhooks and Emby Webhook Plugin)
type EmbyWebhookData struct {
	Event string `json:"Event"`
	User  struct {
		ID   string `json:"Id"`
		Name string `json:"Name"`
	} `json:"User"`
	Item struct {
		ID                string            `json:"Id"`
		Name              string            `json:"Name"`
		Type              string            `json:"Type"`
		SeriesName        string            `json:"SeriesName"`
		SeriesID          string            `json:"SeriesId"`
		IndexNumber       int               `json:"IndexNumber"`
		ParentIndexNumber int               `json:"ParentIndexNumber"`
		ProviderIds       map[string]string `json:"ProviderIds"`
		Played            bool              `json:"Played"`
	} `json:"Item"`
	PlaybackInfo struct {
		PositionTicks int64 `json:"PositionTicks"`
		TotalTicks    int64 `json:"TotalTicks"`
	} `json:"PlaybackInfo"`
}

// ToWebhookData converts an incoming Emby webhook payload into standard WebhookData.
// Returns (data, handled, ignored).
func (e *EmbyWebhookData) ToWebhookData() (*WebhookData, bool, bool) {
	if e.Event == "" {
		return nil, false, false
	}

	var notifType WebhookType
	var saveReason WebhookUserDataSaveReason
	var playedToCompletion bool
	played := e.Item.Played

	switch e.Event {
	case "playback.start":
		notifType = WebhookTypePlaybackStart
	case "playback.stop":
		notifType = WebhookTypePlaybackStop
		// Completed if item is marked as played or progress is >= 90%
		if e.Item.Played || (e.PlaybackInfo.TotalTicks > 0 && float64(e.PlaybackInfo.PositionTicks)/float64(e.PlaybackInfo.TotalTicks) >= 0.90) {
			playedToCompletion = true
			played = true
		}
	case "playback.pause", "playback.unpause":
		// Harmless playback state change
		return nil, false, true
	case "item.markplayed":
		notifType = WebhookTypeUserDataSaved
		saveReason = WebhookUserDataSaveReasonTogglePlayed
		played = true
	case "item.markunplayed":
		notifType = WebhookTypeUserDataSaved
		saveReason = WebhookUserDataSaveReasonTogglePlayed
		played = false
	case "item.rate":
		notifType = WebhookTypeUserDataSaved
	default:
		slog.Debug("ToWebhookData: Unhandled Emby event", "event", e.Event)
		return nil, false, true
	}

	var itemType WebhookItemType
	switch e.Item.Type {
	case "Movie":
		itemType = WebhookItemTypeMovie
	case "Episode":
		itemType = WebhookItemTypeEpisode
	default:
		return nil, false, true
	}

	tmdbID := ""
	imdbID := ""
	tvdbID := ""
	if e.Item.ProviderIds != nil {
		if v, ok := e.Item.ProviderIds["Tmdb"]; ok {
			tmdbID = v
		} else if v, ok := e.Item.ProviderIds["TMDB"]; ok {
			tmdbID = v
		}
		if v, ok := e.Item.ProviderIds["Imdb"]; ok {
			imdbID = v
		} else if v, ok := e.Item.ProviderIds["IMDB"]; ok {
			imdbID = v
		}
		if v, ok := e.Item.ProviderIds["Tvdb"]; ok {
			tvdbID = v
		} else if v, ok := e.Item.ProviderIds["TVDB"]; ok {
			tvdbID = v
		}
	}

	seasonNum := e.Item.ParentIndexNumber
	var seasonPtr *int
	if seasonNum > 0 || e.Item.Type == "Episode" {
		seasonPtr = &seasonNum
	}

	return &WebhookData{
		NotificationType:     notifType,
		SaveReason:           saveReason,
		NotificationUsername: e.User.Name,
		UserID:               e.User.ID,
		Name:                 e.Item.Name,
		ItemType:             itemType,
		ItemID:               e.Item.ID,
		ProviderTMDBID:       tmdbID,
		ProviderIMDBID:       imdbID,
		ProviderTVDBID:       tvdbID,
		SeriesName:           e.Item.SeriesName,
		SeriesID:             e.Item.SeriesID,
		SeasonNumber:         seasonPtr,
		EpisodeNumber:        e.Item.IndexNumber,
		PlayedToCompletion:   playedToCompletion,
		Played:               played,
	}, true, false
}
