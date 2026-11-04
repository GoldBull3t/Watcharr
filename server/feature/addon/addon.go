package addon

import (
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/integration"
	"github.com/sbondCo/Watcharr/media/tmdb"
	"github.com/sbondCo/Watcharr/util"
	"gorm.io/gorm"
)

type WatchedService interface {
	AddWatched(userId uint, ar domain.WatchedAddRequest, extraProps domain.WatchedAddExtraProps) (entity.Watched, error)
	UpdateWatched(userId uint, id uint, ar domain.WatchedUpdateRequest) (domain.WatchedUpdateResponse, error)
	GetWatchedItemById(userId uint, id uint) (entity.Watched, error)
}

type EpisodeService interface {
	SetWatchedEpisode(userId uint, ar domain.WatchedEpisodeSetRequest) (domain.WatchedEpisodeSetResponse, error)
}

type Service struct {
	db             *gorm.DB
	cfg            *config.ServerConfig
	tmdb           *tmdb.TMDB
	watchedService WatchedService
	episodeService EpisodeService
	authService    *integration.Service
}

func NewService(
	db *gorm.DB,
	cfg *config.ServerConfig,
	tmdb *tmdb.TMDB,
	watchedService WatchedService,
	episodeService EpisodeService,
	authService *integration.Service,
) *Service {
	return &Service{
		db:             db,
		cfg:            cfg,
		tmdb:           tmdb,
		watchedService: watchedService,
		episodeService: episodeService,
		authService:    authService,
	}
}

func (s *Service) Authenticate(token string) (uint, error) {
	return s.authService.Authenticate(token)
}

func (s *Service) GetManifest() ManifestResponse {
	return ManifestResponse{
		ID:          "com.watcharr.nuvio.addon",
		Name:        "Watcharr",
		Version:     "1.1.0",
		Description: "Acede à tua biblioteca pessoal, séries a ver e filmes planeados do Watcharr diretamente na Android TV.",
		Logo:        "https://raw.githubusercontent.com/sbondCo/Watcharr/main/docs/static/img/watcharr_logo.png",
		Background:  "https://raw.githubusercontent.com/sbondCo/Watcharr/main/docs/static/img/preview.png",
		Resources:   []string{"catalog", "meta"},
		Types:       []string{"movie", "series"},
		IDPrefixes:  []string{"tmdb:", "tt"},
		BehaviorHints: &BehaviorHints{
			Configurable:          false,
			ConfigurationRequired: false,
		},
		Catalogs: []CatalogDescriptor{
			{
				Type:           "series",
				ID:             "watcharr_watching_series",
				Name:           "Watcharr: Séries A Ver",
				Extra:          []Extra{{Name: "skip"}},
				ExtraSupported: []string{"skip"},
				Genres:         []string{},
			},
			{
				Type:           "movie",
				ID:             "watcharr_watching_movies",
				Name:           "Watcharr: Filmes A Ver",
				Extra:          []Extra{{Name: "skip"}},
				ExtraSupported: []string{"skip"},
				Genres:         []string{},
			},
			{
				Type:           "series",
				ID:             "watcharr_planned_series",
				Name:           "Watcharr: Séries Para Ver",
				Extra:          []Extra{{Name: "skip"}},
				ExtraSupported: []string{"skip"},
				Genres:         []string{},
			},
			{
				Type:           "movie",
				ID:             "watcharr_planned_movies",
				Name:           "Watcharr: Filmes Para Ver",
				Extra:          []Extra{{Name: "skip"}},
				ExtraSupported: []string{"skip"},
				Genres:         []string{},
			},
			{
				Type:           "series",
				ID:             "watcharr_history_series",
				Name:           "Watcharr: Histórico Séries",
				Extra:          []Extra{{Name: "skip"}},
				ExtraSupported: []string{"skip"},
				Genres:         []string{},
			},
			{
				Type:           "movie",
				ID:             "watcharr_history_movies",
				Name:           "Watcharr: Histórico Filmes",
				Extra:          []Extra{{Name: "skip"}},
				ExtraSupported: []string{"skip"},
				Genres:         []string{},
			},
		},
	}
}

func (s *Service) GetCatalog(userId uint, mediaType string, catalogId string) (CatalogResponse, error) {
	var status entity.WatchedStatus
	var targetContentType entity.ContentType

	if mediaType == "movie" {
		targetContentType = entity.MOVIE
	} else if mediaType == "series" || mediaType == "tv" || mediaType == "show" {
		targetContentType = entity.SHOW
	} else {
		return CatalogResponse{}, errors.New("unsupported type")
	}

	switch catalogId {
	case "watcharr_watching_series", "watcharr_watching_movies":
		status = entity.WATCHING
	case "watcharr_planned_series", "watcharr_planned_movies":
		status = entity.PLANNED
	case "watcharr_history_series", "watcharr_history_movies":
		status = entity.FINISHED
	default:
		return CatalogResponse{}, fmt.Errorf("unknown catalog id: %s", catalogId)
	}

	var items []entity.Watched
	err := s.db.Model(&entity.Watched{}).
		Preload("Content").
		Where("user_id = ? AND status = ?", userId, status).
		Order("updated_at DESC").
		Limit(50).
		Find(&items).Error
	if err != nil {
		return CatalogResponse{}, err
	}

	metas := make([]MetaPreview, 0)
	for _, item := range items {
		if item.Content == nil || item.Content.Type != targetContentType {
			continue
		}

		c := item.Content
		poster := integration.BuildPosterURL(c.PosterPath)

		var relYear string
		if c.ReleaseDate != nil {
			relYear = strconv.Itoa(c.ReleaseDate.Year())
		}

		var genres []string
		if c.Genres != "" {
			for _, g := range strings.Split(c.Genres, ",") {
				g = strings.TrimSpace(g)
				if g != "" {
					genres = append(genres, g)
				}
			}
		}

		stremioId := fmt.Sprintf("tmdb:%d", c.TmdbID)
		if c.ImdbID != "" {
			stremioId = c.ImdbID
		}

		metas = append(metas, MetaPreview{
			ID:          stremioId,
			Type:        mediaType,
			Name:        c.Title,
			Poster:      poster,
			PosterShape: "poster",
			Description: c.Overview,
			ReleaseInfo: relYear,
			ImdbRating:  fmt.Sprintf("%.1f", c.VoteAverage),
			Genres:      genres,
		})
	}

	return CatalogResponse{Metas: metas}, nil
}

func (s *Service) GetMeta(userId uint, mediaType string, id string) (MetaDetailResponse, error) {
	var tmdbId int
	var imdbId string

	if strings.HasPrefix(id, "tmdb:") {
		parsed, _ := strconv.Atoi(strings.TrimPrefix(id, "tmdb:"))
		tmdbId = parsed
	} else if strings.HasPrefix(id, "tt") {
		imdbId = id
	}

	var content entity.Content
	query := s.db.Model(&entity.Content{})
	if tmdbId > 0 {
		query = query.Where("tmdb_id = ?", tmdbId)
	} else if imdbId != "" {
		query = query.Where("imdb_id = ?", imdbId)
	} else {
		return MetaDetailResponse{}, errors.New("invalid id format")
	}

	err := query.First(&content).Error
	if err != nil {
		return MetaDetailResponse{}, err
	}

	poster := integration.BuildPosterURL(content.PosterPath)
	var relYear string
	if content.ReleaseDate != nil {
		relYear = strconv.Itoa(content.ReleaseDate.Year())
	}

	var genres []string
	if content.Genres != "" {
		for _, g := range strings.Split(content.Genres, ",") {
			g = strings.TrimSpace(g)
			if g != "" {
				genres = append(genres, g)
			}
		}
	}

	stremioId := id
	meta := MetaDetail{
		ID:          stremioId,
		Type:        mediaType,
		Name:        content.Title,
		Genres:      genres,
		Poster:      poster,
		Description: content.Overview,
		ReleaseInfo: relYear,
		ImdbRating:  fmt.Sprintf("%.1f", content.VoteAverage),
	}

	// If series, populate video entries
	if (mediaType == "series" || mediaType == "tv" || mediaType == "show") && content.Type == entity.SHOW {
		var videos []Video
		for sNum := 1; sNum <= int(content.NumberOfSeasons); sNum++ {
			// Estimate episodes or fallback to standard 24 if seasons not cached
			maxEps := 24
			for epNum := 1; epNum <= maxEps; epNum++ {
				epId := fmt.Sprintf("%s:%d:%d", stremioId, sNum, epNum)
				videos = append(videos, Video{
					ID:      epId,
					Title:   fmt.Sprintf("Episódio %d", epNum),
					Season:  sNum,
					Episode: epNum,
				})
			}
		}
		meta.Videos = videos
	}

	return MetaDetailResponse{Meta: meta}, nil
}

func (s *Service) HandleScrobble(userId uint, p ScrobblePayload) error {
	slog.Info("Addon Scrobble received",
		"userId", userId,
		"event", p.Event,
		"mediaType", p.MediaType,
		"tmdbId", p.TmdbID,
		"season", p.SeasonNumber,
		"episode", p.EpisodeNumber,
		"progress", p.Progress)

	// We trigger watched status when progress is >= 80% or event is "stop" with progress >= 80%
	if p.Progress < 80.0 && p.Event != "stop" {
		slog.Debug("Addon Scrobble: Progress under 80%, not marking watched yet", "progress", p.Progress)
		return nil
	}

	// 1. Resolve content
	var content entity.Content
	query := s.db.Model(&entity.Content{})
	if p.TmdbID > 0 {
		query = query.Where("tmdb_id = ?", p.TmdbID)
	} else if p.ImdbID != "" {
		query = query.Where("imdb_id = ?", p.ImdbID)
	} else {
		return errors.New("scrobble missing tmdbId and imdbId")
	}

	if err := query.First(&content).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) && p.TmdbID > 0 {
			// Content might not be cached locally yet, let's cache it via tmdb
			supportedMedia := util.SupportedMediaMovie
			if p.MediaType == "series" || p.MediaType == "tv" {
				supportedMedia = util.SupportedMediaShow
			}
			// Attempt to add watched
			_, addErr := s.watchedService.AddWatched(userId, domain.WatchedAddRequest{
				Status:      entity.FINISHED,
				ContentType: supportedMedia,
				TMDBID:      p.TmdbID,
			}, domain.WatchedAddExtraProps{})
			return addErr
		}
		return err
	}

	// 2. Find existing watched record for user
	var watched entity.Watched
	err := s.db.Model(&entity.Watched{}).
		Where("user_id = ? AND content_id = ?", userId, content.ID).
		First(&watched).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new watched item
		supportedMedia := util.SupportedMediaMovie
		status := entity.FINISHED
		if content.Type == entity.SHOW {
			supportedMedia = util.SupportedMediaShow
			status = entity.WATCHING
		}

		createdWatched, addErr := s.watchedService.AddWatched(userId, domain.WatchedAddRequest{
			Status:      status,
			ContentType: supportedMedia,
			TMDBID:      content.TmdbID,
		}, domain.WatchedAddExtraProps{})
		if addErr != nil {
			return addErr
		}
		watched = createdWatched
	}

	// 3. Mark Movie as Finished
	if content.Type == entity.MOVIE {
		_, upErr := s.watchedService.UpdateWatched(userId, watched.ID, domain.WatchedUpdateRequest{
			Status: entity.FINISHED,
		})
		return upErr
	}

	// 4. Mark Episode as Finished for TV Show
	if content.Type == entity.SHOW && p.SeasonNumber > 0 && p.EpisodeNumber > 0 {
		now := time.Now()
		_, epErr := s.episodeService.SetWatchedEpisode(userId, domain.WatchedEpisodeSetRequest{
			WatchedID:     watched.ID,
			SeasonNumber:  p.SeasonNumber,
			EpisodeNumber: p.EpisodeNumber,
			Status:        entity.FINISHED,
		})
		if epErr != nil {
			slog.Error("Failed to set watched episode from scrobble", "error", epErr)
			return epErr
		}

		// Keep series in WATCHING status and update timestamp
		_ = s.db.Model(&entity.Watched{}).
			Where("id = ?", watched.ID).
			Updates(map[string]any{
				"status":     entity.WATCHING,
				"updated_at": now,
			}).Error
	}

	return nil
}
