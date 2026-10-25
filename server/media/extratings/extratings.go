package extratings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	gocache "github.com/robfig/go-cache"
	"github.com/sbondCo/Watcharr/domain"
)

var (
	RatingsCache = gocache.New(24*time.Hour, 10*time.Minute)
	httpClient   = &http.Client{
		Timeout: 4 * time.Second,
	}
)

type FetchOptions struct {
	MediaType  string // "movie" or "tv"
	TMDBID     int
	ImdbID     string
	TvdbID     int
	Title      string
	OrigTitle  string
	Year       int
	IsAnime    bool
	OMDBKey    string
	MDBListKey string
}

type omdbResponse struct {
	Response   string `json:"Response"`
	Rated      string `json:"Rated"`
	ImdbRating string `json:"imdbRating"`
	ImdbVotes  string `json:"imdbVotes"`
	Metascore  string `json:"Metascore"`
	Awards     string `json:"Awards"`
	BoxOffice  string `json:"BoxOffice"`
	Ratings    []struct {
		Source string `json:"Source"`
		Value  string `json:"Value"`
	} `json:"Ratings"`
}

type mdblistResponse struct {
	Ratings []struct {
		Source  string      `json:"source"`
		Value   interface{} `json:"value"`
		Score   interface{} `json:"score"`
		Popular interface{} `json:"popular"`
	} `json:"ratings"`
}

type tvmazeShowResponse struct {
	Rating struct {
		Average float64 `json:"average"`
	} `json:"rating"`
	Schedule struct {
		Time string   `json:"time"`
		Days []string `json:"days"`
	} `json:"schedule"`
	Network *struct {
		Name string `json:"name"`
	} `json:"network"`
	WebChannel *struct {
		Name string `json:"name"`
	} `json:"webChannel"`
}

type anilistGraphQLResponse struct {
	Data struct {
		Media struct {
			IDMal        int    `json:"idMal"`
			AverageScore int    `json:"averageScore"`
			MeanScore    int    `json:"meanScore"`
			Source       string `json:"source"`
			Studios      struct {
				Nodes []struct {
					Name string `json:"name"`
				} `json:"nodes"`
			} `json:"studios"`
		} `json:"Media"`
	} `json:"data"`
}

// FetchEnrichments fetches multi-platform ratings & enrichments from OMDb, MDBList, TVMaze, and AniList with 24h caching.
func FetchEnrichments(opts FetchOptions) (*domain.ExternalRatings, string) {
	cacheKey := fmt.Sprintf("ext_%s_%d_%s_%t_%t", opts.MediaType, opts.TMDBID, opts.ImdbID, opts.OMDBKey != "", opts.MDBListKey != "")
	if cached, found := RatingsCache.Get(cacheKey); found {
		if entry, ok := cached.(cachedResult); ok {
			return entry.ratings, entry.certification
		}
	}

	var (
		mu            sync.Mutex
		wg            sync.WaitGroup
		result        domain.ExternalRatings
		hasData       bool
		certification string
	)

	// 1. OMDb API (IMDb, Rotten Tomatoes, Metacritic, Awards, BoxOffice, Rated)
	if opts.OMDBKey != "" && opts.ImdbID != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			u := fmt.Sprintf("https://www.omdbapi.com/?i=%s&apikey=%s", url.QueryEscape(opts.ImdbID), url.QueryEscape(opts.OMDBKey))
			resp, err := httpClient.Get(u)
			if err != nil {
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return
			}
			var data omdbResponse
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || data.Response != "True" {
				return
			}
			mu.Lock()
			defer mu.Unlock()
			if data.ImdbRating != "" && data.ImdbRating != "N/A" {
				result.IMDbRating = data.ImdbRating
				hasData = true
			}
			if data.ImdbVotes != "" && data.ImdbVotes != "N/A" {
				result.IMDbVotes = data.ImdbVotes
			}
			if data.Metascore != "" && data.Metascore != "N/A" {
				result.Metacritic = data.Metascore
				hasData = true
			}
			for _, r := range data.Ratings {
				src := strings.ToLower(r.Source)
				if strings.Contains(src, "rotten tomatoes") && r.Value != "" && r.Value != "N/A" {
					result.RottenTomatoes = r.Value
					hasData = true
				} else if strings.Contains(src, "metacritic") && result.Metacritic == "" && r.Value != "" && r.Value != "N/A" {
					result.Metacritic = strings.TrimSuffix(r.Value, "/100")
					hasData = true
				}
			}
			if data.Awards != "" && data.Awards != "N/A" {
				result.Awards = data.Awards
				hasData = true
			}
			if data.BoxOffice != "" && data.BoxOffice != "N/A" {
				result.BoxOffice = data.BoxOffice
				hasData = true
			}
			if data.Rated != "" && data.Rated != "N/A" && data.Rated != "Not Rated" && data.Rated != "Unrated" {
				certification = data.Rated
			}
		}()
	}

	// 2. MDBList API (IMDb, Rotten Tomatoes Critics & Audience, Metacritic, Letterboxd, Trakt, MAL)
	if opts.MDBListKey != "" && (opts.ImdbID != "" || opts.TMDBID > 0) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var u string
			if opts.ImdbID != "" {
				u = fmt.Sprintf("https://mdblist.com/api/?apikey=%s&i=%s", url.QueryEscape(opts.MDBListKey), url.QueryEscape(opts.ImdbID))
			} else {
				mType := "movie"
				if opts.MediaType == "tv" {
					mType = "show"
				}
				u = fmt.Sprintf("https://mdblist.com/api/?apikey=%s&tm=%d&m=%s", url.QueryEscape(opts.MDBListKey), opts.TMDBID, mType)
			}
			resp, err := httpClient.Get(u)
			if err != nil {
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return
			}
			var data mdblistResponse
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				return
			}
			mu.Lock()
			defer mu.Unlock()
			for _, r := range data.Ratings {
				valStr := formatRatingValue(r.Value)
				if valStr == "" || valStr == "0" {
					valStr = formatRatingValue(r.Score)
				}
				if valStr == "" || valStr == "0" {
					continue
				}
				switch strings.ToLower(r.Source) {
				case "imdb":
					if result.IMDbRating == "" {
						result.IMDbRating = valStr
						hasData = true
					}
				case "tomatoes":
					if result.RottenTomatoes == "" {
						if !strings.HasSuffix(valStr, "%") {
							valStr += "%"
						}
						result.RottenTomatoes = valStr
						hasData = true
					}
				case "tomatoesaudience", "popcorn":
					if !strings.HasSuffix(valStr, "%") {
						valStr += "%"
					}
					result.RottenTomatoesAudience = valStr
					hasData = true
				case "metacritic":
					if result.Metacritic == "" {
						result.Metacritic = valStr
						hasData = true
					}
				case "letterboxd":
					result.Letterboxd = valStr
					hasData = true
				case "trakt":
					if !strings.HasSuffix(valStr, "%") {
						valStr += "%"
					}
					result.Trakt = valStr
					hasData = true
				}
			}
		}()
	}

	// 3. TVMaze Public API (100% Free, Zero API Key — for TV Shows)
	if opts.MediaType == "tv" && (opts.ImdbID != "" || opts.TvdbID > 0) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var u string
			if opts.ImdbID != "" {
				u = "https://api.tvmaze.com/lookup/shows?imdb=" + url.QueryEscape(opts.ImdbID)
			} else {
				u = "https://api.tvmaze.com/lookup/shows?thetvdb=" + strconv.Itoa(opts.TvdbID)
			}
			resp, err := httpClient.Get(u)
			if err != nil {
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return
			}
			var data tvmazeShowResponse
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				return
			}
			mu.Lock()
			defer mu.Unlock()
			if data.Rating.Average > 0 {
				result.TVMaze = fmt.Sprintf("%.1f", data.Rating.Average)
				hasData = true
			}
			if len(data.Schedule.Days) > 0 {
				sched := strings.Join(data.Schedule.Days, ", ")
				if data.Schedule.Time != "" {
					sched += " • " + data.Schedule.Time
				}
				result.NetworkSchedule = sched
				hasData = true
			}
		}()
	}

	// 4. AniList Public GraphQL API (100% Free, Zero API Key — for Anime movies & shows)
	if opts.IsAnime && (opts.Title != "" || opts.OrigTitle != "") {
		wg.Add(1)
		go func() {
			defer wg.Done()
			searchTitle := opts.OrigTitle
			if searchTitle == "" {
				searchTitle = opts.Title
			}
			query := `query ($search: String) {
				Media(search: $search, type: ANIME) {
					idMal
					averageScore
					meanScore
					source
					studios(isMain: true) {
						nodes {
							name
						}
					}
				}
			}`
			payload, _ := json.Marshal(map[string]interface{}{
				"query": query,
				"variables": map[string]interface{}{
					"search": searchTitle,
				},
			})
			req, err := http.NewRequest("POST", "https://graphql.anilist.co", bytes.NewReader(payload))
			if err != nil {
				return
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
			resp, err := httpClient.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return
			}
			var data anilistGraphQLResponse
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				return
			}
			mu.Lock()
			defer mu.Unlock()
			score := data.Data.Media.AverageScore
			if score == 0 {
				score = data.Data.Media.MeanScore
			}
			if score > 0 {
				result.AniList = fmt.Sprintf("%d%%", score)
				hasData = true
			}
			if data.Data.Media.IDMal > 0 {
				result.MyAnimeListID = data.Data.Media.IDMal
				hasData = true
			}
			if len(data.Data.Media.Studios.Nodes) > 0 && data.Data.Media.Studios.Nodes[0].Name != "" {
				result.AnimationStudio = data.Data.Media.Studios.Nodes[0].Name
				hasData = true
			}
			if data.Data.Media.Source != "" {
				result.SourceMaterial = strings.ReplaceAll(strings.Title(strings.ToLower(data.Data.Media.Source)), "_", " ")
				hasData = true
			}
		}()
	}

	wg.Wait()

	var ptr *domain.ExternalRatings
	if hasData {
		ptr = &result
	}
	RatingsCache.Set(cacheKey, cachedResult{ratings: ptr, certification: certification}, 24*time.Hour)
	return ptr, certification
}

type cachedResult struct {
	ratings       *domain.ExternalRatings
	certification string
}

func formatRatingValue(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case float64:
		if val <= 0 {
			return ""
		}
		if val == float64(int(val)) {
			return strconv.Itoa(int(val))
		}
		return fmt.Sprintf("%.1f", val)
	case string:
		return strings.TrimSpace(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

