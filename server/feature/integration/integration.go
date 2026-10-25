package integration

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	gocache "github.com/robfig/go-cache"
	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/calendar"
	"github.com/sbondCo/Watcharr/media/tmdb"
	"gorm.io/gorm"
)

var (
	IntegrationStore       = gocache.New(time.Minute*2, time.Minute*1)
	timeOfNewLoginRequired = time.Date(2023, 8, 18, 20, 30, 0, 0, time.UTC)
)

type NextUpItem struct {
	Title        string `json:"title"`
	Season       int    `json:"season"`
	Episode      int    `json:"episode"`
	EpisodeTitle string `json:"episodeTitle,omitempty"`
	AirDate      string `json:"airDate,omitempty"`
	Poster       string `json:"poster,omitempty"`
}

type WatchlistItem struct {
	ID          uint    `json:"id"`
	TmdbID      int     `json:"tmdbId"`
	Title       string  `json:"title"`
	Type        string  `json:"type"` // "movie" or "tv"
	Poster      string  `json:"poster"`
	ReleaseDate string  `json:"releaseDate,omitempty"`
	Rating      float64 `json:"rating,omitempty"`
	Overview    string  `json:"overview,omitempty"`
}

type UpcomingItem struct {
	Title        string `json:"title"`
	Type         string `json:"type"`
	Season       int    `json:"season,omitempty"`
	Episode      int    `json:"episode,omitempty"`
	EpisodeTitle string `json:"episodeTitle,omitempty"`
	AirDate      string `json:"airDate"`
	ReleaseDate  string `json:"releaseDate"`
	Poster       string `json:"poster,omitempty"`
	Overview     string `json:"overview,omitempty"`
}

type LibraryItem struct {
	ID              uint     `json:"id"`
	TmdbID          int      `json:"tmdbId"`
	Title           string   `json:"title"`
	Type            string   `json:"type"` // "movie" or "tv"
	Status          string   `json:"status"`
	Rating          float64  `json:"rating"`
	Poster          string   `json:"poster"`
	ReleaseDate     string   `json:"releaseDate,omitempty"`
	Genres          []string `json:"genres,omitempty"`
	WatchedEpisodes int      `json:"watchedEpisodes,omitempty"`
	TotalEpisodes   int      `json:"totalEpisodes,omitempty"`
	UpdatedAt       string   `json:"updatedAt"`
}

type LibraryResponse struct {
	Total       int           `json:"total"`
	MoviesCount int           `json:"moviesCount"`
	ShowsCount  int           `json:"showsCount"`
	Library     []LibraryItem `json:"library"`
}

type HomepageResponse struct {
	WatchedMovies     int32           `json:"watchedMovies"`
	WatchedEpisodes   int32           `json:"watchedEpisodes"`
	ShowsWatched      int32           `json:"showsWatched"`
	ShowsWatching     int32           `json:"showsWatching"`
	GamesPlayed       int32           `json:"gamesPlayed"`
	GamesPlaying      int32           `json:"gamesPlaying"`
	HoursWatched      int32           `json:"hoursWatched"`
	HoursPlayed       int32           `json:"hoursPlayed"`
	WatchlistCount    int32           `json:"watchlistCount"`
	UpcomingNext7Days int             `json:"upcomingNext7Days"`
	NextUp            *NextUpItem     `json:"nextUp,omitempty"`
	CurrentlyWatching []NextUpItem    `json:"currentlyWatching,omitempty"`
	Watchlist         []WatchlistItem `json:"watchlist,omitempty"`
	Upcoming          []UpcomingItem  `json:"upcoming,omitempty"`
	TopGenres         []string        `json:"topGenres,omitempty"`
}

type HASensorAttributes struct {
	WatchedMovies     int32           `json:"watchedMovies"`
	WatchedEpisodes   int32           `json:"watchedEpisodes"`
	ShowsWatched      int32           `json:"showsWatched"`
	ShowsWatching     int32           `json:"showsWatching"`
	GamesPlayed       int32           `json:"gamesPlayed"`
	GamesPlaying      int32           `json:"gamesPlaying"`
	HoursWatched      int32           `json:"hoursWatched"`
	HoursPlayed       int32           `json:"hoursPlayed"`
	WatchlistCount    int32           `json:"watchlistCount"`
	LastWatched       string          `json:"lastWatched,omitempty"`
	TodayReleases     int             `json:"todayReleases"`
	UpcomingNext7Days int             `json:"upcomingNext7Days"`
	NextUpTitle       string          `json:"nextUpTitle,omitempty"`
	NextUpSeason      int             `json:"nextUpSeason,omitempty"`
	NextUpEpisode     int             `json:"nextUpEpisode,omitempty"`
	NextUpAirDate     string          `json:"nextUpAirDate,omitempty"`
	CurrentlyWatching []NextUpItem    `json:"currentlyWatching,omitempty"`
	Watchlist         []WatchlistItem `json:"watchlist,omitempty"`
	Upcoming          []UpcomingItem  `json:"upcoming,omitempty"`
	TopGenres         []string        `json:"topGenres,omitempty"`
}

type HASensorResponse struct {
	State      string             `json:"state"`
	Attributes HASensorAttributes `json:"attributes"`
}

type Service struct {
	db       *gorm.DB
	cfg      *config.ServerConfig
	tmdb     *tmdb.TMDB
	calendar *calendar.Service
}

func NewService(
	db *gorm.DB,
	cfg *config.ServerConfig,
	tmdb *tmdb.TMDB,
	calendar *calendar.Service,
) *Service {
	return &Service{
		db:       db,
		cfg:      cfg,
		tmdb:     tmdb,
		calendar: calendar,
	}
}

// GenerateRandomApiKey creates a secure 32-character API key
func GenerateRandomApiKey() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return "wchr_" + hex.EncodeToString(b)
}

// GetOrCreateApiKey returns the existing API key or generates a new one
func (s *Service) GetOrCreateApiKey(userId uint) (string, error) {
	var user entity.User
	if err := s.db.Where("id = ?", userId).First(&user).Error; err != nil {
		return "", err
	}
	if user.ApiKey != "" {
		return user.ApiKey, nil
	}
	newKey := GenerateRandomApiKey()
	if err := s.db.Model(&entity.User{}).Where("id = ?", userId).Update("api_key", newKey).Error; err != nil {
		return "", err
	}
	return newKey, nil
}

// RegenerateApiKey generates a new API key for the user
func (s *Service) RegenerateApiKey(userId uint) (string, error) {
	newKey := GenerateRandomApiKey()
	if err := s.db.Model(&entity.User{}).Where("id = ?", userId).Update("api_key", newKey).Error; err != nil {
		return "", err
	}
	return newKey, nil
}

// Authenticate resolves a JWT token or static API key to a User ID
func (s *Service) Authenticate(tokenStr string) (uint, error) {
	tokenStr = strings.TrimSpace(tokenStr)
	if tokenStr == "" {
		return 0, errors.New("token vazio")
	}

	// 1. Check if token is a valid JWT (with strict HMAC signing method validation)
	if !strings.HasPrefix(tokenStr, "wchr_") {
		claims := &entity.TokenClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(s.cfg.JWT_SECRET), nil
		})
		if err == nil && token.Valid && claims.UserID > 0 {
			if claims.IssuedAt != nil && claims.IssuedAt.Before(timeOfNewLoginRequired) {
				return 0, errors.New("token expirado")
			}
			return claims.UserID, nil
		}
	}

	// 2. Check if token matches a static API key in the database
	var user entity.User
	if err := s.db.Where("api_key = ? AND api_key != ''", tokenStr).Select("id, api_key").First(&user).Error; err == nil && user.ID > 0 {
		if subtle.ConstantTimeCompare([]byte(user.ApiKey), []byte(tokenStr)) == 1 {
			return user.ID, nil
		}
	}

	return 0, errors.New("chave de API ou token inválido")
}

// GetHomepageData calculates homelab statistics with in-memory caching
func (s *Service) GetHomepageData(userId uint) (HomepageResponse, error) {
	cacheKey := fmt.Sprintf("homepage_%d", userId)
	if val, found := IntegrationStore.Get(cacheKey); found {
		if resp, ok := val.(HomepageResponse); ok {
			return resp, nil
		}
	}

	var watched []entity.Watched
	err := s.db.Model(&entity.Watched{}).
		Preload("Content").
		Preload("Game").
		Preload("WatchedEpisodes").
		Preload("WatchedSeasons").
		Where("user_id = ?", userId).
		Order("updated_at DESC").
		Find(&watched).Error
	if err != nil {
		return HomepageResponse{}, err
	}

	var (
		watchedMovies     int32
		watchedEpisodes   int32
		showsWatched      int32
		showsWatching     int32
		gamesPlayed       int32
		gamesPlaying      int32
		totalMinutes      uint32
		totalGameMinutes  uint32
		watchlistCount    int32
		nextUp            *NextUpItem
		currentlyWatching []NextUpItem
		watchlistItems    []WatchlistItem
		genreCounts       = make(map[string]int)
	)

	for _, w := range watched {
		// Watchlist
		if w.Status == entity.PLANNED {
			watchlistCount++
			if w.Content != nil && len(watchlistItems) < 30 {
				var relDate string
				if w.Content.ReleaseDate != nil {
					relDate = w.Content.ReleaseDate.Format("2006-01-02")
				}
				watchlistItems = append(watchlistItems, WatchlistItem{
					ID:          w.ID,
					TmdbID:      w.Content.TmdbID,
					Title:       w.Content.Title,
					Type:        string(w.Content.Type),
					Poster:      buildPosterURL(w.Content.PosterPath),
					ReleaseDate: relDate,
					Rating:      w.Rating,
					Overview:    w.Content.Overview,
				})
			}
		}

		if w.Game != nil {
			if w.Status == entity.WATCHING {
				gamesPlaying++
			} else if w.Status == entity.FINISHED {
				gamesPlayed++
				if w.Game.TimeToBeatNormal > 0 {
					totalGameMinutes += uint32(w.Game.TimeToBeatNormal / 60)
				} else {
					totalGameMinutes += 15 * 60
				}
			}
			continue
		}

		if w.Content == nil {
			continue
		}

		if w.Status == entity.FINISHED || w.Status == entity.WATCHING {
			for _, g := range strings.Split(w.Content.Genres, ",") {
				g = strings.TrimSpace(g)
				if g != "" {
					genreCounts[g]++
				}
			}
		}

		if w.Content.Type == entity.MOVIE {
			if w.Status == entity.FINISHED {
				watchedMovies++
				rt := w.Content.Runtime
				if rt == 0 {
					rt = 110
				}
				totalMinutes += rt
			}
		} else if w.Content.Type == entity.SHOW {
			if w.Status == entity.FINISHED {
				showsWatched++
			} else if w.Status == entity.WATCHING {
				showsWatching++
			}

			effRuntime := w.Content.Runtime
			if effRuntime == 0 {
				effRuntime = 42
			}

			episodesDone := 0
			var highestSeason, highestEpisode int

			for _, we := range w.WatchedEpisodes {
				if we.Status == entity.FINISHED {
					episodesDone++
					if we.SeasonNumber > highestSeason || (we.SeasonNumber == highestSeason && we.EpisodeNumber > highestEpisode) {
						highestSeason = we.SeasonNumber
						highestEpisode = we.EpisodeNumber
					}
				}
			}

			if episodesDone > 0 {
				watchedEpisodes += int32(episodesDone)
				totalMinutes += uint32(episodesDone) * effRuntime
			} else if w.Status == entity.FINISHED {
				epCount := w.Content.NumberOfEpisodes
				if epCount == 0 {
					epCount = 10
				}
				watchedEpisodes += int32(epCount)
				totalMinutes += epCount * effRuntime
			} else if len(w.WatchedSeasons) > 0 {
				finishedSeasons := 0
				for _, ws := range w.WatchedSeasons {
					if ws.Status == entity.FINISHED {
						finishedSeasons++
						if ws.SeasonNumber > highestSeason {
							highestSeason = ws.SeasonNumber
						}
					}
				}
				if finishedSeasons > 0 {
					epsPerSeason := uint32(10)
					if w.Content.NumberOfSeasons > 0 && w.Content.NumberOfEpisodes > 0 {
						epsPerSeason = w.Content.NumberOfEpisodes / w.Content.NumberOfSeasons
						if epsPerSeason == 0 {
							epsPerSeason = 8
						}
					}
					estEps := uint32(finishedSeasons) * epsPerSeason
					watchedEpisodes += int32(estEps)
					totalMinutes += estEps * effRuntime
				}
			}

			// Find next episode to watch for currently watching shows
			if w.Status == entity.WATCHING {
				targetSeason := 1
				targetEpisode := 1
				if highestSeason > 0 {
					targetSeason = highestSeason
					targetEpisode = highestEpisode + 1
				}

				item := NextUpItem{
					Title:   w.Content.Title,
					Season:  targetSeason,
					Episode: targetEpisode,
					Poster:  buildPosterURL(w.Content.PosterPath),
				}
				if nextUp == nil {
					nextUp = &item
				}
				if len(currentlyWatching) < 20 {
					currentlyWatching = append(currentlyWatching, item)
				}
			}
		}
	}

	var topGenres []string
	for len(topGenres) < 3 && len(genreCounts) > 0 {
		bestG := ""
		bestC := 0
		for g, c := range genreCounts {
			if c > bestC {
				bestC = c
				bestG = g
			}
		}
		if bestG == "" {
			break
		}
		topGenres = append(topGenres, bestG)
		delete(genreCounts, bestG)
	}

	hoursWatched := int32(totalMinutes / 60)
	hoursPlayed := int32(totalGameMinutes / 60)

	upcomingNext7Days := 0
	var upcomingItems []UpcomingItem
	if s.calendar != nil {
		now := time.Now()
		events, err := s.calendar.GetEvents(userId, domain.CalendarRequest{
			Start:  now.Format("2006-01-02"),
			End:    now.AddDate(0, 0, 14).Format("2006-01-02"),
			Filter: domain.CalendarFilterLibrary,
		})
		if err == nil {
			next7Str := now.AddDate(0, 0, 7).Format("2006-01-02")
			for i, ev := range events {
				if ev.ReleaseDate <= next7Str {
					upcomingNext7Days++
				}
				if i < 20 {
					upcomingItems = append(upcomingItems, UpcomingItem{
						Title:        ev.Title,
						Type:         ev.MediaType,
						Season:       ev.SeasonNumber,
						Episode:      ev.EpisodeNumber,
						EpisodeTitle: ev.EpisodeName,
						AirDate:      ev.ReleaseDate,
						ReleaseDate:  ev.ReleaseDate,
						Poster:       buildPosterURL(ev.PosterPath),
						Overview:     ev.Overview,
					})
				}
			}
		}
	}

	resp := HomepageResponse{
		WatchedMovies:     watchedMovies,
		WatchedEpisodes:   watchedEpisodes,
		ShowsWatched:      showsWatched,
		ShowsWatching:     showsWatching,
		GamesPlayed:       gamesPlayed,
		GamesPlaying:      gamesPlaying,
		HoursWatched:      hoursWatched,
		HoursPlayed:       hoursPlayed,
		WatchlistCount:    watchlistCount,
		UpcomingNext7Days: upcomingNext7Days,
		NextUp:            nextUp,
		CurrentlyWatching: currentlyWatching,
		Watchlist:         watchlistItems,
		Upcoming:          upcomingItems,
		TopGenres:         topGenres,
	}

	IntegrationStore.Set(cacheKey, resp, time.Second*90)
	return resp, nil
}

// GetHomeAssistantSensor returns Home Assistant formatted sensor data
func (s *Service) GetHomeAssistantSensor(userId uint) (HASensorResponse, error) {
	data, err := s.GetHomepageData(userId)
	if err != nil {
		return HASensorResponse{}, err
	}

	state := "Pronto"
	var lastWatched string
	if data.NextUp != nil {
		state = fmt.Sprintf("%s S%02dE%02d", data.NextUp.Title, data.NextUp.Season, data.NextUp.Episode)
	} else if data.WatchedMovies > 0 || data.WatchedEpisodes > 0 {
		state = fmt.Sprintf("%d Filmes, %d Eps", data.WatchedMovies, data.WatchedEpisodes)
	}

	// Find last watched item title
	var lastW entity.Watched
	if err := s.db.Model(&entity.Watched{}).Preload("Content").Where("user_id = ? AND status = ?", userId, entity.FINISHED).Order("updated_at DESC").First(&lastW).Error; err == nil && lastW.Content != nil {
		lastWatched = lastW.Content.Title
	}

	// Calculate today releases count
	todayStr := time.Now().Format("2006-01-02")
	todayReleases := 0
	if s.calendar != nil {
		events, err := s.calendar.GetEvents(userId, domain.CalendarRequest{
			Start:  todayStr,
			End:    todayStr,
			Filter: domain.CalendarFilterLibrary,
		})
		if err == nil {
			todayReleases = len(events)
		}
	}

	attrs := HASensorAttributes{
		WatchedMovies:     data.WatchedMovies,
		WatchedEpisodes:   data.WatchedEpisodes,
		ShowsWatched:      data.ShowsWatched,
		ShowsWatching:     data.ShowsWatching,
		GamesPlayed:       data.GamesPlayed,
		GamesPlaying:      data.GamesPlaying,
		HoursWatched:      data.HoursWatched,
		HoursPlayed:       data.HoursPlayed,
		WatchlistCount:    data.WatchlistCount,
		LastWatched:       lastWatched,
		TodayReleases:     todayReleases,
		UpcomingNext7Days: data.UpcomingNext7Days,
		CurrentlyWatching: data.CurrentlyWatching,
		Watchlist:         data.Watchlist,
		Upcoming:          data.Upcoming,
		TopGenres:         data.TopGenres,
	}

	if data.NextUp != nil {
		attrs.NextUpTitle = data.NextUp.Title
		attrs.NextUpSeason = data.NextUp.Season
		attrs.NextUpEpisode = data.NextUp.Episode
		attrs.NextUpAirDate = data.NextUp.AirDate
	}

	return HASensorResponse{
		State:      state,
		Attributes: attrs,
	}, nil
}

// GetLibraryData returns the complete media library formatted for Home Assistant
func (s *Service) GetLibraryData(userId uint) (LibraryResponse, error) {
	var watched []entity.Watched
	err := s.db.Model(&entity.Watched{}).
		Preload("Content").
		Preload("Game").
		Preload("WatchedEpisodes").
		Where("user_id = ?", userId).
		Order("updated_at DESC").
		Find(&watched).Error
	if err != nil {
		return LibraryResponse{}, err
	}

	var moviesCount, showsCount int
	var items []LibraryItem

	for _, w := range watched {
		if w.Content == nil {
			continue
		}
		var genres []string
		if w.Content.Genres != "" {
			for _, g := range strings.Split(w.Content.Genres, ",") {
				g = strings.TrimSpace(g)
				if g != "" {
					genres = append(genres, g)
				}
			}
		}

		var relDate string
		if w.Content.ReleaseDate != nil {
			relDate = w.Content.ReleaseDate.Format("2006-01-02")
		}

		watchedEps := 0
		for _, we := range w.WatchedEpisodes {
			if we.Status == entity.FINISHED {
				watchedEps++
			}
		}

		if w.Content.Type == entity.MOVIE {
			moviesCount++
		} else if w.Content.Type == entity.SHOW {
			showsCount++
		}

		items = append(items, LibraryItem{
			ID:              w.ID,
			TmdbID:          w.Content.TmdbID,
			Title:           w.Content.Title,
			Type:            string(w.Content.Type),
			Status:          string(w.Status),
			Rating:          w.Rating,
			Poster:          buildPosterURL(w.Content.PosterPath),
			ReleaseDate:     relDate,
			Genres:          genres,
			WatchedEpisodes: watchedEps,
			TotalEpisodes:   int(w.Content.NumberOfEpisodes),
			UpdatedAt:       w.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return LibraryResponse{
		Total:       len(items),
		MoviesCount: moviesCount,
		ShowsCount:  showsCount,
		Library:     items,
	}, nil
}

// GetUpcomingData returns upcoming releases formatted for Home Assistant
func (s *Service) GetUpcomingData(userId uint) ([]UpcomingItem, error) {
	if s.calendar == nil {
		return []UpcomingItem{}, nil
	}
	now := time.Now()
	events, err := s.calendar.GetEvents(userId, domain.CalendarRequest{
		Start:  now.Format("2006-01-02"),
		End:    now.AddDate(0, 0, 30).Format("2006-01-02"),
		Filter: domain.CalendarFilterLibrary,
	})
	if err != nil {
		return nil, err
	}

	var items []UpcomingItem
	for _, ev := range events {
		items = append(items, UpcomingItem{
			Title:        ev.Title,
			Type:         ev.MediaType,
			Season:       ev.SeasonNumber,
			Episode:      ev.EpisodeNumber,
			EpisodeTitle: ev.EpisodeName,
			AirDate:      ev.ReleaseDate,
			ReleaseDate:  ev.ReleaseDate,
			Poster:       buildPosterURL(ev.PosterPath),
			Overview:     ev.Overview,
		})
	}
	return items, nil
}

// GetCalendarICS generates standard RFC 5545 iCalendar feed
func (s *Service) GetCalendarICS(userId uint) (string, error) {
	now := time.Now()
	startStr := now.AddDate(0, 0, -3).Format("2006-01-02")
	endStr := now.AddDate(0, 0, 30).Format("2006-01-02")

	events, err := s.calendar.GetEvents(userId, domain.CalendarRequest{
		Start:  startStr,
		End:    endStr,
		Filter: domain.CalendarFilterLibrary,
		Type:   domain.CalendarTypeAll,
	})
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString("BEGIN:VCALENDAR\r\n")
	sb.WriteString("VERSION:2.0\r\n")
	sb.WriteString("PRODID:-//Watcharr//Dashboard Calendar//PT\r\n")
	sb.WriteString("CALSCALE:GREGORIAN\r\n")
	sb.WriteString("METHOD:PUBLISH\r\n")
	sb.WriteString("X-WR-CALNAME:Watcharr Estreias\r\n")
	sb.WriteString("X-WR-TIMEZONE:UTC\r\n")

	for _, ev := range events {
		cleanDate := strings.ReplaceAll(ev.ReleaseDate, "-", "")
		if len(cleanDate) != 8 {
			continue
		}

		summary := ev.Title
		if ev.MediaType == "tv" && ev.SeasonNumber > 0 {
			summary = fmt.Sprintf("%s - S%02dE%02d", ev.Title, ev.SeasonNumber, ev.EpisodeNumber)
			if ev.EpisodeName != "" {
				summary += fmt.Sprintf(" (%s)", ev.EpisodeName)
			}
		}

		posterURL := buildPosterURL(ev.PosterPath)

		desc := ev.Overview
		if desc == "" {
			desc = "Estreia de conteúdo da tua biblioteca Watcharr."
		}
		if posterURL != "" {
			desc += "\nPoster: " + posterURL
		}
		desc = strings.ReplaceAll(desc, "\r", "")
		desc = strings.ReplaceAll(desc, "\n", "\\n")
		desc = strings.ReplaceAll(desc, ",", "\\,")

		sb.WriteString("BEGIN:VEVENT\r\n")
		sb.WriteString(fmt.Sprintf("UID:watcharr-%s-%s@homelab\r\n", ev.ID, cleanDate))
		sb.WriteString(fmt.Sprintf("DTSTAMP:%sT000000Z\r\n", time.Now().Format("20060102")))
		sb.WriteString(fmt.Sprintf("DTSTART;VALUE=DATE:%s\r\n", cleanDate))
		sb.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", summary))
		sb.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", desc))
		sb.WriteString("STATUS:CONFIRMED\r\n")
		if posterURL != "" {
			sb.WriteString(fmt.Sprintf("ATTACH;FMTTYPE=image/jpeg:%s\r\n", posterURL))
			sb.WriteString(fmt.Sprintf("X-POSTER:%s\r\n", posterURL))
		}
		sb.WriteString("END:VEVENT\r\n")
	}

	sb.WriteString("END:VCALENDAR\r\n")
	return sb.String(), nil
}

func buildPosterURL(posterPath string) string {
	posterPath = strings.TrimSpace(posterPath)
	if posterPath == "" {
		return ""
	}
	if strings.HasPrefix(posterPath, "http://") || strings.HasPrefix(posterPath, "https://") {
		return posterPath
	}
	if strings.HasPrefix(posterPath, "/") {
		return "https://image.tmdb.org/t/p/w500" + posterPath
	}
	return "https://image.tmdb.org/t/p/w500/" + posterPath
}

// GetNowNextSVG renders a vector dashboard card for Home Assistant Lovelace / Wallpanels
func (s *Service) GetNowNextSVG(userId uint) (string, error) {
	data, err := s.GetHomepageData(userId)
	if err != nil {
		return "", err
	}

	title := "Tudo em Dia"
	sub := "Nenhum episódio pendente na tua lista de séries ativas."
	pill := "TUDO VISTO"

	if data.NextUp != nil {
		title = html.EscapeString(data.NextUp.Title)
		sub = fmt.Sprintf("Temporada %d • Episódio %d", data.NextUp.Season, data.NextUp.Episode)
		pill = "A SEGUIR"
	}

	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="600" height="280" viewBox="0 0 600 280">
	<defs>
		<linearGradient id="bg" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
			<stop offset="0%%" stop-color="#0f172a" />
			<stop offset="50%%" stop-color="#1e1b4b" />
			<stop offset="100%%" stop-color="#090d16" />
		</linearGradient>
		<linearGradient id="accent" x1="0%%" y1="0%%" x2="100%%" y2="0%%">
			<stop offset="0%%" stop-color="#6366f1" />
			<stop offset="100%%" stop-color="#a855f7" />
		</linearGradient>
		<filter id="glow" x="-20%%" y="-20%%" width="140%%" height="140%%">
			<feGaussianBlur stdDeviation="8" result="blur" />
			<feComposite in="SourceGraphic" in2="blur" operator="over" />
		</filter>
	</defs>

	<!-- Card Background -->
	<rect width="600" height="280" rx="20" fill="url(#bg)" stroke="#312e81" stroke-width="1.5" />
	
	<!-- Ambient Glow Orb -->
	<circle cx="530" cy="50" r="90" fill="#6366f1" opacity="0.12" filter="url(#glow)" />
	<circle cx="80" cy="230" r="70" fill="#a855f7" opacity="0.10" filter="url(#glow)" />

	<!-- Header Bar -->
	<g transform="translate(32, 36)">
		<!-- Pill -->
		<rect width="90" height="26" rx="13" fill="#6366f1" opacity="0.25" stroke="#818cf8" stroke-width="1" />
		<text x="45" y="17" fill="#c7d2fe" font-family="-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif" font-size="11" font-weight="700" text-anchor="middle" letter-spacing="1">%s</text>
		
		<!-- Branding -->
		<text x="536" y="17" fill="#64748b" font-family="-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif" font-size="12" font-weight="600" text-anchor="end" letter-spacing="1.5">WATCHARR</text>
	</g>

	<!-- Main Title & Info -->
	<g transform="translate(32, 115)">
		<text x="0" y="0" fill="#ffffff" font-family="-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif" font-size="30" font-weight="800" letter-spacing="-0.5">%s</text>
		<text x="0" y="32" fill="#94a3b8" font-family="-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif" font-size="16" font-weight="500">%s</text>
	</g>

	<!-- Stats Divider -->
	<line x1="32" y1="185" x2="568" y2="185" stroke="#334155" stroke-opacity="0.6" stroke-dasharray="4 4" />

	<!-- Bottom Metrics Grid -->
	<g transform="translate(32, 225)" font-family="-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif">
		<!-- Movies -->
		<text x="0" y="0" fill="#64748b" font-size="11" font-weight="700" letter-spacing="0.8">FILMES</text>
		<text x="0" y="22" fill="#f8fafc" font-size="18" font-weight="700">%d</text>

		<!-- Episodes -->
		<text x="135" y="0" fill="#64748b" font-size="11" font-weight="700" letter-spacing="0.8">EPISÓDIOS</text>
		<text x="135" y="22" fill="#f8fafc" font-size="18" font-weight="700">%d</text>

		<!-- Hours -->
		<text x="275" y="0" fill="#64748b" font-size="11" font-weight="700" letter-spacing="0.8">HORAS VISTAS</text>
		<text x="275" y="22" fill="#a78bfa" font-size="18" font-weight="700">%dh</text>

		<!-- Watchlist -->
		<text x="420" y="0" fill="#64748b" font-size="11" font-weight="700" letter-spacing="0.8">NA LISTA</text>
		<text x="420" y="22" fill="#38bdf8" font-size="18" font-weight="700">%d</text>
	</g>
</svg>`, pill, title, sub, data.WatchedMovies, data.WatchedEpisodes, data.HoursWatched, data.WatchlistCount)

	return svg, nil
}
