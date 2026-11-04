package season

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/internal/testutil"
	"github.com/sbondCo/Watcharr/media/tmdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type mockWatchedProvider struct {
	db *gorm.DB
}

func (m *mockWatchedProvider) GetWatchedItemById(userId uint, id uint) (entity.Watched, error) {
	var w entity.Watched
	err := m.db.Preload("Content").Where("user_id = ? AND id = ?", userId, id).First(&w).Error
	return w, err
}

func (m *mockWatchedProvider) IsWatchedItemContentType(userId uint, id uint, ct entity.ContentType) error {
	return nil
}

type mockUserProvider struct{}

func (m *mockUserProvider) UserGetSettings(userId uint) (entity.UserSettings, error) {
	automate := true
	return entity.UserSettings{
		AutomateShowStatuses: &automate,
	}, nil
}

type roundTripFunc func(req *http.Request) *http.Response

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func TestSeasonChangeHook_PlanNextSeasonWhenCurrentCompleted(t *testing.T) {
	db := testutil.SetupDB(t)

	// Create user & content & watched entry
	user := entity.User{Username: "testuser"}
	require.NoError(t, db.Create(&user).Error)

	content := entity.Content{
		TmdbID:          1001,
		Type:            entity.SHOW,
		Title:           "Test Series",
		NumberOfSeasons: 2,
	}
	require.NoError(t, db.Create(&content).Error)

	watched := entity.Watched{
		UserID:    user.ID,
		ContentID: &content.ID,
		Content:   &content,
		Status:    entity.WATCHING,
	}
	require.NoError(t, db.Create(&watched).Error)

	wp := &mockWatchedProvider{db: db}
	up := &mockUserProvider{}

	tmdbInstance := tmdb.NewTMDB("fake_key")
	mockTransport := roundTripFunc(func(req *http.Request) *http.Response {
		urlStr := req.URL.Path
		var body string
		switch urlStr {
		case "/3/tv/1001":
			body = `{"id":1001,"number_of_seasons":2,"status":"Returning Series"}`
		case "/3/tv/1001/season/1":
			body = `{"id":1,"season_number":1,"episodes":[{"episode_number":1},{"episode_number":2}]}`
		case "/3/tv/1001/season/2":
			body = `{"id":2,"season_number":2,"episodes":[{"episode_number":1}]}`
		default:
			body = `{}`
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
			Header:     make(http.Header),
		}
	})

	httpClient := &http.Client{Transport: mockTransport}
	tmdbInstance.SetClient(httpClient)

	svc := NewService(db, wp, up, tmdbInstance)

	// Season 1 is marked as FINISHED
	resp, err := svc.SetWatchedSeason(user.ID, domain.WatchedSeasonSetRequest{
		WatchedID:    watched.ID,
		SeasonNumber: 1,
		Status:       entity.FINISHED,
	})
	require.NoError(t, err)

	// Verify Season 1 is finished
	assert.Equal(t, entity.FINISHED, resp.WatchedSeason.Status)
	assert.Equal(t, 1, resp.WatchedSeason.SeasonNumber)

	// Verify episodes of season 1 were automatically marked FINISHED
	var epCount int64
	db.Model(&entity.WatchedEpisode{}).
		Where("user_id = ? AND watched_id = ? AND season_number = ? AND status = ?",
			user.ID, watched.ID, 1, entity.FINISHED).
		Count(&epCount)
	assert.Equal(t, int64(2), epCount)

	// Verify Season 2 was automatically created as PLANNED
	require.NotNil(t, resp.NextSeason)
	assert.Equal(t, 2, resp.NextSeason.SeasonNumber)
	assert.Equal(t, entity.PLANNED, resp.NextSeason.Status)

	var nextSeasonInDB entity.WatchedSeason
	err = db.Where("user_id = ? AND watched_id = ? AND season_number = ?", user.ID, watched.ID, 2).
		First(&nextSeasonInDB).Error
	require.NoError(t, err)
	assert.Equal(t, entity.PLANNED, nextSeasonInDB.Status)

	// Verify overall show status transitioned from WATCHING to PLANNED
	assert.Equal(t, entity.PLANNED, resp.StatusChangedHookResponse.NewShowStatus)

	var updatedWatched entity.Watched
	err = db.Where("id = ?", watched.ID).First(&updatedWatched).Error
	require.NoError(t, err)
	assert.Equal(t, entity.PLANNED, updatedWatched.Status)

	// Now start watching Season 2 -> show should transition back to WATCHING!
	resp2, err := svc.SetWatchedSeason(user.ID, domain.WatchedSeasonSetRequest{
		WatchedID:    watched.ID,
		SeasonNumber: 2,
		Status:       entity.WATCHING,
	})
	require.NoError(t, err)
	assert.Equal(t, entity.WATCHING, resp2.StatusChangedHookResponse.NewShowStatus)

	err = db.Where("id = ?", watched.ID).First(&updatedWatched).Error
	require.NoError(t, err)
	assert.Equal(t, entity.WATCHING, updatedWatched.Status)
}

func TestSeasonChangeHook_ShowFinishedWhenAllSeasonsEnded(t *testing.T) {
	db := testutil.SetupDB(t)

	user := entity.User{Username: "testuser2"}
	require.NoError(t, db.Create(&user).Error)

	content := entity.Content{
		TmdbID:          1002,
		Type:            entity.SHOW,
		Title:           "Mini Series",
		NumberOfSeasons: 1,
	}
	require.NoError(t, db.Create(&content).Error)

	watched := entity.Watched{
		UserID:    user.ID,
		ContentID: &content.ID,
		Content:   &content,
		Status:    entity.WATCHING,
	}
	require.NoError(t, db.Create(&watched).Error)

	wp := &mockWatchedProvider{db: db}
	up := &mockUserProvider{}

	tmdbInstance := tmdb.NewTMDB("fake_key")
	mockTransport := roundTripFunc(func(req *http.Request) *http.Response {
		urlStr := req.URL.Path
		var body string
		switch urlStr {
		case "/3/tv/1002":
			body = `{"id":1002,"number_of_seasons":1,"status":"Ended"}`
		case "/3/tv/1002/season/1":
			body = `{"id":1,"season_number":1,"episodes":[{"episode_number":1}]}`
		default:
			body = `{}`
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
			Header:     make(http.Header),
		}
	})

	httpClient := &http.Client{Transport: mockTransport}
	tmdbInstance.SetClient(httpClient)

	svc := NewService(db, wp, up, tmdbInstance)

	resp, err := svc.SetWatchedSeason(user.ID, domain.WatchedSeasonSetRequest{
		WatchedID:    watched.ID,
		SeasonNumber: 1,
		Status:       entity.FINISHED,
	})
	require.NoError(t, err)

	// Overall show status should transition to FINISHED
	assert.Equal(t, entity.FINISHED, resp.StatusChangedHookResponse.NewShowStatus)

	var updatedWatched entity.Watched
	err = db.Where("id = ?", watched.ID).First(&updatedWatched).Error
	require.NoError(t, err)
	assert.Equal(t, entity.FINISHED, updatedWatched.Status)
}
