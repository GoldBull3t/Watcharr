package watched

import (
	"testing"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/internal/testutil"
)

func TestUpdateWatched_AtomicTransaction(t *testing.T) {
	testutil.SetupLogging()
	db := testutil.SetupDB(t)

	user := entity.User{
		Username: "tx_tester",
		Type:     entity.WATCHARR_USER,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	showContent := entity.Content{
		TmdbID: 99901,
		Title:  "Breaking Code",
		Type:   entity.SHOW,
	}
	if err := db.Create(&showContent).Error; err != nil {
		t.Fatalf("failed to create content: %v", err)
	}

	showContentID := int(showContent.ID)
	watched := entity.Watched{
		UserID:    user.ID,
		ContentID: &showContentID,
		Status:    entity.WATCHING,
		Rating:    0,
	}
	if err := db.Create(&watched).Error; err != nil {
		t.Fatalf("failed to create watched entry: %v", err)
	}

	// Create 2 seasons and 3 episodes
	s1 := entity.WatchedSeason{
		UserID:       user.ID,
		WatchedID:    watched.ID,
		SeasonNumber: 1,
		Status:       entity.WATCHING,
	}
	s2 := entity.WatchedSeason{
		UserID:       user.ID,
		WatchedID:    watched.ID,
		SeasonNumber: 2,
		Status:       entity.WATCHING,
	}
	db.Create(&s1)
	db.Create(&s2)

	ep1 := entity.WatchedEpisode{
		UserID:        user.ID,
		WatchedID:     watched.ID,
		SeasonNumber:  1,
		EpisodeNumber: 1,
		Status:        entity.FINISHED,
	}
	ep2 := entity.WatchedEpisode{
		UserID:        user.ID,
		WatchedID:     watched.ID,
		SeasonNumber:  1,
		EpisodeNumber: 2,
		Status:        entity.WATCHING,
	}
	ep3 := entity.WatchedEpisode{
		UserID:        user.ID,
		WatchedID:     watched.ID,
		SeasonNumber:  2,
		EpisodeNumber: 1,
		Status:        entity.WATCHING,
	}
	db.Create(&ep1)
	db.Create(&ep2)
	db.Create(&ep3)

	s := NewService(db, nil, nil, nil, nil)

	// Update watched to FINISHED - this must atomically update watched + seasons + episodes + activities
	updateResp, err := s.UpdateWatched(user.ID, watched.ID, domain.WatchedUpdateRequest{
		Status: entity.FINISHED,
		Rating: 9.5,
	})
	if err != nil {
		t.Fatalf("UpdateWatched failed: %v", err)
	}

	if len(updateResp.NewActivities) == 0 {
		t.Fatalf("expected new activities generated, got 0")
	}

	// Verify watched entry is FINISHED with rating 9.5
	var updatedWatched entity.Watched
	if err := db.Where("id = ?", watched.ID).First(&updatedWatched).Error; err != nil {
		t.Fatalf("failed to fetch updated watched: %v", err)
	}
	if updatedWatched.Status != entity.FINISHED {
		t.Fatalf("expected status FINISHED, got %s", updatedWatched.Status)
	}
	if updatedWatched.Rating != 9.5 {
		t.Fatalf("expected rating 9.5, got %f", updatedWatched.Rating)
	}

	// Verify all seasons were updated to FINISHED
	var seasons []entity.WatchedSeason
	if err := db.Where("watched_id = ?", watched.ID).Find(&seasons).Error; err != nil {
		t.Fatalf("failed to query seasons: %v", err)
	}
	for _, season := range seasons {
		if season.Status != entity.FINISHED {
			t.Fatalf("expected season %d status FINISHED, got %s", season.SeasonNumber, season.Status)
		}
	}

	// Verify all episodes were updated to FINISHED
	var episodes []entity.WatchedEpisode
	if err := db.Where("watched_id = ?", watched.ID).Find(&episodes).Error; err != nil {
		t.Fatalf("failed to query episodes: %v", err)
	}
	for _, episode := range episodes {
		if episode.Status != entity.FINISHED {
			t.Fatalf("expected episode s%de%d status FINISHED, got %s", episode.SeasonNumber, episode.EpisodeNumber, episode.Status)
		}
	}

	// Verify activities were created
	var activities []entity.Activity
	if err := db.Where("watched_id = ?", watched.ID).Find(&activities).Error; err != nil {
		t.Fatalf("failed to query activities: %v", err)
	}
	if len(activities) < 2 { // At least STATUS_CHANGED and RATING_CHANGED
		t.Fatalf("expected at least 2 activities, got %d", len(activities))
	}
}

func TestRemoveWatched_AtomicTransaction(t *testing.T) {
	testutil.SetupLogging()
	db := testutil.SetupDB(t)

	user := entity.User{
		Username: "rm_tester",
		Type:     entity.WATCHARR_USER,
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	movieContent := entity.Content{
		TmdbID: 99902,
		Title:  "Inception Test",
		Type:   entity.MOVIE,
	}
	if err := db.Create(&movieContent).Error; err != nil {
		t.Fatalf("failed to create content: %v", err)
	}

	movieContentID := int(movieContent.ID)
	watched := entity.Watched{
		UserID:    user.ID,
		ContentID: &movieContentID,
		Status:    entity.FINISHED,
	}
	if err := db.Create(&watched).Error; err != nil {
		t.Fatalf("failed to create watched: %v", err)
	}

	s := NewService(db, nil, nil, nil, nil)

	rmResp, err := s.removeWatched(user.ID, watched.ID)
	if err != nil {
		t.Fatalf("removeWatched failed: %v", err)
	}
	if rmResp.NewActivity.Type != entity.REMOVED_WATCHED {
		t.Fatalf("expected activity type REMOVED_WATCHED, got %s", rmResp.NewActivity.Type)
	}

	// Verify watched is soft deleted
	var checkWatched entity.Watched
	res := db.Where("id = ?", watched.ID).First(&checkWatched)
	if res.Error == nil {
		t.Fatalf("expected watched to be soft-deleted and not found, but it was found")
	}

	// Verify unscoped find retrieves it with DeletedAt set
	var unscopedWatched entity.Watched
	if err := db.Unscoped().Where("id = ?", watched.ID).First(&unscopedWatched).Error; err != nil {
		t.Fatalf("failed to find soft deleted watched: %v", err)
	}
	if !unscopedWatched.DeletedAt.Valid {
		t.Fatalf("expected DeletedAt to be valid")
	}

	// Verify REMOVED_WATCHED activity exists in DB
	var rmAct entity.Activity
	if err := db.Where("watched_id = ? AND type = ?", watched.ID, entity.REMOVED_WATCHED).First(&rmAct).Error; err != nil {
		t.Fatalf("failed to find REMOVED_WATCHED activity: %v", err)
	}
}

