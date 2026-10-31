package calendar

import (
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"sync"
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

	regionForCache := req.Region
	if req.Filter == domain.CalendarFilterLibrary {
		regionForCache = ""
	}
	cacheKey := cache.CreateCacheKey("CalendarEvents", userId, req.Start, req.End, string(req.Filter), string(req.Type), regionForCache)
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
	var eventsMu sync.Mutex

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

	// 1. Process Library Movies (in-memory, instant)
	if req.Type == domain.CalendarTypeAll || req.Type == domain.CalendarTypeMovie {
		for i := range watchedItems {
			w := &watchedItems[i]
			if w.Content == nil || w.Content.Type != entity.MOVIE {
				continue
			}
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
	}

	// 2. Process Library Shows (pruned candidates + parallel worker pool)
	if (req.Type == domain.CalendarTypeAll || req.Type == domain.CalendarTypeTV) && s.tmdb != nil {
		var candidateShows []*entity.Watched
		for i := range watchedItems {
			w := &watchedItems[i]
			if w.Content == nil || w.Content.Type != entity.SHOW {
				continue
			}

			// Optimization: Skip shows that definitely cannot have active/new episodes
			// 1) User already finished the show AND TMDB status is "Ended" or "Canceled"
			if w.Status == entity.FINISHED && (w.Content.Status == "Ended" || w.Content.Status == "Canceled") {
				continue
			}

			// 2) The show is "Ended" or "Canceled" and premiered > 6 months before the requested start date
			if (w.Content.Status == "Ended" || w.Content.Status == "Canceled") && w.Content.ReleaseDate != nil {
				if w.Content.ReleaseDate.Before(startDate.AddDate(0, -6, 0)) {
					continue
				}
			}

			candidateShows = append(candidateShows, w)
		}

		if len(candidateShows) > 0 {
			var wg sync.WaitGroup
			sem := make(chan struct{}, 12) // Limit concurrent outbound HTTP calls to 12

			for _, show := range candidateShows {
				wg.Add(1)
				sem <- struct{}{}

				go func(w *entity.Watched) {
					defer func() {
						<-sem
						wg.Done()
					}()

					showDetails, err := s.tmdb.ShowDetails(tmdb.ShowDetailsOptions{
						ID:             strconv.Itoa(w.Content.TmdbID),
						DontRunDBCache: true, // Avoid SQLite locking from concurrent writes
					})
					if err != nil {
						slog.Debug("Calendar: Failed to fetch show details from TMDB", "tmdbId", w.Content.TmdbID, "error", err)
						return
					}

					poster := showDetails.PosterPath
					if poster == "" && w.Content != nil {
						poster = w.Content.PosterPath
					}

					var showEvents []domain.CalendarEvent

					// Check NextEpisodeToAir
					if showDetails.NextEpisodeToAir != nil && showDetails.NextEpisodeToAir.AirDate != "" {
						air := showDetails.NextEpisodeToAir.AirDate
						if air >= req.Start && air <= req.End {
							evID := fmt.Sprintf("tv-%d-s%d-e%d", w.Content.TmdbID, showDetails.NextEpisodeToAir.SeasonNumber, showDetails.NextEpisodeToAir.EpisodeNumber)
							showEvents = append(showEvents, domain.CalendarEvent{
								ID:            evID,
								MediaID:       w.Content.TmdbID,
								MediaType:     "tv",
								Title:         showDetails.Name,
								ReleaseDate:   air,
								PosterPath:    poster,
								BackdropPath:  showDetails.BackdropPath,
								Overview:      showDetails.NextEpisodeToAir.Overview,
								InLibrary:     true,
								Status:        string(w.Status),
								SeasonNumber:  showDetails.NextEpisodeToAir.SeasonNumber,
								EpisodeNumber: showDetails.NextEpisodeToAir.EpisodeNumber,
								EpisodeName:   showDetails.NextEpisodeToAir.Name,
							})
						}
					}

					// Check LastEpisodeToAir
					if showDetails.LastEpisodeToAir.AirDate != "" {
						air := showDetails.LastEpisodeToAir.AirDate
						if air >= req.Start && air <= req.End {
							evID := fmt.Sprintf("tv-%d-s%d-e%d", w.Content.TmdbID, showDetails.LastEpisodeToAir.SeasonNumber, showDetails.LastEpisodeToAir.EpisodeNumber)
							showEvents = append(showEvents, domain.CalendarEvent{
								ID:            evID,
								MediaID:       w.Content.TmdbID,
								MediaType:     "tv",
								Title:         showDetails.Name,
								ReleaseDate:   air,
								PosterPath:    poster,
								BackdropPath:  showDetails.BackdropPath,
								Overview:      showDetails.LastEpisodeToAir.Overview,
								InLibrary:     true,
								Status:        string(w.Status),
								SeasonNumber:  showDetails.LastEpisodeToAir.SeasonNumber,
								EpisodeNumber: showDetails.LastEpisodeToAir.EpisodeNumber,
								EpisodeName:   showDetails.LastEpisodeToAir.Name,
							})
						}
					}

					// Check latest season episodes ONLY if season premiere is reasonably close to calendar window
					if len(showDetails.Seasons) > 0 {
						latestSeason := showDetails.Seasons[len(showDetails.Seasons)-1]
						if latestSeason.SeasonNumber > 0 && latestSeason.AirDate != "" {
							sAirDate, err := time.Parse("2006-01-02", latestSeason.AirDate)
							seasonEligible := false
							if err == nil {
								// Seasons typically air within 4 months of their start date
								if !sAirDate.Before(startDate.AddDate(0, -4, 0)) && !sAirDate.After(endDate.AddDate(0, 1, 0)) {
									seasonEligible = true
								}
							} else {
								seasonEligible = true
							}

							if seasonEligible {
								seasonDetails, err := s.tmdb.SeasonDetails(strconv.Itoa(w.Content.TmdbID), strconv.Itoa(latestSeason.SeasonNumber))
								if err == nil {
									for _, ep := range seasonDetails.Episodes {
										if ep.AirDate != "" && ep.AirDate >= req.Start && ep.AirDate <= req.End {
											evID := fmt.Sprintf("tv-%d-s%d-e%d", w.Content.TmdbID, ep.SeasonNumber, ep.EpisodeNumber)
											showEvents = append(showEvents, domain.CalendarEvent{
												ID:            evID,
												MediaID:       w.Content.TmdbID,
												MediaType:     "tv",
												Title:         showDetails.Name,
												ReleaseDate:   ep.AirDate,
												PosterPath:    poster,
												BackdropPath:  showDetails.BackdropPath,
												Overview:      ep.Overview,
												InLibrary:     true,
												Status:        string(w.Status),
												SeasonNumber:  ep.SeasonNumber,
												EpisodeNumber: ep.EpisodeNumber,
												EpisodeName:   ep.Name,
											})
										}
									}
								}
							}
						}
					}

					if len(showEvents) > 0 {
						eventsMu.Lock()
						for _, ev := range showEvents {
							if _, exists := eventsMap[ev.ID]; !exists {
								eventsMap[ev.ID] = ev
							}
						}
						eventsMu.Unlock()
					}
				}(show)
			}

			wg.Wait()
		}
	}

	// 3. If Filter == "all", add Discover releases from TMDB (running Movies and Shows in parallel)
	if req.Filter == domain.CalendarFilterAll && s.tmdb != nil {
		var discWg sync.WaitGroup

		if req.Type == domain.CalendarTypeAll || req.Type == domain.CalendarTypeMovie {
			discWg.Add(1)
			go func() {
				defer discWg.Done()
				discMovies, err := s.tmdb.DiscoverMovies(tmdb.DiscoverOptions{
					ReleaseDateMin: startDate,
					ReleaseDateMax: endDate,
				}, 1, req.Region)
				if err == nil {
					eventsMu.Lock()
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
					eventsMu.Unlock()
				}
			}()
		}

		if req.Type == domain.CalendarTypeAll || req.Type == domain.CalendarTypeTV {
			discWg.Add(1)
			go func() {
				defer discWg.Done()
				discShows, err := s.tmdb.DiscoverShows(tmdb.DiscoverOptions{
					ReleaseDateMin: startDate,
					ReleaseDateMax: endDate,
				}, 1, req.Region)
				if err == nil {
					eventsMu.Lock()
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
					eventsMu.Unlock()
				}
			}()
		}

		discWg.Wait()
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
