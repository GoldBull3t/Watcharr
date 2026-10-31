package server

import (
	"errors"
	"log/slog"

	"github.com/sbondCo/Watcharr/database/entity"
	"gorm.io/gorm"
)

type ServerStats struct {
	Users            int64          `json:"users"`
	PrivateUsers     int64          `json:"privateUsers"`
	WatchedMovies    int64          `json:"watchedMovies"`
	WatchedShows     int64          `json:"watchedShows"`
	WatchedSeasons   int64          `json:"watchedSeasons"`
	WatchedGames     int64          `json:"watchedGames"`
	MostWatchedMovie entity.Content `json:"mostWatchedMovie"`
	MostWatchedShow  entity.Content `json:"mostWatchedShow"`
	MostWatchedGame  entity.Game    `json:"mostWatchedGame"`
	Activities       int64          `json:"activities"`
}

// Collect and return server stats
// I cant sql so this the best yall gettin
func getServerStats(db *gorm.DB) ServerStats {
	stats := ServerStats{}
	// User counts.
	resp := db.
		Model(&entity.User{}).
		Count(&stats.Users).
		Where("private = 1").
		Count(&stats.PrivateUsers)
	if resp.Error != nil {
		slog.Error("getServerStats - Users query failed", "error", resp.Error)
	}
	// Watched seasons count.
	resp = db.Model(&entity.WatchedSeason{}).Count(&stats.WatchedSeasons)
	if resp.Error != nil {
		slog.Error("getServerStats - WatchedSeasons query failed", "error", resp.Error)
	}
	// Activities count.
	resp = db.Model(&entity.Activity{}).Count(&stats.Activities)
	if resp.Error != nil {
		slog.Error("getServerStats - Activities query failed", "error", resp.Error)
	}
	// Watched shows count.
	resp = db.
		Joins("JOIN contents ON contents.id = watcheds.content_id AND contents.type = ?", "tv").
		Find(&entity.Watched{}).
		Count(&stats.WatchedShows)
	if resp.Error != nil {
		slog.Error("getServerStats - WatchedShows query failed", "error", resp.Error)
	}
	// Watched movies count.
	resp = db.
		Joins("JOIN contents ON contents.id = watcheds.content_id AND contents.type = ?", "movie").
		Find(&entity.Watched{}).
		Count(&stats.WatchedMovies)
	if resp.Error != nil {
		slog.Error("getServerStats - WatchedMovies query failed", "error", resp.Error)
	}
	// Most watched show.
	var w entity.Watched
	resp = db.
		Model(&entity.Watched{}).
		Select("content_id, COUNT(*) AS mag").
		Joins("JOIN contents ON contents.type = ? AND contents.id = watcheds.content_id", "tv").
		Group("content_id").
		Order("mag DESC").
		Preload("Content").
		First(&w)
	if resp.Error != nil {
		if !errors.Is(resp.Error, gorm.ErrRecordNotFound) {
			slog.Error("getServerStats - MostWatchedShow query failed", "error", resp.Error)
		}
	} else if w.Content != nil {
		stats.MostWatchedShow = *w.Content
	}
	// Most watched movie.
	var wm entity.Watched
	resp = db.
		Model(&entity.Watched{}).
		Select("content_id, COUNT(*) AS mag").
		Joins("JOIN contents ON contents.type = ? AND contents.id = watcheds.content_id", "movie").
		Group("content_id").
		Order("mag DESC").
		Preload("Content").
		First(&wm)
	if resp.Error != nil {
		if !errors.Is(resp.Error, gorm.ErrRecordNotFound) {
			slog.Error("getServerStats - MostWatchedMovie query failed", "error", resp.Error)
		}
	} else if wm.Content != nil {
		stats.MostWatchedMovie = *wm.Content
	}
	// Watched games count (#349).
	resp = db.
		Model(&entity.Watched{}).
		Where("game_id IS NOT NULL").
		Count(&stats.WatchedGames)
	if resp.Error != nil {
		slog.Error("getServerStats - WatchedGames query failed", "error", resp.Error)
	}
	// Most watched game (#349).
	var wg entity.Watched
	resp = db.
		Model(&entity.Watched{}).
		Select("game_id, COUNT(*) AS mag").
		Where("game_id IS NOT NULL").
		Group("game_id").
		Order("mag DESC").
		Preload("Game").
		First(&wg)
	if resp.Error != nil {
		if !errors.Is(resp.Error, gorm.ErrRecordNotFound) {
			slog.Error("getServerStats - MostWatchedGame query failed", "error", resp.Error)
		}
	} else if wg.Game != nil {
		stats.MostWatchedGame = *wg.Game
	}
	return stats
}

// repairContentTable removes orphaned/broken content records and triggers metadata backfill (#62).
func repairContentTable(db *gorm.DB) (int64, error) {
	// 1. Delete invalid/corrupted content rows (tmdb_id <= 0 or empty type)
	res := db.Unscoped().Where("tmdb_id <= 0 OR type = '' OR title = ''").Delete(&entity.Content{})
	if res.Error != nil {
		return 0, res.Error
	}
	repaired := res.RowsAffected

	// 2. Mark rows missing overview or poster_path with empty backdrop_path so BackfillMissingMetadata refreshes them
	resUpdate := db.Model(&entity.Content{}).
		Where("overview = '' OR poster_path = ''").
		UpdateColumn("backdrop_path", "")
	if resUpdate.Error == nil {
		repaired += resUpdate.RowsAffected
	}
	return repaired, nil
}
