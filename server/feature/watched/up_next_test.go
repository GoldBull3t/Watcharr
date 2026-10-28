package watched

import (
	"encoding/json"
	"testing"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/internal/testutil"
	"github.com/sbondCo/Watcharr/media/tmdb"
)

type mockTMDBProvider struct{}

func (m *mockTMDBProvider) SeasonDetails(showId string, seasonNumber string, optParams ...map[string]string) (tmdb.SeasonDetails, error) {
	if seasonNumber == "1" {
		jsonData := `{
			"air_date": "2023-01-15",
			"episodes": [
				{"episode_number": 1, "name": "Pilot", "still_path": "/still1.jpg", "air_date": "2023-01-15"},
				{"episode_number": 2, "name": "Infected", "still_path": "/still2.jpg", "air_date": "2023-01-22"},
				{"episode_number": 3, "name": "Long Long Time", "still_path": "/still3.jpg", "air_date": "2023-01-29"}
			]
		}`
		var sd tmdb.SeasonDetails
		_ = json.Unmarshal([]byte(jsonData), &sd)
		return sd, nil
	}
	return tmdb.SeasonDetails{}, nil
}

func TestGetUpNext(t *testing.T) {
	testutil.SetupLogging()
	db := testutil.SetupDB(t)

	user := entity.User{
		Username: "up_next_tester",
		Type:     entity.WATCHARR_USER,
	}
	db.Create(&user)

	show := entity.Content{
		TmdbID:       10001,
		Title:        "The Last Show",
		Type:         entity.SHOW,
		PosterPath:   "/poster.jpg",
		BackdropPath: "/backdrop.jpg",
	}
	db.Create(&show)

	contentID := int(show.ID)
	w := entity.Watched{
		UserID:    user.ID,
		ContentID: &contentID,
		Status:    entity.WATCHING,
		WatchedEpisodes: []entity.WatchedEpisode{
			{
				UserID:        user.ID,
				SeasonNumber:  1,
				EpisodeNumber: 1,
				Status:        entity.FINISHED,
			},
		},
	}
	db.Create(&w)

	s := NewService(db, nil, nil, nil, &mockTMDBProvider{})

	items, err := s.GetUpNext(user.ID)
	if err != nil {
		t.Fatalf("GetUpNext failed: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 up-next item, got %d", len(items))
	}

	item := items[0]
	if item.ShowTitle != "The Last Show" {
		t.Errorf("expected title 'The Last Show', got %s", item.ShowTitle)
	}
	if item.SeasonNumber != 1 || item.EpisodeNumber != 2 {
		t.Errorf("expected S1E2, got S%dE%d", item.SeasonNumber, item.EpisodeNumber)
	}
	if item.EpisodeName != "Infected" {
		t.Errorf("expected episode name 'Infected', got %s", item.EpisodeName)
	}
	if item.EpisodeStill != "/still2.jpg" {
		t.Errorf("expected still '/still2.jpg', got %s", item.EpisodeStill)
	}
	if item.EpisodesRemaining != 2 {
		t.Errorf("expected 2 episodes remaining, got %d", item.EpisodesRemaining)
	}
}
