package profile

import (
	"testing"
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/internal/testutil"
)

func TestProfileService(t *testing.T) {
	db := testutil.SetupDB(t)
	svc := NewService(db)

	// Create test user
	user := entity.User{
		Username: "cinematic_user",
		Type:     entity.WATCHARR_USER,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	// Create Movie (120 min)
	movieDate := time.Date(2023, 5, 10, 0, 0, 0, 0, time.UTC)
	movie := entity.Content{
		TmdbID:      1001,
		Title:       "Oppenheimer",
		Type:        entity.MOVIE,
		Runtime:     180,
		ReleaseDate: &movieDate,
		Genres:      "Drama, History",
	}
	if err := db.Create(&movie).Error; err != nil {
		t.Fatalf("failed to create movie: %v", err)
	}

	// Create TV Show (45 min per episode, 10 episodes total)
	showDate := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	show := entity.Content{
		TmdbID:           2001,
		Title:            "Severance",
		Type:             entity.SHOW,
		Runtime:          50,
		NumberOfEpisodes: 9,
		ReleaseDate:      &showDate,
		Genres:           "Sci-Fi, Thriller",
	}
	if err := db.Create(&show).Error; err != nil {
		t.Fatalf("failed to create show: %v", err)
	}

	// Watched Movie (Finished, Rated 9.0)
	watchedMovie := entity.Watched{
		UserID:    user.ID,
		ContentID: &movie.ID,
		Status:    entity.FINISHED,
		Rating:    9.0,
		Activity: []entity.Activity{
			{
				UserID:      user.ID,
				Type:        entity.STATUS_CHANGED,
				Data:        "FINISHED",
				CountAsPlay: true,
			},
		},
	}
	if err := db.Create(&watchedMovie).Error; err != nil {
		t.Fatalf("failed to create watched movie: %v", err)
	}

	// Watched Show (Watching, 4 episodes watched)
	watchedShow := entity.Watched{
		UserID:    user.ID,
		ContentID: &show.ID,
		Status:    entity.WATCHING,
		Rating:    8.5,
		WatchedEpisodes: []entity.WatchedEpisode{
			{UserID: user.ID, SeasonNumber: 1, EpisodeNumber: 1, Status: entity.FINISHED},
			{UserID: user.ID, SeasonNumber: 1, EpisodeNumber: 2, Status: entity.FINISHED},
			{UserID: user.ID, SeasonNumber: 1, EpisodeNumber: 3, Status: entity.FINISHED},
			{UserID: user.ID, SeasonNumber: 1, EpisodeNumber: 4, Status: entity.FINISHED},
		},
	}
	if err := db.Create(&watchedShow).Error; err != nil {
		t.Fatalf("failed to create watched show: %v", err)
	}

	prof, err := svc.getProfile(user.ID)
	if err != nil {
		t.Fatalf("getProfile error: %v", err)
	}

	if prof.MoviesWatched != 1 {
		t.Errorf("expected 1 movie watched, got %d", prof.MoviesWatched)
	}
	if prof.MoviesWatchedRuntime != 180 {
		t.Errorf("expected 180m movie runtime, got %d", prof.MoviesWatchedRuntime)
	}
	if prof.ShowsWatched != 0 {
		t.Errorf("expected 0 finished shows, got %d", prof.ShowsWatched)
	}
	if prof.ShowsWatching != 1 {
		t.Errorf("expected 1 watching show, got %d", prof.ShowsWatching)
	}
	if prof.EpisodesWatched != 4 {
		t.Errorf("expected 4 episodes watched, got %d", prof.EpisodesWatched)
	}
	// 4 episodes * 50 min = 200 min
	if prof.ShowsWatchedRuntime != 200 {
		t.Errorf("expected 200m show runtime, got %d", prof.ShowsWatchedRuntime)
	}
	if prof.TotalRuntime != 380 {
		t.Errorf("expected 380m total runtime, got %d", prof.TotalRuntime)
	}
	if prof.RatingStats.Count != 2 {
		t.Errorf("expected 2 ratings, got %d", prof.RatingStats.Count)
	}
	if prof.RatingStats.Average != 8.8 {
		t.Errorf("expected 8.8 average rating, got %.1f", prof.RatingStats.Average)
	}
	if prof.StatusDistribution["FINISHED"] != 1 {
		t.Errorf("expected 1 FINISHED status, got %d", prof.StatusDistribution["FINISHED"])
	}
	if prof.StatusDistribution["WATCHING"] != 1 {
		t.Errorf("expected 1 WATCHING status, got %d", prof.StatusDistribution["WATCHING"])
	}
	if len(prof.GenreDistribution) == 0 {
		t.Errorf("expected genre distribution to have items")
	}
}
