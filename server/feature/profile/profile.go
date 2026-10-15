package profile

import (
	"errors"
	"log/slog"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
	"gorm.io/gorm"
)

type RecentActivityItem struct {
	ID         uint                `json:"id"`
	Type       entity.ActivityType `json:"type"`
	Data       string              `json:"data"`
	Date       time.Time           `json:"date"`
	MediaTitle string              `json:"mediaTitle"`
	MediaType  string              `json:"mediaType"`
	MediaID    int                 `json:"mediaId"`
	PosterPath string              `json:"posterPath,omitempty"`
}

type ActivityPeriod struct {
	Period   string `json:"period"` // "2024-05" or "2024"
	Count    int    `json:"count"`
	Movies   int    `json:"movies"`
	Episodes int    `json:"episodes"`
}

type RatingStats struct {
	Count        int            `json:"count"`
	Average      float64        `json:"average"`
	Highest      float64        `json:"highest"`
	Lowest       float64        `json:"lowest"`
	Distribution map[string]int `json:"distribution"` // "1".."10"
}

type GenreStat struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type Profile struct {
	Joined               time.Time `json:"joined"`
	ShowsWatched         int32     `json:"showsWatched"` // Finished shows
	MoviesWatched        int32     `json:"moviesWatched"` // Finished movies
	MoviesWatchedRuntime uint32    `json:"moviesWatchedRuntime"` // in minutes
	ShowsWatchedRuntime  uint32    `json:"showsWatchedRuntime"` // in minutes

	// Extended metrics
	ShowsWatching          int32                     `json:"showsWatching"`
	EpisodesWatched        int32                     `json:"episodesWatched"`
	GamesPlayed            int32                     `json:"gamesPlayed"`
	TotalRuntime           uint64                    `json:"totalRuntime"` // in minutes
	StatusDistribution     map[string]int            `json:"statusDistribution"`
	StatusDistributionType map[string]map[string]int `json:"statusDistributionType"`
	RatingStats            RatingStats               `json:"ratingStats"`
	ActivityByMonth        []ActivityPeriod          `json:"activityByMonth"`
	ActivityByYear         []ActivityPeriod          `json:"activityByYear"`
	GenreDistribution      []GenreStat               `json:"genreDistribution"`
	RecentActivity         []RecentActivityItem      `json:"recentActivity"`
	HasPartialHistory      bool                      `json:"hasPartialHistory"`
}

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db,
	}
}

// Checks if item has been previously watched by scanning for any activity
// that counts as a play.
func (s *Service) hasBeenPreviouslyWatched(a *[]entity.Activity) bool {
	for _, v := range *a {
		if v.CountAsPlay {
			return true
		}
	}
	return false
}

// Gets any data required for profile page
func (s *Service) getProfile(userId uint) (Profile, error) {
	// Get user.
	user := new(entity.User)
	res := s.db.Model(&entity.User{}).Where("id = ?", userId).Take(&user)
	if res.Error != nil {
		slog.Error("Failed to get profile:", "error", res.Error)
		return Profile{}, errors.New("failed to get profile")
	}

	// Process stats with watched items and relations.
	var watched []entity.Watched
	res = s.db.Model(&entity.Watched{}).
		Preload("Content").
		Preload("Game").
		Preload("Game.Poster").
		Preload("Activity").
		Preload("WatchedEpisodes").
		Where("user_id = ?", userId).
		Find(&watched)
	if res.Error != nil {
		slog.Error("Profile: Failed to get watched for processing:", "error", res.Error)
		return Profile{}, errors.New("failed to get watched for processing")
	}

	var (
		showsWatched         int32
		showsWatching        int32
		moviesWatched        int32
		gamesPlayed          int32
		episodesWatched      int32
		moviesWatchedRuntime uint32
		showsWatchedRuntime  uint32
		hasPartialHistory    bool
	)

	statusDist := map[string]int{
		"FINISHED": 0,
		"WATCHING": 0,
		"PLANNED":  0,
		"HOLD":     0,
		"DROPPED":  0,
	}

	statusDistType := map[string]map[string]int{
		"movie": {"FINISHED": 0, "WATCHING": 0, "PLANNED": 0, "HOLD": 0, "DROPPED": 0},
		"tv":    {"FINISHED": 0, "WATCHING": 0, "PLANNED": 0, "HOLD": 0, "DROPPED": 0},
		"game":  {"FINISHED": 0, "WATCHING": 0, "PLANNED": 0, "HOLD": 0, "DROPPED": 0},
	}

	ratingDist := map[string]int{
		"1": 0, "2": 0, "3": 0, "4": 0, "5": 0,
		"6": 0, "7": 0, "8": 0, "9": 0, "10": 0,
	}

	var (
		ratingCount int
		ratingSum   float64
		highestRate float64
		lowestRate  float64 = 10.0
	)

	genreCounts := make(map[string]int)
	monthMap := make(map[string]*ActivityPeriod)
	yearMap := make(map[string]*ActivityPeriod)
	var recentActivities []RecentActivityItem

	for _, w := range watched {
		mediaType := "movie"
		mediaTitle := ""
		mediaID := 0
		posterPath := ""

		if w.Game != nil {
			mediaType = "game"
			mediaTitle = w.Game.Name
			mediaID = w.Game.IgdbID
			if w.Game.Poster != nil {
				posterPath = w.Game.Poster.Path
			}
			if w.Game.Genres != "" {
				for _, g := range strings.Split(w.Game.Genres, "|") {
					g = strings.TrimSpace(g)
					if g != "" {
						genreCounts[g]++
					}
				}
			}
		} else if w.Content != nil {
			mediaTitle = w.Content.Title
			mediaID = w.Content.TmdbID
			posterPath = w.Content.PosterPath
			if w.Content.Type == entity.SHOW {
				mediaType = "tv"
			}
			if w.Content.Genres != "" {
				for _, g := range strings.Split(w.Content.Genres, ",") {
					g = strings.TrimSpace(g)
					if g != "" {
						genreCounts[g]++
					}
				}
			}
		}

		// Status distribution
		st := string(w.Status)
		if st != "" {
			statusDist[st]++
			if m, ok := statusDistType[mediaType]; ok {
				m[st]++
			}
		}

		// Rating stats
		if w.Rating > 0 {
			ratingCount++
			ratingSum += w.Rating
			if w.Rating > highestRate {
				highestRate = w.Rating
			}
			if w.Rating < lowestRate {
				lowestRate = w.Rating
			}
			bucket := int(math.Round(w.Rating))
			if bucket < 1 {
				bucket = 1
			} else if bucket > 10 {
				bucket = 10
			}
			ratingDist[strconv.Itoa(bucket)]++
		}

		// Finished check
		isFinished := w.Status == entity.FINISHED || s.hasBeenPreviouslyWatched(&w.Activity)

		// Media specific processing
		if w.Game != nil {
			if isFinished {
				gamesPlayed++
			}
		} else if w.Content != nil {
			c := *w.Content
			switch c.Type {
			case entity.MOVIE:
				if isFinished {
					moviesWatched++
					if c.Runtime > 0 {
						moviesWatchedRuntime += c.Runtime
					} else {
						hasPartialHistory = true
					}
				}
			case entity.SHOW:
				if isFinished {
					showsWatched++
				} else if w.Status == entity.WATCHING {
					showsWatching++
				}

				// Calculate accurately watched episodes
				episodesInRecord := 0
				for _, we := range w.WatchedEpisodes {
					if we.Status == entity.FINISHED {
						episodesInRecord++
					}
				}

				if episodesInRecord > 0 {
					episodesWatched += int32(episodesInRecord)
					if c.Runtime > 0 {
						showsWatchedRuntime += uint32(episodesInRecord) * c.Runtime
					} else {
						hasPartialHistory = true
					}
				} else if isFinished {
					// Fallback when entire show was marked finished without individual episode entries
					if c.NumberOfEpisodes > 0 {
						episodesWatched += int32(c.NumberOfEpisodes)
						if c.Runtime > 0 {
							showsWatchedRuntime += c.NumberOfEpisodes * c.Runtime
						} else {
							hasPartialHistory = true
						}
					} else {
						hasPartialHistory = true
					}
				}
			}
		}

		// Activity items processing
		for _, a := range w.Activity {
			actDate := a.CreatedAt
			if a.CustomDate != nil && !a.CustomDate.IsZero() {
				actDate = *a.CustomDate
			}

			// Add to recent activity stream
			recentActivities = append(recentActivities, RecentActivityItem{
				ID:         a.ID,
				Type:       a.Type,
				Data:       a.Data,
				Date:       actDate,
				MediaTitle: mediaTitle,
				MediaType:  mediaType,
				MediaID:    mediaID,
				PosterPath: posterPath,
			})

			// Activity period grouping
			isPlayActivity := a.CountAsPlay ||
				(a.Type == entity.STATUS_CHANGED && (a.Data == "FINISHED" || a.Data == "WATCHING")) ||
				a.Type == entity.IMPORTED_ADDED_WATCHED ||
				a.Type == entity.ADDED_WATCHED

			if isPlayActivity && !actDate.IsZero() && actDate.Year() >= 1990 {
				mKey := actDate.Format("2006-01")
				yKey := actDate.Format("2006")

				// Month
				if _, ok := monthMap[mKey]; !ok {
					monthMap[mKey] = &ActivityPeriod{Period: mKey}
				}
				monthMap[mKey].Count++
				if mediaType == "movie" {
					monthMap[mKey].Movies++
				} else {
					monthMap[mKey].Episodes++
				}

				// Year
				if _, ok := yearMap[yKey]; !ok {
					yearMap[yKey] = &ActivityPeriod{Period: yKey}
				}
				yearMap[yKey].Count++
				if mediaType == "movie" {
					yearMap[yKey].Movies++
				} else {
					yearMap[yKey].Episodes++
				}
			}
		}
	}

	// Sort recent activity descending
	sort.Slice(recentActivities, func(i, j int) bool {
		return recentActivities[i].Date.After(recentActivities[j].Date)
	})
	if len(recentActivities) > 15 {
		recentActivities = recentActivities[:15]
	}

	// Format monthly activities
	var activityByMonth []ActivityPeriod
	for _, v := range monthMap {
		activityByMonth = append(activityByMonth, *v)
	}
	sort.Slice(activityByMonth, func(i, j int) bool {
		return activityByMonth[i].Period < activityByMonth[j].Period
	})

	// Format yearly activities
	var activityByYear []ActivityPeriod
	for _, v := range yearMap {
		activityByYear = append(activityByYear, *v)
	}
	sort.Slice(activityByYear, func(i, j int) bool {
		return activityByYear[i].Period < activityByYear[j].Period
	})

	// Format genres
	var genreDist []GenreStat
	for name, count := range genreCounts {
		genreDist = append(genreDist, GenreStat{Name: name, Count: count})
	}
	sort.Slice(genreDist, func(i, j int) bool {
		return genreDist[i].Count > genreDist[j].Count
	})
	if len(genreDist) > 12 {
		genreDist = genreDist[:12]
	}

	// Rating calculations
	var avgRating float64
	if ratingCount > 0 {
		avgRating = math.Round((ratingSum/float64(ratingCount))*10) / 10
	} else {
		lowestRate = 0
	}

	rStats := RatingStats{
		Count:        ratingCount,
		Average:      avgRating,
		Highest:      highestRate,
		Lowest:       lowestRate,
		Distribution: ratingDist,
	}

	totalRuntime := uint64(moviesWatchedRuntime) + uint64(showsWatchedRuntime)

	profile := Profile{
		Joined:                 user.CreatedAt,
		ShowsWatched:           showsWatched,
		ShowsWatching:          showsWatching,
		MoviesWatched:          moviesWatched,
		GamesPlayed:            gamesPlayed,
		EpisodesWatched:        episodesWatched,
		MoviesWatchedRuntime:   moviesWatchedRuntime,
		ShowsWatchedRuntime:    showsWatchedRuntime,
		TotalRuntime:           totalRuntime,
		StatusDistribution:     statusDist,
		StatusDistributionType: statusDistType,
		RatingStats:            rStats,
		ActivityByMonth:        activityByMonth,
		ActivityByYear:         activityByYear,
		GenreDistribution:      genreDist,
		RecentActivity:         recentActivities,
		HasPartialHistory:      hasPartialHistory,
	}

	return profile, nil
}
