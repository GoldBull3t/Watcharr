package integration_test

import (
	"strings"
	"testing"
	"time"

	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/feature/calendar"
	"github.com/sbondCo/Watcharr/feature/integration"
	"github.com/sbondCo/Watcharr/internal/testutil"
)

func TestGetCalendarICS_WithPoster(t *testing.T) {
	calendar.CalendarStore.Flush()
	db := testutil.SetupDB(t)
	cfg := &config.ServerConfig{}
	calSvc := calendar.NewService(db, cfg, nil)
	intSvc := integration.NewService(db, cfg, nil, calSvc)

	user := entity.User{
		Username: "posteruser",
		Type:     entity.WATCHARR_USER,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Movie releasing tomorrow (well within the -3 to +30 day window)
	relDate := time.Now().AddDate(0, 0, 1)
	movie := entity.Content{
		TmdbID:      501,
		Title:       "Filme Com Poster",
		Type:        entity.MOVIE,
		PosterPath:  "/poster123.jpg",
		Overview:    "Sinopse incrível do filme.",
		ReleaseDate: &relDate,
	}
	if err := db.Create(&movie).Error; err != nil {
		t.Fatalf("failed to create movie: %v", err)
	}

	w := entity.Watched{
		UserID:    user.ID,
		ContentID: &movie.ID,
		Status:    entity.PLANNED,
	}
	if err := db.Create(&w).Error; err != nil {
		t.Fatalf("failed to create watched item: %v", err)
	}

	ics, err := intSvc.GetCalendarICS(user.ID)
	if err != nil {
		t.Fatalf("GetCalendarICS failed: %v", err)
	}

	expectedPosterURL := "https://image.tmdb.org/t/p/w500/poster123.jpg"

	// Verify standard VCALENDAR envelope
	if !strings.HasPrefix(ics, "BEGIN:VCALENDAR\r\n") {
		t.Errorf("expected ics to start with BEGIN:VCALENDAR, got:\n%s", ics)
	}
	if !strings.HasSuffix(ics, "END:VCALENDAR\r\n") {
		t.Errorf("expected ics to end with END:VCALENDAR, got:\n%s", ics)
	}

	// Verify standard ATTACH property with image/jpeg FMTTYPE
	expectedAttach := "ATTACH;FMTTYPE=image/jpeg:" + expectedPosterURL + "\r\n"
	if !strings.Contains(ics, expectedAttach) {
		t.Errorf("expected ics to contain %q, but not found. Full ICS:\n%s", expectedAttach, ics)
	}

	// Verify custom X-POSTER property
	expectedXPoster := "X-POSTER:" + expectedPosterURL + "\r\n"
	if !strings.Contains(ics, expectedXPoster) {
		t.Errorf("expected ics to contain %q, but not found. Full ICS:\n%s", expectedXPoster, ics)
	}

	// Verify DESCRIPTION includes poster suffix escaped per RFC 5545
	expectedDescSnippet := "\\nPoster: " + expectedPosterURL
	if !strings.Contains(ics, expectedDescSnippet) {
		t.Errorf("expected ics DESCRIPTION to contain %q, but not found. Full ICS:\n%s", expectedDescSnippet, ics)
	}
}

func TestGetCalendarICS_WithoutPoster(t *testing.T) {
	calendar.CalendarStore.Flush()
	db := testutil.SetupDB(t)
	cfg := &config.ServerConfig{}
	calSvc := calendar.NewService(db, cfg, nil)
	intSvc := integration.NewService(db, cfg, nil, calSvc)

	user := entity.User{
		Username: "noposteruser",
		Type:     entity.WATCHARR_USER,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Movie releasing tomorrow without poster
	relDate := time.Now().AddDate(0, 0, 2)
	movie := entity.Content{
		TmdbID:      502,
		Title:       "Filme Sem Poster",
		Type:        entity.MOVIE,
		PosterPath:  "",
		Overview:    "Sinopse sem poster.",
		ReleaseDate: &relDate,
	}
	if err := db.Create(&movie).Error; err != nil {
		t.Fatalf("failed to create movie: %v", err)
	}

	w := entity.Watched{
		UserID:    user.ID,
		ContentID: &movie.ID,
		Status:    entity.PLANNED,
	}
	if err := db.Create(&w).Error; err != nil {
		t.Fatalf("failed to create watched item: %v", err)
	}

	ics, err := intSvc.GetCalendarICS(user.ID)
	if err != nil {
		t.Fatalf("GetCalendarICS failed: %v", err)
	}

	// Verify event exists
	if !strings.Contains(ics, "SUMMARY:Filme Sem Poster") {
		t.Errorf("expected event to be created, but summary not found. Full ICS:\n%s", ics)
	}

	// Verify NO ATTACH, NO X-POSTER, NO Poster: in description
	if strings.Contains(ics, "ATTACH") {
		t.Errorf("did not expect ATTACH field when poster is empty, but found in:\n%s", ics)
	}
	if strings.Contains(ics, "X-POSTER") {
		t.Errorf("did not expect X-POSTER field when poster is empty, but found in:\n%s", ics)
	}
	if strings.Contains(ics, "Poster:") {
		t.Errorf("did not expect 'Poster:' in description when poster is empty, but found in:\n%s", ics)
	}
}

func TestGetCalendarICS_FullURLPoster(t *testing.T) {
	calendar.CalendarStore.Flush()
	db := testutil.SetupDB(t)
	cfg := &config.ServerConfig{}
	calSvc := calendar.NewService(db, cfg, nil)
	intSvc := integration.NewService(db, cfg, nil, calSvc)

	user := entity.User{
		Username: "fullurluser",
		Type:     entity.WATCHARR_USER,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	fullURL := "https://cdn.example.com/custom_poster.jpg"
	relDate := time.Now().AddDate(0, 0, 3)
	movie := entity.Content{
		TmdbID:      503,
		Title:       "Filme Com URL Completo",
		Type:        entity.MOVIE,
		PosterPath:  fullURL,
		Overview:    "Teste com URL externa.",
		ReleaseDate: &relDate,
	}
	if err := db.Create(&movie).Error; err != nil {
		t.Fatalf("failed to create movie: %v", err)
	}

	w := entity.Watched{
		UserID:    user.ID,
		ContentID: &movie.ID,
		Status:    entity.PLANNED,
	}
	if err := db.Create(&w).Error; err != nil {
		t.Fatalf("failed to create watched item: %v", err)
	}

	ics, err := intSvc.GetCalendarICS(user.ID)
	if err != nil {
		t.Fatalf("GetCalendarICS failed: %v", err)
	}

	expectedAttach := "ATTACH;FMTTYPE=image/jpeg:" + fullURL + "\r\n"
	if !strings.Contains(ics, expectedAttach) {
		t.Errorf("expected ics to contain %q, but not found. Full ICS:\n%s", expectedAttach, ics)
	}
	expectedXPoster := "X-POSTER:" + fullURL + "\r\n"
	if !strings.Contains(ics, expectedXPoster) {
		t.Errorf("expected ics to contain %q, but not found. Full ICS:\n%s", expectedXPoster, ics)
	}
}

