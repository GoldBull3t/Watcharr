package calendar_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/calendar"
	"github.com/sbondCo/Watcharr/internal/testutil"
)

func TestCalendarGetEvents_Library(t *testing.T) {
	calendar.CalendarStore.Flush()
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

func TestCalendarGetEvents_Pruning(t *testing.T) {
	calendar.CalendarStore.Flush()
	db := testutil.SetupDB(t)
	cfg := &config.ServerConfig{}
	svc := calendar.NewService(db, cfg, nil)

	user := entity.User{
		Username: "pruneuser",
		Type:     entity.WATCHARR_USER,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// 1. Ended show marked FINISHED
	oldDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	showEnded := entity.Content{
		TmdbID:      201,
		Title:       "Ended Show",
		Type:        entity.SHOW,
		Status:      "Ended",
		ReleaseDate: &oldDate,
	}
	db.Create(&showEnded)
	db.Create(&entity.Watched{
		UserID:    user.ID,
		ContentID: &showEnded.ID,
		Status:    entity.FINISHED,
	})

	// 2. Dropped show
	showDropped := entity.Content{
		TmdbID:      202,
		Title:       "Dropped Show",
		Type:        entity.SHOW,
		Status:      "Returning Series",
		ReleaseDate: &oldDate,
	}
	db.Create(&showDropped)
	db.Create(&entity.Watched{
		UserID:    user.ID,
		ContentID: &showDropped.ID,
		Status:    entity.DROPPED,
	})

	events, err := svc.GetEvents(user.ID, domain.CalendarRequest{
		Start:  "2026-10-01",
		End:    "2026-10-31",
		Filter: domain.CalendarFilterLibrary,
		Type:   domain.CalendarTypeAll,
	})
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}

	// Both should be excluded / produce 0 events
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func BenchmarkCalendarGetEvents_LargeLibrary(b *testing.B) {
	db := testutil.SetupDB(&testing.T{})
	cfg := &config.ServerConfig{}
	svc := calendar.NewService(db, cfg, nil)

	user := entity.User{
		Username: "benchuser",
		Type:     entity.WATCHARR_USER,
	}
	db.Create(&user)

	oldDate := time.Date(2021, 5, 1, 0, 0, 0, 0, time.UTC)
	for i := 1; i <= 100; i++ {
		c := entity.Content{
			TmdbID:      1000 + i,
			Title:       fmt.Sprintf("Ended Show %d", i),
			Type:        entity.SHOW,
			Status:      "Ended",
			ReleaseDate: &oldDate,
		}
		db.Create(&c)
		db.Create(&entity.Watched{
			UserID:    user.ID,
			ContentID: &c.ID,
			Status:    entity.FINISHED,
		})
	}

	for i := 1; i <= 50; i++ {
		rel := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)
		c := entity.Content{
			TmdbID:      2000 + i,
			Title:       fmt.Sprintf("October Movie %d", i),
			Type:        entity.MOVIE,
			ReleaseDate: &rel,
		}
		db.Create(&c)
		db.Create(&entity.Watched{
			UserID:    user.ID,
			ContentID: &c.ID,
			Status:    entity.WATCHING,
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calendar.CalendarStore.Flush()
		_, err := svc.GetEvents(user.ID, domain.CalendarRequest{
			Start:  "2026-10-01",
			End:    "2026-10-31",
			Filter: domain.CalendarFilterLibrary,
			Type:   domain.CalendarTypeAll,
		})
		if err != nil {
			b.Fatalf("GetEvents failed: %v", err)
		}
	}
}
