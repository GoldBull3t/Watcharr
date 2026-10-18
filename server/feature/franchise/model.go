package franchise

import "github.com/sbondCo/Watcharr/domain"

type FranchiseItem struct {
	TmdbID        int                `json:"tmdbId"`
	Type          string             `json:"type"` // "movie" or "tv"
	Title         string             `json:"title"`
	Year          string             `json:"year"`
	PosterPath    string             `json:"posterPath,omitempty"`
	BackdropPath  string             `json:"backdropPath,omitempty"`
	Overview      string             `json:"overview,omitempty"`
	PhaseOrEra    string             `json:"phaseOrEra,omitempty"`
	TimelineOrder int                `json:"timelineOrder"`
	ReleaseOrder  int                `json:"releaseOrder"`
	ReleaseDate   string             `json:"releaseDate,omitempty"`
	Watched       *domain.WatchedDto `json:"watched,omitempty"`
	IsWatched     bool               `json:"isWatched"`
	IsNextUp      bool               `json:"isNextUp"`
}

type FranchiseRoadmap struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	BackdropPath   string          `json:"backdropPath,omitempty"`
	PosterPath     string          `json:"posterPath,omitempty"`
	ActiveOrder    string          `json:"activeOrder"` // "timeline" or "release"
	TotalItems     int             `json:"totalItems"`
	WatchedCount   int             `json:"watchedCount"`
	PercentWatched int             `json:"percentWatched"`
	NextUpItem     *FranchiseItem  `json:"nextUpItem,omitempty"`
	Items          []FranchiseItem `json:"items"`
}

type FranchiseSummary struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	BackdropPath   string `json:"backdropPath,omitempty"`
	PosterPath     string `json:"posterPath,omitempty"`
	TotalItems     int    `json:"totalItems"`
	WatchedCount   int    `json:"watchedCount"`
	PercentWatched int    `json:"percentWatched"`
	ItemTypes      string `json:"itemTypes"` // e.g. "Filmes & Séries" or "Filmes"
}

type FranchiseMatch struct {
	InFranchise   bool   `json:"inFranchise"`
	FranchiseID   string `json:"franchiseId,omitempty"`
	FranchiseName string `json:"franchiseName,omitempty"`
	PhaseOrEra    string `json:"phaseOrEra,omitempty"`
	TimelineOrder int    `json:"timelineOrder,omitempty"`
	ReleaseOrder  int    `json:"releaseOrder,omitempty"`
}

type CuratedFranchiseDefinition struct {
	ID           string
	Name         string
	Description  string
	BackdropPath string
	PosterPath   string
	ItemTypes    string
	Items        []FranchiseItem
}

