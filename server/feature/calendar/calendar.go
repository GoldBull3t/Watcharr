package calendar

import (
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"time"

	gocache "github.com/robfig/go-cache"
	"github.com/sbondCo/Watcharr/cache"
	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/media/tmdb"
	"gorm.io/gorm"
)

var CalendarStore = gocache.New(time.Minute*10, time.Minute*5)

type Service struct {
	db   *gorm.DB
	cfg  *config.ServerConfig
	tmdb *tmdb.TMDB
}

func NewService(db *gorm.DB, cfg *config.ServerConfig, tmdb *tmdb.TMDB) *Service {
	return &Service{
		db:   db,
		cfg:  cfg,
		tmdb: tmdb,
	}
}

func (s *Service) GetEvents(userId uint, req domain.CalendarRequest) ([]domain.CalendarEvent, error) {
	now := time.Now()
	if req.Start == "" {
		req.Start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
	}
	if req.End == "" {
		firstNextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC)
		req.End = firstNextMonth.AddDate(0, 0, 7).Format("2006-01-02")
	}
	if req.Filter == "" {
		req.Filter = domain.CalendarFilterLibrary
	}
	if req.Type == "" {
		req.Type = domain.CalendarTypeAll
	}

	cacheKey := cache.CreateCacheKey("CalendarEvents", userId, req.Start, req.End, string(req.Filter), string(req.Type), req.Region)
	var cachedEvents []domain.CalendarEvent
	if cache.GetCache(CalendarStore, cacheKey, &cachedEvents) {
		slog.Debug("Calendar: Returning cached events", "key", cacheKey)
		return cachedEvents, nil
	}

	startDate, err := time.Parse("2006-01-02", req.Start)
	if err != nil {
		startDate = time.Now().AddDate(0, 0, -15)
	}
	endDate, err := time.Parse("2006-01-02", req.End)
	if err != nil {
		endDate = time.Now().AddDate(0, 0, 45)
	}

	eventsMap := make(map[string]domain.CalendarEvent)

	// Fetch user's library items (exclude soft-deleted and dropped)
	var watchedItems []entity.Watched
	err = s.db.Preload("Content").
		Where("user_id = ? AND status != ?", userId, entity.DROPPED).
		Find(&watchedItems).Error
	if err != nil {
		slog.Error("Calendar: Failed to query watched items", "error", err)
	}

	// Map of library items by tmdbId and contentType
	type mediaKey struct {
		id int
		ct entity.ContentType
	}
	libraryMap := make(map[mediaKey]*entity.Watched)
	for i := range watchedItems {
		w := &watchedItems[i]
		if w.Content != nil {
			libraryMap[mediaKey{id: w.Content.TmdbID, ct: w.Content.Type}] = w
		}
	}

	// 1. Process Library items
	for i := range watchedItems {
		w := &watchedItems[i]
		if w.Content == nil {
			continue
		}

		// Movies in user's library
		if (req.Type == domain.CalendarTypeAll || req.Type == domain.CalendarTypeMovie) && w.Content.Type == entity.MOVIE {
			if w.Content.ReleaseDate != nil {
				relStr := w.Content.ReleaseDate.Format("2006-01-02")
				if relStr >= req.Start && relStr <= req.End {
					evID := fmt.Sprintf("movie-%d", w.Content.TmdbID)
					eventsMap[evID] = domain.CalendarEvent{
						ID:          evID,
						MediaID:     w.Content.TmdbID,
						MediaType:   "movie",
						Title:       w.Content.Title,
						ReleaseDate: relStr,
						PosterPath:  w.Content.PosterPath,
						Overview:    w.Content.Overview,
						InLibrary:   true,
						Status:      string(w.Status),
					}
				}
			}
		}

		// Shows in user's library
		if (req.Type == domain.CalendarTypeAll || req.Type == domain.CalendarTypeTV) && w.Content.Type == entity.SHOW {
			if s.tmdb != nil {
				showDetails, err := s.tmdb.ShowDetails(tmdb.ShowDetailsOptions{
					ID: strconv.Itoa(w.Content.TmdbID),
				})
				if err == nil {
					// Check NextEpisodeToAir
					if showDetails.NextEpisodeToAir != nil && showDetails.NextEpisodeToAir.AirDate != "" {
						air := showDetails.NextEpisodeToAir.AirDate
						if air >= req.Start && air <= req.End {
							evID := fmt.Sprintf("tv-%d-s%d-e%d", w.Content.TmdbID, showDetails.NextEpisodeToAir.SeasonNumber, showDetails.NextEpisodeToAir.EpisodeNumber)
							eventsMap[evID] = domain.CalendarEvent{
								ID:            evID,
								MediaID:       w.Content.TmdbID,
								MediaType:     "tv",
								Title:         showDetails.Name,
								ReleaseDate:   air,
								PosterPath:    showDetails.PosterPath,
								BackdropPath:  showDetails.BackdropPath,
								Overview:      showDetails.NextEpisodeToAir.Overview,
								InLibrary:     true,
								Status:        string(w.Status),
								SeasonNumber:  showDetails.NextEpisodeToAir.SeasonNumber,
								EpisodeNumber: showDetails.NextEpisodeToAir.EpisodeNumber,
								EpisodeName:   showDetails.NextEpisodeToAir.Name,
							}
						}
					}

					// Check LastEpisodeToAir
					if showDetails.LastEpisodeToAir.AirDate != "" {
						air := showDetails.LastEpisodeToAir.AirDate
						if air >= req.Start && air <= req.End {
							evID := fmt.Sprintf("tv-%d-s%d-e%d", w.Content.TmdbID, showDetails.LastEpisodeToAir.SeasonNumber, showDetails.LastEpisodeToAir.EpisodeNumber)
							if _, exists := eventsMap[evID]; !exists {
								eventsMap[evID] = domain.CalendarEvent{
									ID:            evID,
									MediaID:       w.Content.TmdbID,
									MediaType:     "tv",
									Title:         showDetails.Name,
									ReleaseDate:   air,
									PosterPath:    showDetails.PosterPath,
									BackdropPath:  showDetails.BackdropPath,
									Overview:      showDetails.LastEpisodeToAir.Overview,
									InLibrary:     true,
									Status:        string(w.Status),
									SeasonNumber:  showDetails.LastEpisodeToAir.SeasonNumber,
									EpisodeNumber: showDetails.LastEpisodeToAir.EpisodeNumber,
									EpisodeName:   showDetails.LastEpisodeToAir.Name,
								}
							}
						}
					}

					// If has seasons, check the latest season's episodes for dates in the range
					if len(showDetails.Seasons) > 0 {
						latestSeason := showDetails.Seasons[len(showDetails.Seasons)-1]
						if latestSeason.SeasonNumber > 0 {
							seasonDetails, err := s.tmdb.SeasonDetails(strconv.Itoa(w.Content.TmdbID), strconv.Itoa(latestSeason.SeasonNumber))
							if err == nil {
								for _, ep := range seasonDetails.Episodes {
									if ep.AirDate != "" && ep.AirDate >= req.Start && ep.AirDate <= req.End {
										evID := fmt.Sprintf("tv-%d-s%d-e%d", w.Content.TmdbID, ep.SeasonNumber, ep.EpisodeNumber)
										if _, exists := eventsMap[evID]; !exists {
											eventsMap[evID] = domain.CalendarEvent{
												ID:            evID,
												MediaID:       w.Content.TmdbID,
												MediaType:     "tv",
												Title:         showDetails.Name,
												ReleaseDate:   ep.AirDate,
												PosterPath:    showDetails.PosterPath,
												BackdropPath:  showDetails.BackdropPath,
												Overview:      ep.Overview,
												InLibrary:     true,
												Status:        string(w.Status),
												SeasonNumber:  ep.SeasonNumber,
												EpisodeNumber: ep.EpisodeNumber,
												EpisodeName:   ep.Name,
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// 2. If Filter == "all", add Discover releases from TMDB
	if req.Filter == domain.CalendarFilterAll && s.tmdb != nil {
		if req.Type == domain.CalendarTypeAll || req.Type == domain.CalendarTypeMovie {
			discMovies, err := s.tmdb.DiscoverMovies(tmdb.DiscoverOptions{
				ReleaseDateMin: startDate,
				ReleaseDateMax: endDate,
			}, 1, req.Region)
			if err == nil {
				for _, m := range discMovies.Results {
					if m.ReleaseDate != "" && m.ReleaseDate >= req.Start && m.ReleaseDate <= req.End {
						evID := fmt.Sprintf("movie-%d", m.ID)
						w, inLib := libraryMap[mediaKey{id: m.ID, ct: entity.MOVIE}]
						var status string
						if inLib && w != nil {
							status = string(w.Status)
						}
						eventsMap[evID] = domain.CalendarEvent{
							ID:           evID,
							MediaID:      m.ID,
							MediaType:    "movie",
							Title:        m.Title,
							ReleaseDate:  m.ReleaseDate,
							PosterPath:   m.PosterPath,
							BackdropPath: m.BackdropPath,
							Overview:     m.Overview,
							InLibrary:    inLib,
							Status:       status,
						}
					}
				}
			}
		}

		if req.Type == domain.CalendarTypeAll || req.Type == domain.CalendarTypeTV {
			discShows, err := s.tmdb.DiscoverShows(tmdb.DiscoverOptions{
				ReleaseDateMin: startDate,
				ReleaseDateMax: endDate,
			}, 1, req.Region)
			if err == nil {
				for _, sh := range discShows.Results {
					if sh.FirstAirDate != "" && sh.FirstAirDate >= req.Start && sh.FirstAirDate <= req.End {
						evID := fmt.Sprintf("tv-%d", sh.ID)
						w, inLib := libraryMap[mediaKey{id: sh.ID, ct: entity.SHOW}]
						var status string
						if inLib && w != nil {
							status = string(w.Status)
						}
						if _, exists := eventsMap[evID]; !exists {
							eventsMap[evID] = domain.CalendarEvent{
								ID:           evID,
								MediaID:      sh.ID,
								MediaType:    "tv",
								Title:        sh.Name,
								ReleaseDate:  sh.FirstAirDate,
								PosterPath:   sh.PosterPath,
								BackdropPath: sh.BackdropPath,
								Overview:     sh.Overview,
								InLibrary:    inLib,
								Status:       status,
							}
						}
					}
				}
			}
		}
	}

	events := make([]domain.CalendarEvent, 0, len(eventsMap))
	for _, ev := range eventsMap {
		events = append(events, ev)
	}

	// Sort by release date ascending, then by title
	sort.Slice(events, func(i, j int) bool {
		if events[i].ReleaseDate != events[j].ReleaseDate {
			return events[i].ReleaseDate < events[j].ReleaseDate
		}
		return events[i].Title < events[j].Title
	})

	CalendarStore.Set(cacheKey, events, time.Minute*10)
	return events, nil
}
