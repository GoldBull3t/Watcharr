package domain

type CalendarFilter string

const (
	CalendarFilterLibrary CalendarFilter = "library"
	CalendarFilterAll     CalendarFilter = "all"
)

type CalendarType string

const (
	CalendarTypeAll   CalendarType = "all"
	CalendarTypeMovie CalendarType = "movie"
	CalendarTypeTV    CalendarType = "tv"
)

type CalendarRequest struct {
	Start  string         `form:"start"`  // "YYYY-MM-DD"
	End    string         `form:"end"`    // "YYYY-MM-DD"
	Filter CalendarFilter `form:"filter"` // "library" | "all"
	Type   CalendarType   `form:"type"`   // "all" | "movie" | "tv"
	Region string         `form:"region"` // e.g. "US"
}

type CalendarEvent struct {
	ID            string `json:"id"`
	MediaID       int    `json:"mediaId"`
	MediaType     string `json:"mediaType"` // "movie" | "tv"
	Title         string `json:"title"`
	ReleaseDate   string `json:"releaseDate"` // "YYYY-MM-DD"
	PosterPath    string `json:"posterPath,omitempty"`
	BackdropPath  string `json:"backdropPath,omitempty"`
	Overview      string `json:"overview,omitempty"`
	InLibrary     bool   `json:"inLibrary"`
	Status        string `json:"status,omitempty"` // "WATCHING", "PLANNED", etc.
	SeasonNumber  int    `json:"seasonNumber,omitempty"`
	EpisodeNumber int    `json:"episodeNumber,omitempty"`
	EpisodeName   string `json:"episodeName,omitempty"`
}

