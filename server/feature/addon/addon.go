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
		Version:     "1.3.0",
		Description: "Acede à tua biblioteca pessoal, séries a ver e filmes planeados do Watcharr diretamente na Android TV com scrobble bidirecional ativo.",
		Logo:        "https://raw.githubusercontent.com/sbondCo/Watcharr/main/docs/static/img/watcharr_logo.png",
		Background:  "https://raw.githubusercontent.com/sbondCo/Watcharr/main/docs/static/img/preview.png",
		Resources:   []string{"catalog", "meta", "stream"},
		Types:       []string{"movie", "series"},
		IDPrefixes:  []string{"tmdb:", "tmdb", "tt"},
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
		},
	}
}

func (s *Service) GetCatalog(userId uint, mediaType string, catalogId string, skip int) (CatalogResponse, string, error) {
	var status entity.WatchedStatus
	var targetContentType entity.ContentType

	if mediaType == "movie" {
		targetContentType = entity.MOVIE
	} else if mediaType == "series" || mediaType == "tv" || mediaType == "show" {
		targetContentType = entity.SHOW
	} else {
		return CatalogResponse{}, "", errors.New("unsupported type")
	}

	switch catalogId {
	case "watcharr_watching_series", "watcharr_watching_movies":
		status = entity.WATCHING
	case "watcharr_planned_series", "watcharr_planned_movies":
		status = entity.PLANNED
	default:
		return CatalogResponse{}, "", fmt.Errorf("unknown catalog id: %s", catalogId)
	}

	var items []entity.Watched
	query := s.db.Model(&entity.Watched{}).
		Joins("INNER JOIN contents ON contents.id = watcheds.content_id").
		Preload("Content").
		Where("watcheds.user_id = ? AND watcheds.status = ? AND contents.type = ?", userId, status, targetContentType).
		Order("watcheds.updated_at DESC")

	if skip > 0 {
		query = query.Offset(skip)
	}

	err := query.Limit(50).Find(&items).Error
	if err != nil {
		return CatalogResponse{}, "", err
	}

	var maxUpdatedAt int64 = 0
	metas := make([]MetaPreview, 0)
	for _, item := range items {
		if item.UpdatedAt.Unix() > maxUpdatedAt {
			maxUpdatedAt = item.UpdatedAt.Unix()
		}

		if item.Content == nil || item.Content.Type != targetContentType {
			continue
		}

		c := item.Content
		poster := integration.BuildPosterURL(c.PosterPath)
		backdrop := integration.BuildPosterURL(c.BackdropPath)

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
			Background:  backdrop,
			Description: c.Overview,
			ReleaseInfo: relYear,
			ImdbRating:  fmt.Sprintf("%.1f", c.VoteAverage),
			Genres:      genres,
		})
	}

	etag := fmt.Sprintf("\"w-%d-%s-%d\"", maxUpdatedAt, catalogId, skip)
	return CatalogResponse{Metas: metas}, etag, nil
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

	var targetContentType entity.ContentType
	if mediaType == "movie" {
		targetContentType = entity.MOVIE
	} else if mediaType == "series" || mediaType == "tv" || mediaType == "show" {
		targetContentType = entity.SHOW
	}

	var content entity.Content
	query := s.db.Model(&entity.Content{})
	if tmdbId > 0 {
		if targetContentType != "" {
			query = query.Where("tmdb_id = ? AND type = ?", tmdbId, targetContentType)
		} else {
			query = query.Where("tmdb_id = ?", tmdbId)
		}
	} else if imdbId != "" {
		if targetContentType != "" {
			query = query.Where("imdb_id = ? AND type = ?", imdbId, targetContentType)
		} else {
			query = query.Where("imdb_id = ?", imdbId)
		}
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
		numSeasons := int(content.NumberOfSeasons)
		if numSeasons > 100 {
			numSeasons = 100
		}
		for sNum := 1; sNum <= numSeasons; sNum++ {
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

// ParseStreamID parses a Stremio / Nuvio resource ID format
// Examples:
// - tt1234567:1:2 -> imdbID: tt1234567, season: 1, episode: 2
// - tmdb:12345:1:2 -> tmdbID: 12345, season: 1, episode: 2
// - tmdb:12345 -> tmdbID: 12345
// - tt1234567 -> imdbID: tt1234567
func ParseStreamID(id string) (tmdbID int, imdbID string, seasonNum int, epNum int) {
	id = strings.TrimSuffix(id, ".json")
	parts := strings.Split(id, ":")

	if len(parts) >= 3 {
		if parts[0] == "tmdb" {
			if len(parts) == 4 {
				tmdbID, _ = strconv.Atoi(parts[1])
				seasonNum, _ = strconv.Atoi(parts[2])
				epNum, _ = strconv.Atoi(parts[3])
			} else {
				tmdbID, _ = strconv.Atoi(parts[1])
			}
		} else if strings.HasPrefix(parts[0], "tt") && len(parts) == 3 {
			imdbID = parts[0]
			seasonNum, _ = strconv.Atoi(parts[1])
			epNum, _ = strconv.Atoi(parts[2])
		}
	} else if len(parts) == 2 && parts[0] == "tmdb" {
		tmdbID, _ = strconv.Atoi(parts[1])
	} else if strings.HasPrefix(id, "tt") {
		imdbID = id
	}
	return
}

// HandleStreamPlayback intercepts playback start from Stremio / Nuvio
func (s *Service) HandleStreamPlayback(userId uint, mediaType string, rawId string) error {
	slog.Info("Addon Stream Playback intercept", "userId", userId, "mediaType", mediaType, "id", rawId)
	tmdbID, imdbID, seasonNum, epNum := ParseStreamID(rawId)
	if tmdbID == 0 && imdbID == "" {
		return errors.New("cannot parse id from stream request")
	}

	targetContentType := entity.MOVIE
	if mediaType == "series" || mediaType == "tv" || mediaType == "show" || seasonNum > 0 {
		targetContentType = entity.SHOW
	}

	// 1. Resolve content in local database
	var content entity.Content
	query := s.db.Model(&entity.Content{})
	if tmdbID > 0 {
		query = query.Where("tmdb_id = ? AND type = ?", tmdbID, targetContentType)
	} else if imdbID != "" {
		query = query.Where("imdb_id = ? AND type = ?", imdbID, targetContentType)
	}

	if err := query.First(&content).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) && tmdbID > 0 {
			// Not cached locally yet, add it as WATCHING
			supportedMedia := util.SupportedMediaMovie
			if mediaType == "series" || mediaType == "tv" || seasonNum > 0 {
				supportedMedia = util.SupportedMediaShow
			}
			_, _ = s.watchedService.AddWatched(userId, domain.WatchedAddRequest{
				Status:      entity.WATCHING,
				ContentType: supportedMedia,
				TMDBID:      tmdbID,
			}, domain.WatchedAddExtraProps{})
			return nil
		}
		return err
	}

	now := time.Now().UTC()

	// 2. Find existing watched record for user
	var watched entity.Watched
	err := s.db.Model(&entity.Watched{}).
		Where("user_id = ? AND content_id = ?", userId, content.ID).
		First(&watched).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		supportedMedia := util.SupportedMediaMovie
		if content.Type == entity.SHOW {
			supportedMedia = util.SupportedMediaShow
		}
		createdWatched, addErr := s.watchedService.AddWatched(userId, domain.WatchedAddRequest{
			Status:      entity.WATCHING,
			ContentType: supportedMedia,
			TMDBID:      content.TmdbID,
		}, domain.WatchedAddExtraProps{})
		if addErr != nil {
			return addErr
		}
		watched = createdWatched
	} else if err == nil {
		// Update watched status to WATCHING if not already WATCHING or if it was PLANNED
		updates := map[string]any{
			"updated_at": now,
		}
		if watched.Status == entity.PLANNED {
			updates["status"] = entity.WATCHING
		}
		if seasonNum > 0 {
			updates["last_viewed_season"] = seasonNum
		}
		_ = s.db.Model(&entity.Watched{}).Where("id = ?", watched.ID).Updates(updates).Error
	}

	// 3. If TV show episode is played, update episode to WATCHING
	if content.Type == entity.SHOW && seasonNum > 0 && epNum > 0 {
		var ep entity.WatchedEpisode
		err := s.db.Where("user_id = ? AND watched_id = ? AND season_number = ? AND episode_number = ?",
			userId, watched.ID, seasonNum, epNum).First(&ep).Error
		if err != nil {
			// Create episode as WATCHING
			newEp := entity.WatchedEpisode{
				UserID:        userId,
				WatchedID:     watched.ID,
				SeasonNumber:  seasonNum,
				EpisodeNumber: epNum,
				Status:        entity.WATCHING,
			}
			_ = s.db.Create(&newEp).Error
		} else if ep.Status != entity.FINISHED {
			_ = s.db.Model(&entity.WatchedEpisode{}).Where("id = ?", ep.ID).Updates(map[string]any{
				"status":     entity.WATCHING,
				"updated_at": now,
			}).Error
		}
	}

	return nil
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

	now := time.Now().UTC()

	if p.Event == "start" {
		rawId := p.ImdbID
		if p.TmdbID > 0 {
			rawId = fmt.Sprintf("tmdb:%d", p.TmdbID)
		}
		if rawId != "" && p.SeasonNumber > 0 && p.EpisodeNumber > 0 {
			rawId = fmt.Sprintf("%s:%d:%d", rawId, p.SeasonNumber, p.EpisodeNumber)
		}
		if rawId != "" {
			return s.HandleStreamPlayback(userId, p.MediaType, rawId)
		}
		return nil
	}

	// We trigger watched status when progress is >= 80%
	if p.Progress < 80.0 {
		slog.Debug("Addon Scrobble: Progress under 80%, not marking watched yet", "progress", p.Progress)
		return nil
	}

	targetContentType := entity.MOVIE
	if p.MediaType == "series" || p.MediaType == "tv" || p.MediaType == "show" || p.SeasonNumber > 0 {
		targetContentType = entity.SHOW
	}

	// 1. Resolve content
	var content entity.Content
	query := s.db.Model(&entity.Content{})
	if p.TmdbID > 0 {
		query = query.Where("tmdb_id = ? AND type = ?", p.TmdbID, targetContentType)
	} else if p.ImdbID != "" {
		query = query.Where("imdb_id = ? AND type = ?", p.ImdbID, targetContentType)
	} else {
		return errors.New("scrobble missing tmdbId and imdbId")
	}

	if err := query.First(&content).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) && p.TmdbID > 0 {
			// Content might not be cached locally yet, let's cache it via tmdb
			supportedMedia := util.SupportedMediaMovie
			initialStatus := entity.FINISHED
			if targetContentType == entity.SHOW {
				supportedMedia = util.SupportedMediaShow
				initialStatus = entity.WATCHING
			}
			// Attempt to add watched
			createdWatched, addErr := s.watchedService.AddWatched(userId, domain.WatchedAddRequest{
				Status:      initialStatus,
				ContentType: supportedMedia,
				TMDBID:      p.TmdbID,
			}, domain.WatchedAddExtraProps{})
			if addErr != nil {
				return addErr
			}
			if targetContentType == entity.SHOW && p.SeasonNumber > 0 && p.EpisodeNumber > 0 {
				_, epErr := s.episodeService.SetWatchedEpisode(userId, domain.WatchedEpisodeSetRequest{
					WatchedID:     createdWatched.ID,
					SeasonNumber:  p.SeasonNumber,
					EpisodeNumber: p.EpisodeNumber,
					Status:        entity.FINISHED,
				})
				return epErr
			}
			return nil
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
		_ = s.db.Model(&entity.Watched{}).Where("id = ?", watched.ID).Update("updated_at", now).Error
		return upErr
	}

	// 4. Mark Episode as Finished for TV Show
	if content.Type == entity.SHOW && p.SeasonNumber > 0 && p.EpisodeNumber > 0 {
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

		// Keep series in WATCHING status and synchronize real timestamp
		_ = s.db.Model(&entity.Watched{}).
			Where("id = ?", watched.ID).
			Updates(map[string]any{
				"status":             entity.WATCHING,
				"last_viewed_season": p.SeasonNumber,
				"updated_at":         now,
			}).Error
	}

	return nil
}
