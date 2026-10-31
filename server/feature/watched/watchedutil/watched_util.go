package watchedutil

import (
	"strconv"

	"github.com/sbondCo/Watcharr/database/entity"
)

// Get biggest season watching or biggest season watched (#753).
// When a season is FINISHED (and no season/episode is actively WATCHING),
// returns a compact season indicator ("S1") instead of verbose "Season 1" or "S1E10".
func GetLatestWatchedInTv(
	ws []entity.WatchedSeason,
	we []entity.WatchedEpisode,
) string {
	if len(ws) <= 0 && len(we) <= 0 {
		return ""
	}

	seasonNum, seasonIsWatching := getLatestWatchedSeasonStateInTv(ws)
	episode := getLatestWatchedEpisodeInTv(we, seasonNum)

	if seasonNum > -1 && episode.ID != 0 {
		// If the season itself is WATCHING or the latest episode in it is WATCHING, show SxEy
		if seasonIsWatching || episode.Status == entity.WATCHING {
			return SeasonAndEpToReadable(seasonNum, episode.EpisodeNumber)
		}
		// Season is FINISHED: show only the season number ("S1") (#753)
		return "S" + strconv.Itoa(seasonNum)
	} else if seasonNum > -1 {
		// If we only have a season num
		return "S" + strconv.Itoa(seasonNum)
	} else if episode.ID != 0 {
		// If we only have an episode
		return SeasonAndEpToReadable(episode.SeasonNumber, episode.EpisodeNumber)
	}

	return ""
}

func getLatestWatchedSeasonStateInTv(ws []entity.WatchedSeason) (int, bool) {
	if len(ws) <= 0 {
		return -1, false
	}

	biggestWatched := -1
	biggestWatching := -1

	for i := range ws {
		v := &ws[i]
		switch v.Status {
		case entity.WATCHING:
			if v.SeasonNumber > biggestWatching {
				biggestWatching = v.SeasonNumber
			}
		case entity.FINISHED:
			if v.SeasonNumber > biggestWatched {
				biggestWatched = v.SeasonNumber
			}
		}
	}

	if biggestWatching >= 0 {
		return biggestWatching, true
	}

	return biggestWatched, false
}

func getLatestWatchedEpisodeInTv(
	we []entity.WatchedEpisode,
	seasonNum int,
) entity.WatchedEpisode {
	if len(we) <= 0 {
		return entity.WatchedEpisode{}
	}

	biggestWatchedIdx := -1
	biggestWatchingIdx := -1

	for i := range we {
		v := &we[i]

		if seasonNum >= 0 && v.SeasonNumber != seasonNum {
			// If we have a seasonNum, ensure the episode we find is in
			// that season.
			continue
		}

		switch v.Status {
		case entity.WATCHING:
			if biggestWatchingIdx == -1 {
				biggestWatchingIdx = i
				continue
			}
			if v.EpisodeNumber > we[biggestWatchingIdx].EpisodeNumber ||
				v.SeasonNumber > we[biggestWatchingIdx].SeasonNumber {
				biggestWatchingIdx = i
			}
		case entity.FINISHED:
			if biggestWatchedIdx == -1 {
				biggestWatchedIdx = i
				continue
			}
			if v.EpisodeNumber > we[biggestWatchedIdx].EpisodeNumber ||
				v.SeasonNumber > we[biggestWatchedIdx].SeasonNumber {
				biggestWatchedIdx = i
			}
		}
	}

	if biggestWatchingIdx > -1 {
		return we[biggestWatchingIdx]
	} else if biggestWatchedIdx > -1 {
		return we[biggestWatchedIdx]
	}

	return entity.WatchedEpisode{}
}

func SeasonAndEpToReadable(
	seasonNum int,
	episodeNum int,
) string {
	return "S" + strconv.Itoa(seasonNum) + "E" + strconv.Itoa(episodeNum)
}
