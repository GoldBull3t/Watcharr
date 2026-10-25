package igdb

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/sbondCo/Watcharr/domain"
)

type TwitchTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// So we can unmarshall the unix timestamps returned from igdb into time.Time.
type UnixTime struct {
	time.Time
}

func (u *UnixTime) UnmarshalJSON(b []byte) error {
	var timestamp int64
	err := json.Unmarshal(b, &timestamp)
	if err != nil {
		return err
	}
	u.Time = time.Unix(timestamp, 0)
	return nil
}

// Website Type Enum
type WebsiteType int

const (
	WebsiteTypeOfficial  WebsiteType = 1
	WebsiteTypeWikia     WebsiteType = 2
	WebsiteTypeWikipedia WebsiteType = 3
	WebsiteTypeSteam     WebsiteType = 13
	WebsiteTypeReddit    WebsiteType = 14
	WebsiteTypeItch      WebsiteType = 15
	WebsiteTypeEpicGames WebsiteType = 16
	WebsiteTypeGOG       WebsiteType = 17
	WebsiteTypeDiscord   WebsiteType = 18
)

// Only the fields we request included in each struct

// Search

type GameSearchResponseResult struct {
	ID    int `json:"id"`
	Cover struct {
		ID      int    `json:"id"`
		ImageID string `json:"image_id"`
	} `json:"cover"`
	Artworks []struct {
		ImageID string `json:"image_id"`
	} `json:"artworks"`
	FirstReleaseDate UnixTime `json:"first_release_date"`
	Name             string   `json:"name"`
	Summary          string   `json:"summary,omitempty"`
	VersionTitle     string   `json:"version_title,omitempty"`
	Rating           float64  `json:"rating,omitempty"`
	RatingCount      int      `json:"rating_count,omitempty"`
}

func (t *GameSearchResponseResult) AsMedia() domain.Media {
	m := domain.Media{
		Type: domain.MediaTypeIGDBGame,
		IDs: domain.MediaIDs{
			IGDB: t.ID,
		},
		Name:          t.Name,
		Summary:       t.Summary,
		ExtPosterPath: t.Cover.ImageID,
		ReleaseDate:   t.FirstReleaseDate.Time,
		Rating:        uint(t.Rating),
		RatingCount:   uint(t.RatingCount),
	}
	if len(t.Artworks) > 0 {
		m.ExtBackdropPath = t.Artworks[0].ImageID
	}
	return m
}

// This type is used for quite a few of our service funcs, maybe be a bit
// careful when updating it and separate out stuff if need be.
type GameSearchResponse []GameSearchResponseResult

// Similar

type GameSimilar struct {
	ID               int      `json:"id"`
	Name             string   `json:"name"`
	Summary          string   `json:"summary"`
	FirstReleaseDate UnixTime `json:"first_release_date"`
	Rating           float64  `json:"rating,omitempty"`
	RatingCount      int      `json:"rating_count,omitempty"`
	Cover            struct {
		ID      int    `json:"id"`
		ImageID string `json:"image_id"`
	} `json:"cover"`
	Artworks []struct {
		ImageID string `json:"image_id"`
	} `json:"artworks"`
}

func (t *GameSimilar) AsMedia() domain.Media {
	m := domain.Media{
		Type: domain.MediaTypeIGDBGame,
		IDs: domain.MediaIDs{
			IGDB: t.ID,
		},
		Name:          t.Name,
		Summary:       t.Summary,
		ExtPosterPath: t.Cover.ImageID,
		ReleaseDate:   t.FirstReleaseDate.Time,
		Rating:        uint(t.Rating),
		RatingCount:   uint(t.RatingCount),
	}
	if len(t.Artworks) > 0 {
		m.ExtBackdropPath = t.Artworks[0].ImageID
	}
	return m
}

// Details

type GameTimeToBeatResponse struct {
	Hastily    int `json:"hastily"`    // seconds
	Normally   int `json:"normally"`   // seconds
	Completely int `json:"completely"` // seconds
	Count      int `json:"count"`
}

type GameDetailsResponse struct {
	ID       int `json:"id"`
	Artworks []struct {
		Width   int    `json:"width"`
		Height  int    `json:"height"`
		ImageID string `json:"image_id"`
	} `json:"artworks"`
	Screenshots []struct {
		Width   int    `json:"width"`
		Height  int    `json:"height"`
		ImageID string `json:"image_id"`
	} `json:"screenshots"`
	GameType int `json:"game_type"`
	Cover    struct {
		ID      int    `json:"id"`
		ImageID string `json:"image_id"`
	} `json:"cover"`
	FirstReleaseDate UnixTime `json:"first_release_date"`
	GameModes        []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"game_modes"`
	Genres []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
	Themes []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"themes"`
	PlayerPerspectives []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"player_perspectives"`
	InvolvedCompanies []struct {
		ID      int `json:"id"`
		Company struct {
			ID          int    `json:"id"`
			Description string `json:"description"`
			Name        string `json:"name"`
			Slug        string `json:"slug"`
			Websites    []struct {
				ID      int    `json:"id"`
				Type    int    `json:"type"`
				Trusted bool   `json:"trusted"`
				URL     string `json:"url"`
			} `json:"websites"`
		} `json:"company"`
		Developer  bool `json:"developer"`
		Porting    bool `json:"porting"`
		Publisher  bool `json:"publisher"`
		Supporting bool `json:"supporting"`
	} `json:"involved_companies"`
	Name      string `json:"name"`
	Platforms []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"platforms"`
	Rating                float64 `json:"rating"`
	RatingCount           int     `json:"rating_count"`
	AggregatedRating      float64 `json:"aggregated_rating"`
	AggregatedRatingCount int     `json:"aggregated_rating_count"`
	Summary               string  `json:"summary"`
	Storyline             string  `json:"storyline"`
	Status                int     `json:"status"`
	URL                   string  `json:"url"`
	Videos                []struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		VideoID string `json:"video_id"`
	} `json:"videos"`
	Websites []struct {
		ID      int         `json:"id"`
		Type    WebsiteType `json:"type"`
		Trusted bool        `json:"trusted"`
		URL     string      `json:"url"`
	} `json:"websites"`

	SimilarGame []GameSimilar           `json:"similar_games"`
	TimeToBeat  *GameTimeToBeatResponse `json:"-"`
}

func (t *GameDetailsResponse) AsMedia() domain.Media {
	m := domain.Media{
		Type: domain.MediaTypeIGDBGame,
		IDs: domain.MediaIDs{
			IGDB: t.ID,
		},
		Name:              t.Name,
		Summary:           t.Summary,
		Storyline:         t.Storyline,
		ExtPosterPath:     t.Cover.ImageID,
		ReleaseDate:       t.FirstReleaseDate.Time,
		Rating:            uint(t.Rating),
		RatingCount:       uint(t.RatingCount),
		CriticRating:      uint(t.AggregatedRating),
		CriticRatingCount: uint(t.AggregatedRatingCount),
	}
	if len(t.Artworks) > 0 {
		m.ExtBackdropPath = t.Artworks[0].ImageID
		for _, a := range t.Artworks {
			if a.ImageID != "" {
				m.Artworks = append(m.Artworks, a.ImageID)
			}
		}
	} else if len(t.Screenshots) > 0 {
		m.ExtBackdropPath = t.Screenshots[0].ImageID
	}
	for _, s := range t.Screenshots {
		if s.ImageID != "" {
			m.Screenshots = append(m.Screenshots, s.ImageID)
		}
	}
	if t.TimeToBeat != nil {
		m.TimeToBeat = &domain.MediaTimeToBeat{
			Hastily:    t.TimeToBeat.Hastily / 60,
			Normally:   t.TimeToBeat.Normally / 60,
			Completely: t.TimeToBeat.Completely / 60,
			Count:      t.TimeToBeat.Count,
		}
		if m.TimeToBeat.Normally > 0 {
			m.Runtime = uint(m.TimeToBeat.Normally)
		} else if m.TimeToBeat.Hastily > 0 {
			m.Runtime = uint(m.TimeToBeat.Hastily)
		}
	}
	// Platforms, Developers, Publishers, Perspectives
	for _, p := range t.Platforms {
		if p.Name != "" {
			m.Platforms = append(m.Platforms, p.Name)
		}
	}
	for _, ic := range t.InvolvedCompanies {
		if ic.Company.Name == "" {
			continue
		}
		if ic.Developer {
			m.Developers = append(m.Developers, ic.Company.Name)
		}
		if ic.Publisher {
			m.Publishers = append(m.Publishers, ic.Company.Name)
		}
	}
	for _, pp := range t.PlayerPerspectives {
		if pp.Name != "" {
			m.PlayerPerspectives = append(m.PlayerPerspectives, pp.Name)
		}
	}
	// Process websites & store links
	seenProviders := make(map[string]bool)
	addGameProvider := func(name, link string) {
		if seenProviders[name] || link == "" {
			return
		}
		seenProviders[name] = true
		m.Providers = append(m.Providers, domain.MediaProvider{
			Name: name,
			Link: link,
		})
	}
	for _, v := range t.Websites {
		switch v.Type {
		case WebsiteTypeOfficial:
			m.Homepage = v.URL
		case WebsiteTypeSteam:
			addGameProvider("Steam", v.URL)
		case WebsiteTypeEpicGames:
			addGameProvider("Epic Games", v.URL)
		case WebsiteTypeGOG:
			addGameProvider("GOG", v.URL)
		case WebsiteTypeItch:
			addGameProvider("Itch", v.URL)
		case WebsiteTypeWikipedia:
			addGameProvider("Wikipedia", v.URL)
		case WebsiteTypeWikia:
			addGameProvider("Wiki", v.URL)
		case WebsiteTypeReddit:
			addGameProvider("Reddit", v.URL)
		case WebsiteTypeDiscord:
			addGameProvider("Discord", v.URL)
		default:
			uLower := strings.ToLower(v.URL)
			if strings.Contains(uLower, "store.playstation.com") || strings.Contains(uLower, "playstation.com") {
				addGameProvider("PlayStation", v.URL)
			} else if strings.Contains(uLower, "xbox.com") || strings.Contains(uLower, "microsoft.com") {
				addGameProvider("Xbox", v.URL)
			} else if strings.Contains(uLower, "nintendo.com") {
				addGameProvider("Nintendo", v.URL)
			}
		}
	}
	// Genres, Themes, Game modes
	for _, v := range t.Genres {
		m.Genres = append(m.Genres, domain.MediaGenre{
			ID:   uint(v.ID),
			Name: v.Name,
		})
	}
	for _, v := range t.Themes {
		m.Themes = append(m.Themes, domain.MediaGenre{
			ID:   uint(v.ID),
			Name: v.Name,
		})
	}
	for _, v := range t.GameModes {
		m.GameModes = append(m.GameModes, domain.MediaGenre{
			ID:   uint(v.ID),
			Name: v.Name,
		})
	}
	// Videos (Trailers, Gameplay, Teasers, Launch Trailers)
	hasBest := false
	for _, v := range t.Videos {
		if v.VideoID == "" {
			continue
		}
		nameLower := strings.ToLower(v.Name)
		vType := domain.MediaVideoTypeOther
		if strings.Contains(nameLower, "trailer") {
			vType = domain.MediaVideoTypeTrailer
		} else if strings.Contains(nameLower, "teaser") {
			vType = domain.MediaVideoTypeTeaser
		} else if strings.Contains(nameLower, "gameplay") || strings.Contains(nameLower, "clip") {
			vType = domain.MediaVideoTypeClip
		} else if strings.Contains(nameLower, "intro") || strings.Contains(nameLower, "opening") {
			vType = domain.MediaVideoTypeOpeningCredits
		}
		isBest := false
		if (nameLower == "trailer" || nameLower == "launch trailer" || nameLower == "official trailer") && !hasBest {
			isBest = true
			hasBest = true
		}
		m.Videos = append(m.Videos, domain.MediaVideo{
			ID:   v.VideoID,
			Name: v.Name,
			Type: vType,
			Best: isBest,
		})
	}
	if !hasBest {
		for i := range m.Videos {
			if m.Videos[i].Type == domain.MediaVideoTypeTrailer {
				m.Videos[i].Best = true
				break
			}
		}
	}
	// Convert similar items to media too.
	for i := range t.SimilarGame {
		m.Similar = append(m.Similar, t.SimilarGame[i].AsMedia())
	}
	return m
}

// Basic Details

type GameDetailsBasicResponse struct {
	ID       int `json:"id"`
	Category int `json:"category"`
	Cover    struct {
		ImageID string `json:"image_id"`
	} `json:"cover"`
	Artworks []struct {
		ImageID string `json:"image_id"`
	} `json:"artworks"`
	FirstReleaseDate UnixTime `json:"first_release_date"`
	GameModes        []struct {
		Name string `json:"name"`
	} `json:"game_modes"`
	Genres []struct {
		Name string `json:"name"`
	} `json:"genres"`
	Name      string `json:"name"`
	Platforms []struct {
		Name string `json:"name"`
	} `json:"platforms"`
	InvolvedCompanies []struct {
		Developer bool `json:"developer"`
		Publisher bool `json:"publisher"`
		Company   struct {
			Name string `json:"name"`
		} `json:"company"`
	} `json:"involved_companies"`
	Rating           float64 `json:"rating"`
	RatingCount      int     `json:"rating_count"`
	AggregatedRating float64 `json:"aggregated_rating"`
	TimeToBeatNormal int     `json:"-"`
	Summary          string  `json:"summary"`
	Storyline        string  `json:"storyline"`
	Status           int     `json:"status"`
}

// Popularity primitives where we just want the game id.

type PopularityPrimitivesGameIdsResponseResult struct {
	GameID int `json:"game_id"`
}

type PopularityPrimitivesGameIdsResponse []PopularityPrimitivesGameIdsResponseResult
