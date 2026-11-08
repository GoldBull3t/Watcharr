package calendar

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"sort"
	"strconv"
	"strings"
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

var CalendarStore = gocache.New(6*time.Hour, 10*time.Minute)

// InvalidateUserCache evicts all cached calendar event lists for a specific user
func InvalidateUserCache(userId uint) {
	if CalendarStore == nil {
		return
	}
	userPrefix := fmt.Sprintf("CalendarEvents:%d:", userId)
	var keysToDelete []string
	CalendarStore.Lock()
	cv := reflect.ValueOf(CalendarStore)
	if cv.Kind() == reflect.Ptr && !cv.IsNil() {
		inner := cv.Elem().Field(0)
		if inner.Kind() == reflect.Ptr && !inner.IsNil() {
			items := inner.Elem().FieldByName("items")
			if items.IsValid() && items.Kind() == reflect.Map {
				for _, kv := range items.MapKeys() {
					k := kv.String()
					if strings.HasPrefix(k, userPrefix) {
						keysToDelete = append(keysToDelete, k)
					}
				}
			}
		}
	}
	CalendarStore.Unlock()
	for _, k := range keysToDelete {
		CalendarStore.Delete(k)
	}
}

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

func sortEvents(events []domain.CalendarEvent) {
	sort.Slice(events, func(i, j int) bool {
		if events[i].ReleaseDate != events[j].ReleaseDate {
			return events[i].ReleaseDate < events[j].ReleaseDate
		}
		return events[i].Title < events[j].Title
	})
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

	// 1. Exact cache lookup
	cacheKey := cache.CreateCacheKey("CalendarEvents", userId, req.Start, req.End, string(req.Filter), string(req.Type), regionForCache)
	var cachedEvents []domain.CalendarEvent
	if cache.GetCache(CalendarStore, cacheKey, &cachedEvents) {
		slog.Debug("Calendar: Returning cached events", "key", cacheKey)
		return cachedEvents, nil
	}

	// 2. Cross-cache lookup: If "all" is cached, instantly slice for "movie" or "tv"
	if req.Type == domain.CalendarTypeMovie || req.Type == domain.CalendarTypeTV {
		allKey := cache.CreateCacheKey("CalendarEvents", userId, req.Start, req.End, string(req.Filter), string(domain.CalendarTypeAll), regionForCache)
		var allCached []domain.CalendarEvent
		if cache.GetCache(CalendarStore, allKey, &allCached) {
			targetType := "movie"
			if req.Type == domain.CalendarTypeTV {
				targetType = "tv"
			}
			filtered := make([]domain.CalendarEvent, 0, len(allCached))
			for _, ev := range allCached {
				if ev.MediaType == targetType {
					filtered = append(filtered, ev)
				}
			}
			CalendarStore.Set(cacheKey, filtered, 6*time.Hour)
			return filtered, nil
		}
	}

	// 3. Cross-cache lookup: If both "movie" and "tv" are cached, combine for "all"
	if req.Type == domain.CalendarTypeAll {
		movieKey := cache.CreateCacheKey("CalendarEvents", userId, req.Start, req.End, string(req.Filter), string(domain.CalendarTypeMovie), regionForCache)
		tvKey := cache.CreateCacheKey("CalendarEvents", userId, req.Start, req.End, string(req.Filter), string(domain.CalendarTypeTV), regionForCache)
		var movieCached, tvCached []domain.CalendarEvent
		hasMovie := cache.GetCache(CalendarStore, movieKey, &movieCached)
		hasTv := cache.GetCache(CalendarStore, tvKey, &tvCached)
		if hasMovie && hasTv {
			combined := make([]domain.CalendarEvent, 0, len(movieCached)+len(tvCached))
			combined = append(combined, movieCached...)
			combined = append(combined, tvCached...)
			sortEvents(combined)
			CalendarStore.Set(cacheKey, combined, 6*time.Hour)
			return combined, nil
		}
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

	// Fetch user's library items (exclude soft-deleted, dropped, and cancelled)
	var watchedItems []entity.Watched
	err = s.db.Preload("Content").
		Where("user_id = ? AND status NOT IN (?)", userId, []entity.WatchedStatus{entity.DROPPED, entity.CANCELLED}).
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

	// 2. Process Library Shows (pruned candidates + parallel worker pool + per-show caching)
	if (req.Type == domain.CalendarTypeAll || req.Type == domain.CalendarTypeTV) && s.tmdb != nil {
		var candidateShows []*entity.Watched
		for i := range watchedItems {
			w := &watchedItems[i]
			if w.Content == nil || w.Content.Type != entity.SHOW {
				continue
			}

			// Optimization: Skip shows that definitely cannot have active/new episodes
			// 1) Show status is known to be Ended or Canceled
			if w.Content.Status == "Ended" || w.Content.Status == "Canceled" {
				// If user already finished the show, it can never have new episodes
				if w.Status == entity.FINISHED {
					continue
				}
				// If it premiered > 3 months before requested start date, it is concluded
				if w.Content.ReleaseDate != nil && w.Content.ReleaseDate.Before(startDate.AddDate(0, -3, 0)) {
					continue
				}
			}

			// 2) User finished the show AND it premiered > 1 year ago (concluded series)
			if w.Status == entity.FINISHED && w.Content.ReleaseDate != nil && w.Content.ReleaseDate.Before(startDate.AddDate(-1, 0, 0)) {
				continue
			}

			// 3) On HOLD or PLANNED and premiered > 2 years ago with Ended/Canceled status
			if (w.Status == entity.HOLD || w.Status == entity.PLANNED) && w.Content.ReleaseDate != nil && w.Content.ReleaseDate.Before(startDate.AddDate(-2, 0, 0)) {
				if w.Content.Status == "Ended" || w.Content.Status == "Canceled" {
					continue
				}
			}

			candidateShows = append(candidateShows, w)
		}

		if len(candidateShows) > 0 {
			// Sort candidate shows so WATCHING shows come first, then Returning Series, then others
			sort.SliceStable(candidateShows, func(i, j int) bool {
				scoreI := 0
				if candidateShows[i].Status == entity.WATCHING {
					scoreI += 10
				}
				if candidateShows[i].Content != nil && candidateShows[i].Content.Status == "Returning Series" {
					scoreI += 5
				}

				scoreJ := 0
				if candidateShows[j].Status == entity.WATCHING {
					scoreJ += 10
				}
				if candidateShows[j].Content != nil && candidateShows[j].Content.Status == "Returning Series" {
					scoreJ += 5
				}

				return scoreI > scoreJ
			})

			var showsToFetch []*entity.Watched

			// Check per-show cache first
			for _, show := range candidateShows {
				showCacheKey := cache.CreateCacheKey("ShowCalEvents", show.Content.TmdbID, req.Start, req.End)
				var cachedShowEvents []domain.CalendarEvent
				if cache.GetCache(CalendarStore, showCacheKey, &cachedShowEvents) {
					for _, ev := range cachedShowEvents {
						ev.Status = string(show.Status) // update with current library status
						eventsMu.Lock()
						eventsMap[ev.ID] = ev
						eventsMu.Unlock()
					}
				} else {
					showsToFetch = append(showsToFetch, show)
				}
			}

			if len(showsToFetch) > 0 {
				ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
				defer cancel()

				var wg sync.WaitGroup
				sem := make(chan struct{}, 8) // Limit concurrent outbound HTTP calls to 8

				for _, show := range showsToFetch {
					select {
					case <-ctx.Done():
						slog.Warn("Calendar: Aggregation timeout reached, completing with collected events")
						break
					default:
					}

					wg.Add(1)
					sem <- struct{}{}

					go func(w *entity.Watched) {
						defer func() {
							<-sem
							wg.Done()
						}()

						select {
						case <-ctx.Done():
							return
						default:
						}

						showDetails, err := s.tmdb.ShowDetails(tmdb.ShowDetailsOptions{
							ID:             strconv.Itoa(w.Content.TmdbID),
							DontRunDBCache: true, // Avoid SQLite locking from concurrent writes
						})
						if err != nil {
							slog.Debug("Calendar: Failed to fetch show details from TMDB", "tmdbId", w.Content.TmdbID, "error", err)
							return
						}

						// Asynchronously sync Content.Status in DB if known and changed
						if showDetails.Status != "" && (w.Content.Status != string(showDetails.Status)) {
							go func(contentID int, newStatus string) {
								_ = s.db.Model(&entity.Content{}).Where("id = ?", contentID).Update("status", newStatus).Error
							}(w.Content.ID, string(showDetails.Status))
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
							latestSeasonIdx := -1
							for idx := len(showDetails.Seasons) - 1; idx >= 0; idx-- {
								if showDetails.Seasons[idx].SeasonNumber > 0 && showDetails.Seasons[idx].AirDate != "" {
									latestSeasonIdx = idx
									break
								}
							}
							if latestSeasonIdx >= 0 {
								latestSeason := showDetails.Seasons[latestSeasonIdx]
								sAirDate, err := time.Parse("2006-01-02", latestSeason.AirDate)
								seasonEligible := false
								if err == nil {
									// Seasons typically air within 2 months of their start date
									if !sAirDate.Before(startDate.AddDate(0, -2, 0)) && !sAirDate.After(endDate.AddDate(0, 1, 0)) {
										seasonEligible = true
									}
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

						// Cache the show's events (even if empty) for 12 hours so future calls don't re-query TMDB
						showCacheKey := cache.CreateCacheKey("ShowCalEvents", w.Content.TmdbID, req.Start, req.End)
						CalendarStore.Set(showCacheKey, showEvents, time.Hour*12)

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

	sortEvents(events)

	// Save to main request cache (6 hours for high performance across user sessions)
	CalendarStore.Set(cacheKey, events, 6*time.Hour)

	// Pre-populate movie and tv caches if this was an "all" request
	if req.Type == domain.CalendarTypeAll {
		movieKey := cache.CreateCacheKey("CalendarEvents", userId, req.Start, req.End, string(req.Filter), string(domain.CalendarTypeMovie), regionForCache)
		tvKey := cache.CreateCacheKey("CalendarEvents", userId, req.Start, req.End, string(req.Filter), string(domain.CalendarTypeTV), regionForCache)
		movieSlice := make([]domain.CalendarEvent, 0)
		tvSlice := make([]domain.CalendarEvent, 0)
		for _, ev := range events {
			if ev.MediaType == "movie" {
				movieSlice = append(movieSlice, ev)
			} else if ev.MediaType == "tv" {
				tvSlice = append(tvSlice, ev)
			}
		}
		CalendarStore.Set(movieKey, movieSlice, 6*time.Hour)
		CalendarStore.Set(tvKey, tvSlice, 6*time.Hour)
	}

	return events, nil
}
