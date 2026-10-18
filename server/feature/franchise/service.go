package franchise

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/watched"
	"github.com/sbondCo/Watcharr/feature/watched/addedtocontent"
	"github.com/sbondCo/Watcharr/media/tmdb"
	"github.com/sbondCo/Watcharr/util"
	"gorm.io/gorm"
)

type Service struct {
	db             *gorm.DB
	cfg            *config.ServerConfig
	tmdb           *tmdb.TMDB
	watchedService *watched.Service
}

func NewService(db *gorm.DB, cfg *config.ServerConfig, tmdb *tmdb.TMDB, watchedService *watched.Service) *Service {
	return &Service{
		db:             db,
		cfg:            cfg,
		tmdb:           tmdb,
		watchedService: watchedService,
	}
}

// GetFranchises returns a summary list of all curated franchises with user progress
func (s *Service) GetFranchises(userId uint) ([]FranchiseSummary, error) {
	pairsMap := make(map[string]addedtocontent.IdToTypePair)
	for _, f := range CuratedFranchises {
		for _, item := range f.Items {
			key := fmt.Sprintf("%s:%d", item.Type, item.TmdbID)
			mediaType := util.SupportedMediaMovie
			if item.Type == "tv" {
				mediaType = util.SupportedMediaShow
			}
			pairsMap[key] = addedtocontent.IdToTypePair{
				Id:   item.TmdbID,
				Type: mediaType,
			}
		}
	}

	pairs := make([]addedtocontent.IdToTypePair, 0, len(pairsMap))
	for _, p := range pairsMap {
		pairs = append(pairs, p)
	}

	watchedMap := make(map[string]bool)
	if len(pairs) > 0 && s.watchedService != nil {
		watchedItems, err := s.watchedService.GetWatchedItemsBySupportedMediaIds(userId, pairs)
		if err == nil {
			for _, w := range watchedItems {
				if w.Content != nil {
					key := fmt.Sprintf("%s:%d", w.Content.Type, w.Content.TmdbID)
					if w.Status == entity.FINISHED {
						watchedMap[key] = true
					}
				}
			}
		}
	}

	summaries := make([]FranchiseSummary, len(CuratedFranchises))
	for i, f := range CuratedFranchises {
		watchedCount := 0
		totalItems := len(f.Items)
		for _, item := range f.Items {
			key := fmt.Sprintf("%s:%d", item.Type, item.TmdbID)
			if watchedMap[key] {
				watchedCount++
			}
		}

		pct := 0
		if totalItems > 0 {
			pct = int(math.Round(float64(watchedCount) / float64(totalItems) * 100))
		}

		summaries[i] = FranchiseSummary{
			ID:             f.ID,
			Name:           f.Name,
			Description:    f.Description,
			BackdropPath:   f.BackdropPath,
			PosterPath:     f.PosterPath,
			TotalItems:     totalItems,
			WatchedCount:   watchedCount,
			PercentWatched: pct,
			ItemTypes:      f.ItemTypes,
		}
	}

	return summaries, nil
}

// GetFranchiseRoadmap builds a full ordered timeline for a franchise
func (s *Service) GetFranchiseRoadmap(userId uint, franchiseId string, order string, userLang string) (*FranchiseRoadmap, error) {
	if strings.TrimSpace(order) == "" {
		order = "timeline"
	}
	order = strings.ToLower(order)
	if order != "release" && order != "timeline" {
		order = "timeline"
	}

	// TMDB Collection Fallback
	if strings.HasPrefix(franchiseId, "tmdb-") {
		colIdStr := strings.TrimPrefix(franchiseId, "tmdb-")
		colId, err := strconv.Atoi(colIdStr)
		if err != nil || colId <= 0 {
			return nil, errors.New("id de coleção tmdb inválido")
		}

		tmdbCol, err := s.tmdb.Collection(colIdStr, map[string]string{"language": userLang})
		if err != nil {
			slog.Error("GetFranchiseRoadmap: failed fetching TMDB collection", "colId", colIdStr, "error", err)
			return nil, errors.New("coleção não encontrada no TMDB")
		}

		// Sort parts by release date
		sort.Slice(tmdbCol.Parts, func(i, j int) bool {
			return tmdbCol.Parts[i].ReleaseDate < tmdbCol.Parts[j].ReleaseDate
		})

		items := make([]FranchiseItem, len(tmdbCol.Parts))
		pairs := make([]addedtocontent.IdToTypePair, len(tmdbCol.Parts))
		for i, p := range tmdbCol.Parts {
			year := ""
			if len(p.ReleaseDate) >= 4 {
				year = p.ReleaseDate[:4]
			}
			items[i] = FranchiseItem{
				TmdbID:        p.ID,
				Type:          "movie",
				Title:         p.Title,
				Year:          year,
				PosterPath:    p.PosterPath,
				BackdropPath:  p.BackdropPath,
				Overview:      p.Overview,
				PhaseOrEra:    fmt.Sprintf("Parte %d", i+1),
				TimelineOrder: i + 1,
				ReleaseOrder:  i + 1,
				ReleaseDate:   p.ReleaseDate,
			}
			pairs[i] = addedtocontent.IdToTypePair{
				Id:   p.ID,
				Type: util.SupportedMediaMovie,
			}
		}

		// Match watched
		watchedMap := make(map[int]entity.Watched)
		if len(pairs) > 0 && s.watchedService != nil {
			wItems, _ := s.watchedService.GetWatchedItemsBySupportedMediaIds(userId, pairs)
			for _, w := range wItems {
				if w.Content != nil {
					watchedMap[w.Content.TmdbID] = w
				}
			}
		}

		watchedCount := 0
		var nextUp *FranchiseItem
		for i := range items {
			if w, ok := watchedMap[items[i].TmdbID]; ok {
				dto := domain.NewWatchedDtoForLists(&w)
				items[i].Watched = &dto
				items[i].IsWatched = (w.Status == entity.FINISHED)
			}
			if items[i].IsWatched {
				watchedCount++
			} else if nextUp == nil {
				items[i].IsNextUp = true
				nextUp = &items[i]
			}
		}

		pct := 0
		if len(items) > 0 {
			pct = int(math.Round(float64(watchedCount) / float64(len(items)) * 100))
		}

		return &FranchiseRoadmap{
			ID:             franchiseId,
			Name:           tmdbCol.Name,
			Description:    tmdbCol.Overview,
			BackdropPath:   tmdbCol.BackdropPath,
			PosterPath:     tmdbCol.PosterPath,
			ActiveOrder:    order,
			TotalItems:     len(items),
			WatchedCount:   watchedCount,
			PercentWatched: pct,
			NextUpItem:     nextUp,
			Items:          items,
		}, nil
	}

	// Curated Franchise
	var curated *CuratedFranchiseDefinition
	for i := range CuratedFranchises {
		if CuratedFranchises[i].ID == franchiseId {
			curated = &CuratedFranchises[i]
			break
		}
	}
	if curated == nil {
		return nil, errors.New("franquia não encontrada")
	}

	// Deep copy items
	items := make([]FranchiseItem, len(curated.Items))
	copy(items, curated.Items)

	// Fetch watched items for user
	pairs := make([]addedtocontent.IdToTypePair, len(items))
	contentKeys := make([]string, len(items))
	for i, item := range items {
		mType := util.SupportedMediaMovie
		if item.Type == "tv" {
			mType = util.SupportedMediaShow
		}
		pairs[i] = addedtocontent.IdToTypePair{
			Id:   item.TmdbID,
			Type: mType,
		}
		contentKeys[i] = fmt.Sprintf("%s:%d", item.Type, item.TmdbID)
	}

	watchedMap := make(map[string]entity.Watched)
	if len(pairs) > 0 && s.watchedService != nil {
		wList, err := s.watchedService.GetWatchedItemsBySupportedMediaIds(userId, pairs)
		if err == nil {
			for _, w := range wList {
				if w.Content != nil {
					key := fmt.Sprintf("%s:%d", w.Content.Type, w.Content.TmdbID)
					watchedMap[key] = w
				}
			}
		}
	}

	// Fetch cached content from database to fill posters/backdrops if missing (scoped to franchise items)
	tmdbIds := make([]int, len(items))
	for i, item := range items {
		tmdbIds[i] = item.TmdbID
	}
	var cachedContents []entity.Content
	if len(tmdbIds) > 0 {
		s.db.Model(&entity.Content{}).Where("tmdb_id IN ?", tmdbIds).Find(&cachedContents)
	}
	contentMap := make(map[string]entity.Content)
	for _, c := range cachedContents {
		key := fmt.Sprintf("%s:%d", c.Type, c.TmdbID)
		contentMap[key] = c
	}

	for i := range items {
		key := fmt.Sprintf("%s:%d", items[i].Type, items[i].TmdbID)
		if w, ok := watchedMap[key]; ok {
			dto := domain.NewWatchedDtoForLists(&w)
			items[i].Watched = &dto
			items[i].IsWatched = (w.Status == entity.FINISHED)
		}
		if c, ok := contentMap[key]; ok {
			if items[i].PosterPath == "" && c.PosterPath != "" {
				items[i].PosterPath = c.PosterPath
			}
			if items[i].Overview == "" && c.Overview != "" {
				items[i].Overview = c.Overview
			}
		}
	}

	// Sort items according to active order
	if order == "release" {
		sort.Slice(items, func(i, j int) bool {
			return items[i].ReleaseOrder < items[j].ReleaseOrder
		})
	} else {
		sort.Slice(items, func(i, j int) bool {
			return items[i].TimelineOrder < items[j].TimelineOrder
		})
	}

	// Calculate watched count and first nextUp item
	watchedCount := 0
	var nextUp *FranchiseItem
	for i := range items {
		if items[i].IsWatched {
			watchedCount++
		} else if nextUp == nil {
			items[i].IsNextUp = true
			nextUp = &items[i]
		}
	}

	pct := 0
	if len(items) > 0 {
		pct = int(math.Round(float64(watchedCount) / float64(len(items)) * 100))
	}

	return &FranchiseRoadmap{
		ID:             curated.ID,
		Name:           curated.Name,
		Description:    curated.Description,
		BackdropPath:   curated.BackdropPath,
		PosterPath:     curated.PosterPath,
		ActiveOrder:    order,
		TotalItems:     len(items),
		WatchedCount:   watchedCount,
		PercentWatched: pct,
		NextUpItem:     nextUp,
		Items:          items,
	}, nil
}

// FindFranchiseForContent checks if a movie or tv show belongs to a known universe
func (s *Service) FindFranchiseForContent(contentType string, tmdbId int) (*FranchiseMatch, error) {
	for _, f := range CuratedFranchises {
		for _, item := range f.Items {
			if item.Type == contentType && item.TmdbID == tmdbId {
				return &FranchiseMatch{
					InFranchise:   true,
					FranchiseID:   f.ID,
					FranchiseName: f.Name,
					PhaseOrEra:    item.PhaseOrEra,
					TimelineOrder: item.TimelineOrder,
					ReleaseOrder:  item.ReleaseOrder,
				}, nil
			}
		}
	}

	return &FranchiseMatch{
		InFranchise: false,
	}, nil
}
