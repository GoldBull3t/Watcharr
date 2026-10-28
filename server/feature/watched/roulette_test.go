package watched

import (
	"testing"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/internal/testutil"
)

func TestPickRoulette(t *testing.T) {
	testutil.SetupLogging()
	db := testutil.SetupDB(t)

	user := entity.User{
		Username: "roulette_tester",
		Type:     entity.WATCHARR_USER,
	}
	db.Create(&user)

	// Create test movies
	shortMovie := entity.Content{
		TmdbID:      20001,
		Title:       "Quick Comedy",
		Type:        entity.MOVIE,
		Runtime:     85,
		VoteAverage: 7.5,
		Genres:      "Comedy, Animation",
	}
	db.Create(&shortMovie)

	longMovie := entity.Content{
		TmdbID:      20002,
		Title:       "Epic Sci-Fi",
		Type:        entity.MOVIE,
		Runtime:     165,
		VoteAverage: 8.8,
		Genres:      "Science Fiction, Adventure",
	}
	db.Create(&longMovie)

	tvShow := entity.Content{
		TmdbID:      20003,
		Title:       "Mystery Miniseries",
		Type:        entity.SHOW,
		VoteAverage: 8.1,
		Genres:      "Mystery, Drama",
	}
	db.Create(&tvShow)

	shortMovieID := int(shortMovie.ID)
	longMovieID := int(longMovie.ID)
	tvShowID := int(tvShow.ID)

	// Add to user's planned list
	db.Create(&entity.Watched{
		UserID:    user.ID,
		ContentID: &shortMovieID,
		Status:    entity.PLANNED,
		Rating:    0,
	})
	db.Create(&entity.Watched{
		UserID:    user.ID,
		ContentID: &longMovieID,
		Status:    entity.PLANNED,
		Rating:    0,
	})
	db.Create(&entity.Watched{
		UserID:    user.ID,
		ContentID: &tvShowID,
		Status:    entity.PLANNED,
		Rating:    0,
	})

	s := NewService(db, nil, nil, nil, nil)

	// 1. Pick with maxRuntime <= 90min (should only pick Quick Comedy)
	res1, err := s.PickRoulette(user.ID, domain.RoulettePickRequest{
		Type:       "movie",
		Status:     "planned",
		MaxRuntime: 90,
		Count:      1,
	})
	if err != nil {
		t.Fatalf("PickRoulette failed: %v", err)
	}
	if res1.TotalPool != 1 {
		t.Fatalf("expected totalPool 1 for runtime<=90, got %d", res1.TotalPool)
	}
	if len(res1.Items) != 1 || res1.Items[0].Name != "Quick Comedy" {
		t.Fatalf("expected 'Quick Comedy', got %+v", res1.Items)
	}

	// 2. Pick with Genre "Science Fiction"
	res2, err := s.PickRoulette(user.ID, domain.RoulettePickRequest{
		Type:   "movie",
		Status: "planned",
		Genre:  "science fiction",
	})
	if err != nil {
		t.Fatalf("PickRoulette failed: %v", err)
	}
	if res2.TotalPool != 1 || len(res2.Items) != 1 || res2.Items[0].Name != "Epic Sci-Fi" {
		t.Fatalf("expected 'Epic Sci-Fi', got %+v", res2.Items)
	}

	// 3. Pick TV shows
	res3, err := s.PickRoulette(user.ID, domain.RoulettePickRequest{
		Type:   "tv",
		Status: "planned",
	})
	if err != nil {
		t.Fatalf("PickRoulette failed: %v", err)
	}
	if res3.TotalPool != 1 || res3.Items[0].Name != "Mystery Miniseries" {
		t.Fatalf("expected 'Mystery Miniseries', got %+v", res3.Items)
	}

	// 4. Min rating filter >= 8.5 (only Epic Sci-Fi has 8.8)
	res4, err := s.PickRoulette(user.ID, domain.RoulettePickRequest{
		Type:      "movie",
		Status:    "planned",
		MinRating: 8.5,
	})
	if err != nil {
		t.Fatalf("PickRoulette failed: %v", err)
	}
	if res4.TotalPool != 1 || res4.Items[0].Name != "Epic Sci-Fi" {
		t.Fatalf("expected 'Epic Sci-Fi', got %+v", res4.Items)
	}
}
