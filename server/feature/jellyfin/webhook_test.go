package jellyfin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/router"
)

func TestEmbyWebhookDataConversion(t *testing.T) {
	t.Run("Emby Movie PlaybackStop Played", func(t *testing.T) {
		var emby EmbyWebhookData
		emby.Event = "playback.stop"
		emby.User.ID = "emby_user_123"
		emby.User.Name = "Bernardo"
		emby.Item.ID = "item_1"
		emby.Item.Name = "Inception"
		emby.Item.Type = "Movie"
		emby.Item.Played = true
		emby.Item.ProviderIds = map[string]string{
			"Tmdb": "27205",
			"Imdb": "tt1375666",
		}

		data, handled, ignored := emby.ToWebhookData()
		if ignored {
			t.Fatal("expected handled, got ignored")
		}
		if !handled || data == nil {
			t.Fatal("expected handled true and data not nil")
		}
		if data.NotificationType != WebhookTypePlaybackStop {
			t.Errorf("expected WebhookTypePlaybackStop, got %s", data.NotificationType)
		}
		if !data.PlayedToCompletion {
			t.Errorf("expected PlayedToCompletion true")
		}
		if data.ProviderTMDBID != "27205" {
			t.Errorf("expected tmdb 27205, got %s", data.ProviderTMDBID)
		}
		if data.UserID != "emby_user_123" {
			t.Errorf("expected user id emby_user_123, got %s", data.UserID)
		}
	})

	t.Run("Emby Episode PlaybackStop 95% Progress", func(t *testing.T) {
		var emby EmbyWebhookData
		emby.Event = "playback.stop"
		emby.User.ID = "emby_user_123"
		emby.Item.ID = "item_ep_5"
		emby.Item.Name = "Pilot"
		emby.Item.Type = "Episode"
		emby.Item.SeriesName = "Severance"
		emby.Item.SeriesID = "series_99"
		emby.Item.IndexNumber = 1
		emby.Item.ParentIndexNumber = 1
		emby.Item.Played = false
		emby.PlaybackInfo.PositionTicks = 9500000000
		emby.PlaybackInfo.TotalTicks = 10000000000

		data, handled, ignored := emby.ToWebhookData()
		if ignored || !handled || data == nil {
			t.Fatal("expected handled true and data not nil")
		}
		if !data.PlayedToCompletion {
			t.Errorf("expected PlayedToCompletion true for 95%% progress")
		}
		if data.EpisodeNumber != 1 {
			t.Errorf("expected episode number 1, got %d", data.EpisodeNumber)
		}
		if data.SeasonNumber == nil || *data.SeasonNumber != 1 {
			t.Errorf("expected season number 1")
		}
		if data.SeriesID != "series_99" {
			t.Errorf("expected series id series_99, got %s", data.SeriesID)
		}
	})

	t.Run("Emby ItemMarkPlayed Event", func(t *testing.T) {
		var emby EmbyWebhookData
		emby.Event = "item.markplayed"
		emby.User.ID = "emby_user_123"
		emby.Item.ID = "item_mov_2"
		emby.Item.Name = "The Batman"
		emby.Item.Type = "Movie"
		emby.Item.ProviderIds = map[string]string{"TMDB": "414906"}

		data, handled, ignored := emby.ToWebhookData()
		if ignored || !handled || data == nil {
			t.Fatal("expected handled true and data not nil")
		}
		if data.NotificationType != WebhookTypeUserDataSaved {
			t.Errorf("expected WebhookTypeUserDataSaved, got %s", data.NotificationType)
		}
		if data.SaveReason != WebhookUserDataSaveReasonTogglePlayed {
			t.Errorf("expected TogglePlayed, got %s", data.SaveReason)
		}
		if !data.Played {
			t.Errorf("expected Played true")
		}
		if data.ProviderTMDBID != "414906" {
			t.Errorf("expected TMDB 414906, got %s", data.ProviderTMDBID)
		}
	})

	t.Run("Emby Ignored Event (playback.pause)", func(t *testing.T) {
		var emby EmbyWebhookData
		emby.Event = "playback.pause"
		data, handled, ignored := emby.ToWebhookData()
		if !ignored {
			t.Errorf("expected ignored true, got false")
		}
		if handled || data != nil {
			t.Errorf("expected handled false and data nil")
		}
	})
}

// Mock webhook service to test router dispatching
type mockWebhookIngester struct {
	ingestedData *WebhookData
}

func (m *mockWebhookIngester) Ingest(data WebhookData) error {
	m.ingestedData = &data
	return nil
}

func TestWebhookRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	webhookKey := uuid.New()
	cfg := &config.ServerConfig{
		JELLYFIN_WEBHOOK_ENABLED: true,
		JELLYFIN_WEBHOOK_KEY:     webhookKey,
	}

	engine := gin.New()
	br := &router.BaseRouter{
		Router: &engine.RouterGroup,
		Cfg:    cfg,
	}

	whService := &WebhookService{
		cfg: cfg,
	}

	r := NewRouter(br, nil, nil, whService)
	r.AddRoutes()

	t.Run("Reject when webhook disabled", func(t *testing.T) {
		cfg.JELLYFIN_WEBHOOK_ENABLED = false
		defer func() { cfg.JELLYFIN_WEBHOOK_ENABLED = true }()

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/jellyfin/webhook/"+webhookKey.String(), nil)
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400 when disabled, got %d", w.Code)
		}
	})

	t.Run("Reject with invalid key", func(t *testing.T) {
		wrongKey := uuid.New().String()
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/emby/webhook/"+wrongKey, nil)
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400 for wrong key, got %d", w.Code)
		}
	})

	t.Run("Accept Emby Ignored Event with 200 OK", func(t *testing.T) {
		payload := map[string]any{
			"Event": "playback.pause",
			"User": map[string]string{
				"Id": "u1",
			},
			"Item": map[string]any{
				"Name": "Show",
				"Type": "Episode",
			},
		}
		body, _ := json.Marshal(payload)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/emby/webhook/"+webhookKey.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200 OK for ignored event, got %d", w.Code)
		}
	})
}
