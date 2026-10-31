package game

import (
	"errors"
	"log/slog"
	"strconv"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/image"
	"github.com/sbondCo/Watcharr/media/igdb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Service struct {
	db   *gorm.DB
	igdb *igdb.IGDB
}

func NewService(db *gorm.DB, igdb *igdb.IGDB) *Service {
	return &Service{
		db,
		igdb,
	}
}

// Cache(save) game to our table
func (s *Service) saveGame(c *entity.Game, onlyUpdate bool) error {
	slog.Info("Saving game to db", "id", c.IgdbID, "name", c.Name)
	if c.IgdbID == 0 || c.Name == "" {
		slog.Error("saveGame: content missing id or name!", "id", c.IgdbID, "name", c.Name)
		return errors.New("game missing id or title")
	}
	if onlyUpdate {
		var existingCount int64
		if err := s.db.Model(&entity.Game{}).Where("igdb_id = ?", c.IgdbID).Count(&existingCount).Error; err != nil || existingCount == 0 {
			return nil
		}
	}
	if c.CoverID != "" {
		p, err := image.
			NewSaver(
				s.db,
				"games",
				image.ValidateOptions{
					// To avoid losing quality, we want to keep png format
					// for our game posters.
					ToFormat: image.ValidateAllowedFormatPNG,
				},
			).
			DownloadAndInsertFromUrl(
				"https://images.igdb.com/igdb/image/upload/t_cover_big/" +
					c.CoverID + ".png")
		if err != nil {
			slog.Error("saveGame: Failed to cache game cover.", "error", err)
		} else {
			slog.Debug("saveGame: Cached game cover", "p", p)
			c.PosterID = &p.ID
		}
	}
	var res *gorm.DB
	if onlyUpdate {
		// We only want to update an existing row, if it exists.
		res = s.db.Model(&entity.Game{}).Where("igdb_id = ?", c.IgdbID).Updates(c)
		if res.Error != nil {
			slog.Error("saveGame: Error updating game in database", "error", res.Error.Error())
			return errors.New("failed to update cached game in database")
		}
	} else {
		// On conflict, update existing row with details incase any were updated/missing.
		res = s.db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "igdb_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"name",
				"cover_id",
				"backdrop_id",
				"summary",
				"storyline",
				"release_date",
				"rating",
				"rating_count",
				"aggregated_rating",
				"time_to_beat_normal",
				"status",
				"game_modes",
				"genres",
				"platforms",
				"developers",
				"publishers",
			}),
		}).Create(&c)
		if res.Error != nil {
			// Error if anything but unique contraint error
			if res.Error != gorm.ErrDuplicatedKey {
				slog.Error("saveGame: Error creating game in database", "error", res.Error.Error())
				return errors.New("failed to cache game in database")
			}
		}
	}
	return nil
}

func (s *Service) cacheGame(g igdb.GameDetailsBasicResponse, onlyUpdate bool) (entity.Game, error) {
	slog.Debug("cacheGame", "game_details", g)
	var (
		gameModes  string
		genres     string
		platforms  string
		developers string
		publishers string
		backdropID string
	)
	if len(g.Artworks) > 0 {
		backdropID = g.Artworks[0].ImageID
	}
	if len(g.GameModes) > 0 {
		for _, v := range g.GameModes {
			gameModes += v.Name + "|"
		}
	}
	if len(g.Genres) > 0 {
		for _, v := range g.Genres {
			genres += v.Name + "|"
		}
	}
	if len(g.Platforms) > 0 {
		for _, v := range g.Platforms {
			platforms += v.Name + "|"
		}
	}
	for _, ic := range g.InvolvedCompanies {
		if ic.Company.Name == "" {
			continue
		}
		if ic.Developer {
			developers += ic.Company.Name + "|"
		}
		if ic.Publisher {
			publishers += ic.Company.Name + "|"
		}
	}
	c := entity.Game{
		IgdbID:           g.ID,
		Name:             g.Name,
		CoverID:          g.Cover.ImageID,
		BackdropID:       backdropID,
		Summary:          g.Summary,
		Storyline:        g.Storyline,
		ReleaseDate:      &g.FirstReleaseDate.Time,
		Rating:           (g.Rating),
		RatingCount:      g.RatingCount,
		AggregatedRating: g.AggregatedRating,
		TimeToBeatNormal: g.TimeToBeatNormal,
		Status:           g.Status,
		Category:         g.Category,
		GameModes:        gameModes,
		Genres:           genres,
		Platforms:        platforms,
		Developers:       developers,
		Publishers:       publishers,
	}
	err := s.saveGame(&c, onlyUpdate)
	if err != nil {
		slog.Error("cacheGame: Failed to save game!", "error", err)
		return entity.Game{}, errors.New("failed to save game")
	}
	return c, nil
}

// CacheGameFromDetails updates or inserts a cached game record from full GameDetailsResponse (#1144).
func (s *Service) CacheGameFromDetails(g igdb.GameDetailsResponse, onlyUpdate bool) (entity.Game, error) {
	var (
		gameModes        string
		genres           string
		platforms        string
		developers       string
		publishers       string
		backdropID       string
		timeToBeatNormal int
	)
	if len(g.Artworks) > 0 {
		backdropID = g.Artworks[0].ImageID
	} else if len(g.Screenshots) > 0 {
		backdropID = g.Screenshots[0].ImageID
	}
	for _, v := range g.GameModes {
		gameModes += v.Name + "|"
	}
	for _, v := range g.Genres {
		genres += v.Name + "|"
	}
	for _, v := range g.Platforms {
		platforms += v.Name + "|"
	}
	for _, ic := range g.InvolvedCompanies {
		if ic.Company.Name == "" {
			continue
		}
		if ic.Developer {
			developers += ic.Company.Name + "|"
		}
		if ic.Publisher {
			publishers += ic.Company.Name + "|"
		}
	}
	if g.TimeToBeat != nil && g.TimeToBeat.Normally > 0 {
		timeToBeatNormal = g.TimeToBeat.Normally
	}
	c := entity.Game{
		IgdbID:           g.ID,
		Name:             g.Name,
		CoverID:          g.Cover.ImageID,
		BackdropID:       backdropID,
		Summary:          g.Summary,
		Storyline:        g.Storyline,
		ReleaseDate:      &g.FirstReleaseDate.Time,
		Rating:           g.Rating,
		RatingCount:      g.RatingCount,
		AggregatedRating: g.AggregatedRating,
		TimeToBeatNormal: timeToBeatNormal,
		Status:           g.Status,
		Category:         g.GameType,
		GameModes:        gameModes,
		Genres:           genres,
		Platforms:        platforms,
		Developers:       developers,
		Publishers:       publishers,
	}
	if err := s.saveGame(&c, onlyUpdate); err != nil {
		slog.Error("CacheGameFromDetails: Failed to save game!", "error", err)
		return entity.Game{}, err
	}
	return c, nil
}

func (s *Service) GetOrCache(igdbID int) (entity.Game, error) {
	var game entity.Game
	s.db.Where("igdb_id = ?", igdbID).Find(&game)

	// Create game if not found from our db
	if game == (entity.Game{}) {
		slog.Debug("GetOrCache: Game not in db, fetching...")

		resp, err := s.igdb.GameDetailsBasic(strconv.Itoa(igdbID))
		if err != nil {
			slog.Error("GetOrCache: content api request failed", "error", err)
			return game, errors.New("failed to find requested games")
		}

		game, err = s.cacheGame(resp, false)
		if err != nil {
			slog.Error("GetOrCache: failed to cache game",
				"igdb_id", igdbID,
				"err", err)
			return game, errors.New("failed to cache content")
		}
	}

	return game, nil
}
