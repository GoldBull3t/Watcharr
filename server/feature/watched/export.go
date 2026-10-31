package watched

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/router"
)

func resolveExportItemMeta(item entity.Watched) (mediaType, title, year, tmdbID, imdbID, igdbID, ratingStr string) {
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
	if item.Rating > 0 {
		ratingStr = fmt.Sprintf("%.1f", item.Rating)
	}
	return
}

// ExportWatchedList exports the user's complete watched library in JSON, CSV, TXT, Markdown, or HTML format (#810, #925).
func (r *Router) ExportWatchedList(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	format := strings.ToLower(strings.TrimSpace(c.DefaultQuery("format", "json")))

	watched, err := r.s.getWatched(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: "failed to retrieve watched list for export"})
		return
	}

	dateStr := time.Now().Format("2006-01-02")

	switch format {
	case "csv":
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
			mediaType, title, year, tmdbID, imdbID, igdbID, ratingStr := resolveExportItemMeta(item)
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

	case "txt":
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("WATCHARR LIBRARY EXPORT (%s)\n", dateStr))
		sb.WriteString(fmt.Sprintf("Total entries: %d\n", len(watched)))
		sb.WriteString(strings.Repeat("=", 60) + "\n\n")

		statuses := []entity.WatchedStatus{
			entity.FINISHED,
			entity.WATCHING,
			entity.PLANNED,
			entity.HOLD,
			entity.DROPPED,
		}
		for _, st := range statuses {
			var group []entity.Watched
			for _, item := range watched {
				if item.Status == st {
					group = append(group, item)
				}
			}
			if len(group) == 0 {
				continue
			}
			sb.WriteString(fmt.Sprintf("[%s] (%d)\n", st, len(group)))
			sb.WriteString(strings.Repeat("-", 40) + "\n")
			for _, item := range group {
				mediaType, title, year, _, _, _, ratingStr := resolveExportItemMeta(item)
				if title == "" {
					continue
				}
				line := fmt.Sprintf("- %s", title)
				if year != "" {
					line += fmt.Sprintf(" (%s)", year)
				}
				if mediaType != "" {
					line += fmt.Sprintf(" [%s]", strings.ToUpper(mediaType))
				}
				if ratingStr != "" {
					line += fmt.Sprintf(" - %s/10", ratingStr)
				}
				sb.WriteString(line + "\n")
				if strings.TrimSpace(item.Thoughts) != "" {
					sb.WriteString(fmt.Sprintf("    Note: %s\n", strings.ReplaceAll(strings.TrimSpace(item.Thoughts), "\n", " ")))
				}
			}
			sb.WriteString("\n")
		}

		filename := fmt.Sprintf("watcharr_export_%s.txt", dateStr)
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		c.Header("Content-Type", "text/plain; charset=utf-8")
		c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(sb.String()))
		return

	case "md", "markdown":
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# Watcharr Library Export (%s)\n\n", dateStr))
		sb.WriteString(fmt.Sprintf("Total items: **%d**\n\n", len(watched)))

		statuses := []entity.WatchedStatus{
			entity.FINISHED,
			entity.WATCHING,
			entity.PLANNED,
			entity.HOLD,
			entity.DROPPED,
		}
		for _, st := range statuses {
			var group []entity.Watched
			for _, item := range watched {
				if item.Status == st {
					group = append(group, item)
				}
			}
			if len(group) == 0 {
				continue
			}
			sb.WriteString(fmt.Sprintf("## %s (%d)\n\n", st, len(group)))
			sb.WriteString("| Title | Year | Type | Rating | Notes |\n")
			sb.WriteString("| :--- | :--- | :--- | :--- | :--- |\n")
			for _, item := range group {
				mediaType, title, year, _, _, _, ratingStr := resolveExportItemMeta(item)
				if title == "" {
					continue
				}
				safeTitle := strings.ReplaceAll(title, "|", "\\|")
				safeNotes := strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(item.Thoughts), "\n", " "), "|", "\\|")
				if ratingStr != "" {
					ratingStr += "/10"
				} else {
					ratingStr = "-"
				}
				if year == "" {
					year = "-"
				}
				sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n", safeTitle, year, strings.ToUpper(mediaType), ratingStr, safeNotes))
			}
			sb.WriteString("\n")
		}

		filename := fmt.Sprintf("watcharr_export_%s.md", dateStr)
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		c.Header("Content-Type", "text/markdown; charset=utf-8")
		c.Data(http.StatusOK, "text/markdown; charset=utf-8", []byte(sb.String()))
		return

	case "html":
		var sb strings.Builder
		sb.WriteString("<!DOCTYPE html><html lang=\"en\"><head><meta charset=\"utf-8\">")
		sb.WriteString(fmt.Sprintf("<title>Watcharr Export - %s</title>", html.EscapeString(dateStr)))
		sb.WriteString("<style>body{font-family:system-ui,-apple-system,sans-serif;background:#0d1117;color:#e6edf3;max-width:960px;margin:2rem auto;padding:0 1rem}h1,h2{border-bottom:1px solid #30363d;padding-bottom:.4rem}table{width:100%;border-collapse:collapse;margin-bottom:2rem;font-size:.9rem}th,td{border:1px solid #30363d;padding:.5rem .75rem;text-align:left}th{background:#161b22;color:#8b949e}</style></head><body>")
		sb.WriteString(fmt.Sprintf("<h1>Watcharr Library Export (%s)</h1><p>Total items: <strong>%d</strong></p>", html.EscapeString(dateStr), len(watched)))
		sb.WriteString("<table><thead><tr><th>Title</th><th>Year</th><th>Type</th><th>Status</th><th>Rating</th><th>Notes</th></tr></thead><tbody>")
		for _, item := range watched {
			mediaType, title, year, _, _, _, ratingStr := resolveExportItemMeta(item)
			if title == "" {
				continue
			}
			sb.WriteString(fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>",
				html.EscapeString(title),
				html.EscapeString(year),
				html.EscapeString(strings.ToUpper(mediaType)),
				html.EscapeString(string(item.Status)),
				html.EscapeString(ratingStr),
				html.EscapeString(item.Thoughts),
			))
		}
		sb.WriteString("</tbody></table></body></html>")

		filename := fmt.Sprintf("watcharr_export_%s.html", dateStr)
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(sb.String()))
		return
	}

	// Default: JSON export
	filename := fmt.Sprintf("watcharr_export_%s.json", dateStr)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.JSON(http.StatusOK, watched)
}
