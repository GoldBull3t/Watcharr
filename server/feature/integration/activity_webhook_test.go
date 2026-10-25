package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/feature/integration"
	"github.com/sbondCo/Watcharr/internal/testutil"
)

func TestDispatchActivityWebhook_Generic(t *testing.T) {
	os.Setenv("ALLOW_LOCAL_WEBHOOKS", "true")
	defer os.Unsetenv("ALLOW_LOCAL_WEBHOOKS")

	received := make(chan integration.ActivityWebhookPayload, 1)

	// Mock webhook receiver (Home Assistant simulation)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		var p integration.ActivityWebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			t.Errorf("failed to decode webhook body: %v", err)
		}
		received <- p
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	db := testutil.SetupDB(t)

	enabled := true
	webhookUrl := server.URL
	filter := "all"

	user := entity.User{
		Username:               "homelabber",
		Type:                   entity.WATCHARR_USER,
		ActivityWebhookEnabled: &enabled,
		ActivityWebhookUrl:     &webhookUrl,
		ActivityWebhookFilter:  &filter,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	payload := integration.ActivityWebhookPayload{
		Event:     "media.finished",
		Status:    "FINISHED",
		MediaType: "movie",
		Title:     "Dune: Part Two",
		Rating:    5.0,
		Poster:    "https://image.tmdb.org/t/p/w500/dune2.jpg",
	}

	integration.DispatchActivityWebhook(db, user.ID, payload)

	select {
	case p := <-received:
		if p.Title != "Dune: Part Two" {
			t.Errorf("expected title 'Dune: Part Two', got '%s'", p.Title)
		}
		if p.Event != "media.finished" {
			t.Errorf("expected event 'media.finished', got '%s'", p.Event)
		}
		if p.Status != "FINISHED" {
			t.Errorf("expected status 'FINISHED', got '%s'", p.Status)
		}
		if p.User != "homelabber" {
			t.Errorf("expected user 'homelabber', got '%s'", p.User)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for webhook delivery")
	}
}

func TestDispatchActivityWebhook_FilterSkip(t *testing.T) {
	os.Setenv("ALLOW_LOCAL_WEBHOOKS", "true")
	defer os.Unsetenv("ALLOW_LOCAL_WEBHOOKS")

	called := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	db := testutil.SetupDB(t)

	enabled := true
	webhookUrl := server.URL
	filter := "finished" // only send finished events

	user := entity.User{
		Username:               "filteruser",
		Type:                   entity.WATCHARR_USER,
		ActivityWebhookEnabled: &enabled,
		ActivityWebhookUrl:     &webhookUrl,
		ActivityWebhookFilter:  &filter,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Dispatch watching event (should be ignored by filter)
	payload := integration.ActivityWebhookPayload{
		Event:     "media.watching",
		Status:    "WATCHING",
		MediaType: "tv",
		Title:     "Shōgun",
	}

	integration.DispatchActivityWebhook(db, user.ID, payload)

	time.Sleep(200 * time.Millisecond)
	if called {
		t.Errorf("webhook was called, but should have been filtered out")
	}
}

