package watched

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/router"
)

// ExportWatchedList exports the user's complete watched library in JSON or CSV format.
func (r *Router) ExportWatchedList(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	format := c.DefaultQuery("format", "json")

	watched, err := r.s.getWatched(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: "failed to retrieve watched list for export"})
		return
	}

	dateStr := time.Now().Format("2006-01-02")

	if format == "csv" {
		buf := new(bytes.Buffer)
		w := csv.NewWriter(buf)

		header := []string{
			"media_type",
			"title",
			"year",
			"tmdb_id",
			"imdb_id",
			"igdb_id",
			"status",
			"rating",
			"thoughts",
			"date_added",
			"date_modified",
			"watched_episodes",
		}
		if err := w.Write(header); err != nil {
			c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: "failed to generate csv header"})
			return
		}

		for _, item := range watched {
			mediaType := ""
			title := ""
			year := ""
			tmdbID := ""
			imdbID := ""
			igdbID := ""

			if item.Content != nil {
				mediaType = string(item.Content.Type)
				title = item.Content.Title
				tmdbID = strconv.Itoa(item.Content.TmdbID)
				imdbID = item.Content.ImdbID
				if item.Content.ReleaseDate != nil {
					year = strconv.Itoa(item.Content.ReleaseDate.Year())
				}
			} else if item.Game != nil {
				mediaType = "game"
				title = item.Game.Name
				igdbID = strconv.Itoa(item.Game.IgdbID)
				if item.Game.ReleaseDate != nil {
					year = strconv.Itoa(item.Game.ReleaseDate.Year())
				}
			}

			ratingStr := ""
			if item.Rating > 0 {
				ratingStr = fmt.Sprintf("%.1f", item.Rating)
			}

			epCount := strconv.Itoa(len(item.WatchedEpisodes))

			row := []string{
				mediaType,
				title,
				year,
				tmdbID,
				imdbID,
				igdbID,
				string(item.Status),
				ratingStr,
				item.Thoughts,
				item.CreatedAt.Format(time.RFC3339),
				item.UpdatedAt.Format(time.RFC3339),
				epCount,
			}
			if err := w.Write(row); err != nil {
				slog.Error("ExportWatched: Failed to write CSV row", "error", err)
				c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: "failed to write csv export"})
				return
			}
		}

		w.Flush()
		if err := w.Error(); err != nil {
			c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: "failed to finalize csv export"})
			return
		}

		filename := fmt.Sprintf("watcharr_export_%s.csv", dateStr)
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Data(http.StatusOK, "text/csv; charset=utf-8", buf.Bytes())
		return
	}

	// Default: JSON export
	filename := fmt.Sprintf("watcharr_export_%s.json", dateStr)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.JSON(http.StatusOK, watched)
}
