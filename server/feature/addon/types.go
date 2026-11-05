package addon

// Stremio / Nuvio Manifest Specification DTOs

type ManifestResponse struct {
	ID            string              `json:"id"`
	Name          string              `json:"name"`
	Version       string              `json:"version"`
	Description   string              `json:"description,omitempty"`
	Logo          string              `json:"logo,omitempty"`
	Background    string              `json:"background,omitempty"`
	Resources     []string            `json:"resources"`
	Types         []string            `json:"types"`
	Catalogs      []CatalogDescriptor `json:"catalogs"`
	IDPrefixes    []string            `json:"idPrefixes,omitempty"`
	BehaviorHints *BehaviorHints      `json:"behaviorHints,omitempty"`
}

type BehaviorHints struct {
	Configurable          bool `json:"configurable,omitempty"`
	ConfigurationRequired bool `json:"configurationRequired,omitempty"`
}

type CatalogDescriptor struct {
	Type           string   `json:"type"`
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Extra          []Extra  `json:"extra,omitempty"`
	ExtraSupported []string `json:"extraSupported,omitempty"`
	Genres         []string `json:"genres,omitempty"`
}

type Extra struct {
	Name       string   `json:"name"`
	IsRequired bool     `json:"isRequired,omitempty"`
	Options    []string `json:"options,omitempty"`
}

type CatalogResponse struct {
	Metas []MetaPreview `json:"metas"`
}

type MetaPreview struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Name        string   `json:"name"`
	Poster      string   `json:"poster,omitempty"`
	PosterShape string   `json:"posterShape,omitempty"`
	Background  string   `json:"background,omitempty"`
	Logo        string   `json:"logo,omitempty"`
	Description string   `json:"description,omitempty"`
	ReleaseInfo string   `json:"releaseInfo,omitempty"`
	ImdbRating  string   `json:"imdbRating,omitempty"`
	Genres      []string `json:"genres,omitempty"`
}

type MetaDetailResponse struct {
	Meta MetaDetail `json:"meta"`
}

type MetaDetail struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Name        string   `json:"name"`
	Genres      []string `json:"genres,omitempty"`
	Poster      string   `json:"poster,omitempty"`
	Background  string   `json:"background,omitempty"`
	Logo        string   `json:"logo,omitempty"`
	Description string   `json:"description,omitempty"`
	ReleaseInfo string   `json:"releaseInfo,omitempty"`
	ImdbRating  string   `json:"imdbRating,omitempty"`
	Videos      []Video  `json:"videos,omitempty"`
}

type Video struct {
	ID       string `json:"id"`
	Title    string `json:"title,omitempty"`
	Season   int    `json:"season"`
	Episode  int    `json:"number"`
	Released string `json:"released,omitempty"`
}

// Stremio / Nuvio Stream response for active scrobble interception
type StreamResponse struct {
	Streams []Stream `json:"streams"`
}

type Stream struct {
	Title string `json:"title,omitempty"`
	URL   string `json:"url,omitempty"`
}

// Scrobble / Playback Payload from Nuvio / Player
type ScrobblePayload struct {
	Event         string  `json:"event"`         // "start", "pause", "stop"
	MediaType     string  `json:"mediaType"`     // "movie" or "series" / "show"
	TmdbID        int     `json:"tmdbId"`
	ImdbID        string  `json:"imdbId"`
	SeasonNumber  int     `json:"seasonNumber"`  // For TV
	EpisodeNumber int     `json:"episodeNumber"` // For TV
	Progress      float64 `json:"progress"`      // 0 - 100 percentage
}
