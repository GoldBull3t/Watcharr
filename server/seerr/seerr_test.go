package seerr

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSeerrClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-Api-Key")
		if apiKey != "valid_key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/api/v1/status":
			json.NewEncoder(w).Encode(StatusResponse{
				Version:   "1.7.0",
				CommitTag: "v1.7.0",
			})
		case "/api/v1/movie/550":
			json.NewEncoder(w).Encode(MediaResponse{
				ID: 550,
				MediaInfo: &MediaInfo{
					ID:     10,
					TmdbID: 550,
					Status: MediaStatusAvailable,
				},
			})
		case "/api/v1/movie/999":
			w.WriteHeader(http.StatusNotFound)
		case "/api/v1/request":
			var req RequestPayload
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(RequestResponse{
				ID:     42,
				Status: 1,
				Media: &MediaInfo{
					TmdbID: req.MediaID,
					Status: MediaStatusPendingApproval,
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	t.Run("Test Connection Success", func(t *testing.T) {
		client := NewClient(ts.URL, "valid_key")
		status, err := client.Test()
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if status.Version != "1.7.0" {
			t.Errorf("expected version 1.7.0, got %s", status.Version)
		}
	})

	t.Run("Test Connection Invalid Key", func(t *testing.T) {
		client := NewClient(ts.URL, "invalid_key")
		_, err := client.Test()
		if err == nil {
			t.Fatal("expected unauthorized error, got nil")
		}
	})

	t.Run("Get Media Available", func(t *testing.T) {
		client := NewClient(ts.URL, "valid_key")
		media, err := client.GetMediaStatus("movie", 550)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if media.Status != MediaStatusAvailable {
			t.Errorf("expected status %d, got %d", MediaStatusAvailable, media.Status)
		}
		if media.Status.String() != "AVAILABLE" {
			t.Errorf("expected status string AVAILABLE, got %s", media.Status.String())
		}
	})

	t.Run("Get Media Not Found", func(t *testing.T) {
		client := NewClient(ts.URL, "valid_key")
		media, err := client.GetMediaStatus("movie", 999)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if media.Status != MediaStatusUnknown {
			t.Errorf("expected unknown status for unindexed media, got %d", media.Status)
		}
	})

	t.Run("Create Request Success", func(t *testing.T) {
		client := NewClient(ts.URL, "valid_key")
		resp, err := client.CreateRequest(RequestPayload{
			MediaType: "movie",
			MediaID:   12345,
		})
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if resp.ID != 42 {
			t.Errorf("expected request id 42, got %d", resp.ID)
		}
	})
}

