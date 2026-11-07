package watchedutil_test

import (
	"testing"

	"github.com/sbondCo/Watcharr/database/dbmodel"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/feature/watched/watchedutil"
)

func TestGetLatestWatchedInTv(t *testing.T) {
	watchedSeasons := []entity.WatchedSeason{
		{
			SeasonNumber: 1,
			Status:       entity.FINISHED,
		},
		{
			SeasonNumber: 2,
			Status:       entity.WATCHING,
		},
		{
			SeasonNumber: 3,
			Status:       entity.FINISHED,
		},
	}
	watchedEps := []entity.WatchedEpisode{
		{
			GormModelNoDel: dbmodel.GormModelNoDel{ID: 60},
			EpisodeNumber:  1,
			SeasonNumber:   1,
			Status:         entity.FINISHED,
		},
		{
			GormModelNoDel: dbmodel.GormModelNoDel{ID: 70},
			EpisodeNumber:  5,
			SeasonNumber:   2,
			Status:         entity.FINISHED,
		},
		{
			GormModelNoDel: dbmodel.GormModelNoDel{ID: 72},
			EpisodeNumber:  6,
			SeasonNumber:   3,
			Status:         entity.DROPPED,
		},
		{
			GormModelNoDel: dbmodel.GormModelNoDel{ID: 90},
			EpisodeNumber:  2,
			SeasonNumber:   3,
			Status:         entity.FINISHED,
		},
	}
	resp := watchedutil.GetLatestWatchedInTv(watchedSeasons, watchedEps)
	want := "S2E5"
	if resp != want {
		t.Errorf("%s != %s", resp, want)
	}

	// #753: When all watched seasons are FINISHED (no season WATCHING), show only "S3"
	completedSeasons := []entity.WatchedSeason{
		{SeasonNumber: 1, Status: entity.FINISHED},
		{SeasonNumber: 3, Status: entity.FINISHED},
	}
	respCompleted := watchedutil.GetLatestWatchedInTv(completedSeasons, watchedEps)
	if respCompleted != "S3" {
		t.Errorf("expected S3 for completed season, got %s", respCompleted)
	}

	// BUG-06: When S1 is FINISHED in WatchedSeasons and user watches episodes in S2 (without S2 in WatchedSeasons)
	// Also tests lexicographical comparison (S1E10 should NOT beat S2E3)
	s1FinishedOnly := []entity.WatchedSeason{
		{SeasonNumber: 1, Status: entity.FINISHED},
	}
	s2Eps := []entity.WatchedEpisode{
		{
			GormModelNoDel: dbmodel.GormModelNoDel{ID: 101},
			SeasonNumber:   1,
			EpisodeNumber:  10,
			Status:         entity.FINISHED,
		},
		{
			GormModelNoDel: dbmodel.GormModelNoDel{ID: 102},
			SeasonNumber:   2,
			EpisodeNumber:  3,
			Status:         entity.FINISHED,
		},
	}
	respNewerSeason := watchedutil.GetLatestWatchedInTv(s1FinishedOnly, s2Eps)
	if respNewerSeason != "S2E3" {
		t.Errorf("expected S2E3 for episode in newer season than finished season, got %s", respNewerSeason)
	}
}
