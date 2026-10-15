package plex

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/dbmodel"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/internal/testutil"
)

type mockWatchedProvider struct {
	added []domain.WatchedAddRequest
}

func (m *mockWatchedProvider) AddWatched(
	userId uint,
	ar domain.WatchedAddRequest,
	extraProps domain.WatchedAddExtraProps,
) (entity.Watched, error) {
	m.added = append(m.added, ar)
	return entity.Watched{
		GormModel: dbmodel.GormModel{ID: uint(len(m.added))},
		Status:    ar.Status,
	}, nil
}

type mockSeasonProvider struct{}

func (m *mockSeasonProvider) SetWatchedSeason(
	userId uint,
	ar domain.WatchedSeasonSetRequest,
) (domain.WatchedSeasonSetResponse, error) {
	return domain.WatchedSeasonSetResponse{}, nil
}

type mockEpisodeProvider struct{}

func (m *mockEpisodeProvider) SetWatchedEpisode(
	userId uint,
	ar domain.WatchedEpisodeSetRequest,
) (domain.WatchedEpisodeSetResponse, error) {
	return domain.WatchedEpisodeSetResponse{}, nil
}

func TestPlexSync_ShowStatusHandling(t *testing.T) {
	// Setup mock Plex API server
	mockPlexServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/library/sections":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"MediaContainer": map[string]interface{}{
					"Directory": []map[string]interface{}{
						{
							"key":   "1",
							"type":  "show",
							"title": "TV Shows",
						},
					},
				},
			})
		case "/library/sections/1/all":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"MediaContainer": map[string]interface{}{
					"Metadata": []map[string]interface{}{
						// 1. Unwatched show: 10 episodes, 0 watched
						{
							"ratingKey":       "101",
							"title":           "Unwatched Show",
							"leafCount":       10,
							"viewedLeafCount": 0,
							"guid":            "plex://show/101",
							"Guid": []map[string]interface{}{
								{"id": "tmdb://1001"},
							},
						},
						// 2. Partially watched show: 10 episodes, 4 watched
						{
							"ratingKey":       "102",
							"title":           "Partially Watched Show",
							"leafCount":       10,
							"viewedLeafCount": 4,
							"lastViewedAt":    time.Now().Unix(),
							"guid":            "plex://show/102",
							"Guid": []map[string]interface{}{
								{"id": "tmdb://1002"},
							},
						},
						// 3. Fully watched show: 10 episodes, 10 watched
						{
							"ratingKey":       "103",
							"title":           "Fully Watched Show",
							"leafCount":       10,
							"viewedLeafCount": 10,
							"lastViewedAt":    time.Now().Unix(),
							"guid":            "plex://show/103",
							"Guid": []map[string]interface{}{
								{"id": "tmdb://1003"},
							},
						},
					},
				},
			})
		default:
			// For seasons / episodes endpoints, return empty metadata container
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"MediaContainer": map[string]interface{}{
					"Metadata": []interface{}{},
				},
			})
		}
	}))
	defer mockPlexServer.Close()

	db := testutil.SetupDB(t)
	cfg := &config.ServerConfig{PLEX_HOST: mockPlexServer.URL}
	plexSvc := NewService(cfg)
	wp := &mockWatchedProvider{}
	wsp := &mockSeasonProvider{}
	wep := &mockEpisodeProvider{}
	syncSvc := NewSyncService(db, plexSvc, wp, wsp, wep)

	// Run sync synchronously
	syncSvc.startPlexSync("job-test-1", 1, "test-token")

	// We expect 2 shows added:
	// - Partially Watched Show (TMDB: 1002) with status WATCHING
	// - Fully Watched Show (TMDB: 1003) with status FINISHED
	// Unwatched Show (TMDB: 1001) should NOT be added.
	if len(wp.added) != 2 {
		t.Fatalf("expected 2 shows to be added, got %d", len(wp.added))
	}

	partial := wp.added[0]
	if partial.TMDBID != 1002 {
		t.Errorf("expected first show TMDB 1002, got %d", partial.TMDBID)
	}
	if partial.Status != entity.WATCHING {
		t.Errorf("expected partially watched show to have status WATCHING, got %s", partial.Status)
	}

	fully := wp.added[1]
	if fully.TMDBID != 1003 {
		t.Errorf("expected second show TMDB 1003, got %d", fully.TMDBID)
	}
	if fully.Status != entity.FINISHED {
		t.Errorf("expected fully watched show to have status FINISHED, got %s", fully.Status)
	}
}
