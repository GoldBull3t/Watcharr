package calendar_test

import (
	"testing"
	"time"

	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/calendar"
	"github.com/sbondCo/Watcharr/internal/testutil"
)

func TestCalendarGetEvents_Library(t *testing.T) {
	db := testutil.SetupDB(t)
	cfg := &config.ServerConfig{}
	svc := calendar.NewService(db, cfg, nil)

	// Create test user
	user := entity.User{
		Username: "calendaruser",
		Type:     entity.WATCHARR_USER,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Movie releasing on 2026-10-15
	relDate := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
	movieInWindow := entity.Content{
		TmdbID:      101,
		Title:       "October Movie",
		Type:        entity.MOVIE,
		ReleaseDate: &relDate,
	}
	if err := db.Create(&movieInWindow).Error; err != nil {
		t.Fatalf("failed to create movie: %v", err)
	}

	// Movie releasing on 2026-11-20 (outside October window)
	relDateNov := time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC)
	movieOutWindow := entity.Content{
		TmdbID:      102,
		Title:       "November Movie",
		Type:        entity.MOVIE,
		ReleaseDate: &relDateNov,
	}
	if err := db.Create(&movieOutWindow).Error; err != nil {
		t.Fatalf("failed to create november movie: %v", err)
	}

	// Add watched records
	w1 := entity.Watched{
		UserID:    user.ID,
		ContentID: &movieInWindow.ID,
		Status:    entity.PLANNED,
	}
	w2 := entity.Watched{
		UserID:    user.ID,
		ContentID: &movieOutWindow.ID,
		Status:    entity.PLANNED,
	}
	if err := db.Create(&w1).Error; err != nil {
		t.Fatalf("failed to create watched w1: %v", err)
	}
	if err := db.Create(&w2).Error; err != nil {
		t.Fatalf("failed to create watched w2: %v", err)
	}

	// Query October
	events, err := svc.GetEvents(user.ID, domain.CalendarRequest{
		Start:  "2026-10-01",
		End:    "2026-10-31",
		Filter: domain.CalendarFilterLibrary,
		Type:   domain.CalendarTypeAll,
	})
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event in October window, got %d", len(events))
	}

	ev := events[0]
	if ev.MediaID != 101 {
		t.Errorf("expected mediaId 101, got %d", ev.MediaID)
	}
	if ev.ReleaseDate != "2026-10-15" {
		t.Errorf("expected release date 2026-10-15, got %s", ev.ReleaseDate)
	}
	if !ev.InLibrary {
		t.Errorf("expected InLibrary true, got false")
	}
	if ev.Status != "PLANNED" {
		t.Errorf("expected status PLANNED, got %s", ev.Status)
	}

	// Query with TV filter (should return 0 because it's a movie)
	tvEvents, err := svc.GetEvents(user.ID, domain.CalendarRequest{
		Start:  "2026-10-01",
		End:    "2026-10-31",
		Filter: domain.CalendarFilterLibrary,
		Type:   domain.CalendarTypeTV,
	})
	if err != nil {
		t.Fatalf("GetEvents with TV filter failed: %v", err)
	}
	if len(tvEvents) != 0 {
		t.Errorf("expected 0 TV events, got %d", len(tvEvents))
	}
}

