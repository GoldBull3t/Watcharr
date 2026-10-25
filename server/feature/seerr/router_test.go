package seerr

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/router"
	"github.com/sbondCo/Watcharr/seerr"
)

func TestSeerrRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Mock Overseerr upstream server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-Api-Key")
		if apiKey != "test_key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		switch r.URL.Path {
		case "/api/v1/status":
			json.NewEncoder(w).Encode(seerr.StatusResponse{
				Version:   "1.7.0",
				CommitTag: "v1.7.0",
			})
		case "/api/v1/movie/100":
			json.NewEncoder(w).Encode(seerr.MediaResponse{
				ID: 100,
				MediaInfo: &seerr.MediaInfo{
					ID:     1,
					TmdbID: 100,
					Status: seerr.MediaStatusAvailable,
				},
			})
		case "/api/v1/request":
			var req seerr.RequestPayload
			_ = json.NewDecoder(r.Body).Decode(&req)
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(seerr.RequestResponse{
				ID:     77,
				Status: 1,
				Media: &seerr.MediaInfo{
					TmdbID: req.MediaID,
					Status: seerr.MediaStatusPendingApproval,
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	cfg := &config.ServerConfig{
		OVERSEERR_HOST: ts.URL,
		OVERSEERR_KEY:  "test_key",
	}

	engine := gin.New()
	br := &router.BaseRouter{
		Router: &engine.RouterGroup,
		Cfg:    cfg,
	}

	r := NewRouter(br)

	// Direct handler tests
	t.Run("TestConnection Handler", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(seerr.TestRequest{Host: ts.URL, Key: "test_key"})
		c.Request = httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		r.TestConnection(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp seerr.StatusResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal status: %v", err)
		}
		if resp.Version != "1.7.0" {
			t.Errorf("expected version 1.7.0, got %s", resp.Version)
		}
	})

	t.Run("GetStatus Handler", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{
			{Key: "type", Value: "movie"},
			{Key: "tmdbId", Value: "100"},
		}
		c.Request = httptest.NewRequest(http.MethodGet, "/status/movie/100", nil)

		r.GetStatus(c)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp seerr.ContentStatusResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal content status: %v", err)
		}
		if !resp.IsAvailable {
			t.Errorf("expected isAvailable true, got false")
		}
		if resp.StatusName != "AVAILABLE" {
			t.Errorf("expected AVAILABLE status, got %s", resp.StatusName)
		}
	})

	t.Run("CreateRequest Handler", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(seerr.RequestPayload{
			MediaType: "movie",
			MediaID:   100,
		})
		c.Request = httptest.NewRequest(http.MethodPost, "/request", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		r.CreateRequest(c)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		var resp seerr.RequestResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to unmarshal request response: %v", err)
		}
		if resp.ID != 77 {
			t.Errorf("expected request id 77, got %d", resp.ID)
		}
	})
}
