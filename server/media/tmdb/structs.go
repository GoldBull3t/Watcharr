package tmdb

import (
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/util"
)

// Separated from `TMDBSearchResponse` so we can embed it for
// easily assigning all page fields in one.
type PageFields struct {
	Page         int `json:"page"`
	TotalPages   int `json:"total_pages"`
	TotalResults int `json:"total_results"`
}

type SearchResponse[R any] struct {
	PageFields
	Results []R `json:"results"`
}

// A common "base" type for search results.
// Some properties are used commonly for all types except Person, but
// are still embedded in person for ease of use right now.
type SearchResult struct {
	// TMDB ID
	ID int `json:"id"`
	// Media Type (movie, tv, person)
	// **Some requests won't return this value
	// (namely any request other than a multi
	// type search), but we add it in manually.**
	MediaType string `json:"media_type"`
	// Summary (only for movie/tv)
	Overview string `json:"overview"`
	// Poster path (only for movie/tv)
	PosterPath string `json:"poster_path"`
	// Rating (only for movie/tv)
	VoteAverage float32 `json:"vote_average"`
	// Amount of votes for rating (only for movie/tv)
	VoteCount uint32 `json:"vote_count"`
}

// Adds the base items to a Media struct, which can be used in the
// structs that embed TMDBSearchResult to simplify and reduce duplication.
func (t *SearchResult) AsMedia() domain.Media {
	m := domain.Media{
		IDs: domain.MediaIDs{
			TMDB: t.ID,
		},
		Summary:       t.Overview,
		ExtPosterPath: t.PosterPath,
		Rating:        uint(t.VoteAverage * 10),
		RatingCount:   uint(t.VoteCount),
	}
	switch t.MediaType {
	case "movie":
		m.Type = domain.MediaTypeTMDBMovie
	case "tv":
		m.Type = domain.MediaTypeTMDBShow
		if c := domain.GetCliffhangerInfo(t.ID); c != nil {
			m.Cancellation = c
		}
	case "person":
		m.Type = domain.MediaTypeTMDBPerson
	}
	return m
}

//
// Multi Search
//

type SearchMultiResponse struct {
	SearchResponse[SearchMultiResult]
}

type SearchMultiResult struct {
	SearchResult
	Adult            bool     `json:"adult"`
	BackdropPath     string   `json:"backdrop_path"`
	Title            string   `json:"title,omitempty"`
	OriginalLanguage string   `json:"original_language"`
	OriginalTitle    string   `json:"original_title,omitempty"`
	ProfilePath      string   `json:"profile_path"`
	GenreIds         []int64  `json:"genre_ids"`
	Popularity       float32  `json:"popularity"`
	ReleaseDate      string   `json:"release_date,omitempty"`
	Video            bool     `json:"video,omitempty"`
	Name             string   `json:"name,omitempty"`
	OriginalName     string   `json:"original_name,omitempty"`
	FirstAirDate     string   `json:"first_air_date,omitempty"`
	OriginCountry    []string `json:"origin_country,omitempty"`
	// Below are for tv episode results
	AirDate        string `json:"air_date,omitempty"`
	EpisodeNumber  int    `json:"episode_number,omitempty"`
	EpisodeType    string `json:"episode_type,omitempty"`
	ProductionCode string `json:"production_code,omitempty"`
	Runtime        int    `json:"runtime,omitempty"`
	SeasonNumber   int    `json:"season_number,omitempty"`
	ShowId         int    `json:"show_id,omitempty"`
	StillPath      string `json:"still_path,omitempty"`
}

func (t *SearchMultiResult) AsMedia() domain.Media {
	m := t.SearchResult.AsMedia()
	m.ExtBackdropPath = t.BackdropPath
	m.OriginalLanguage = t.OriginalLanguage

	m.Name = t.Title
	m.OriginalTitle = t.OriginalTitle
	if t.Name != "" {
		m.Name = t.Name
	}
	if t.OriginalName != "" {
		m.OriginalTitle = t.OriginalName
	}

	var tmdbReleaseDate string
	switch t.MediaType {
	case "movie":
		tmdbReleaseDate = t.ReleaseDate
	case "tv":
		tmdbReleaseDate = t.FirstAirDate
	case "person":
		m.ExtPosterPath = t.ProfilePath
	}
	if releaseDate, err := time.Parse("2006-01-02", tmdbReleaseDate); err == nil {
		m.ReleaseDate = releaseDate
	} else {
		slog.Error("AsMedia: Failed to parse release date", "name", m.Name, "error", err)
	}
	return m
}

//
// Movie Search
//

type SearchMoviesResponse struct {
	SearchResponse[SearchMovieResult]
}

type SearchMovieResult struct {
	SearchResult
	Adult            bool    `json:"adult"`
	BackdropPath     string  `json:"backdrop_path"`
	GenreIds         []int   `json:"genre_ids"`
	OriginalLanguage string  `json:"original_language"`
	OriginalTitle    string  `json:"original_title"`
	Popularity       float64 `json:"popularity"`
	ReleaseDate      string  `json:"release_date"`
	Title            string  `json:"title"`
	Video            bool    `json:"video"`
}

func (t *SearchMovieResult) AsMedia() domain.Media {
	m := t.SearchResult.AsMedia()
	m.Name = t.Title
	m.OriginalTitle = t.OriginalTitle
	m.OriginalLanguage = t.OriginalLanguage
	m.ExtBackdropPath = t.BackdropPath
	if releaseDate, err := time.Parse("2006-01-02", t.ReleaseDate); err == nil {
		m.ReleaseDate = releaseDate
	} else {
		slog.Error("AsMedia: Failed to parse release date", "name", m.Name, "error", err)
	}
	return m
}

//
// Tv Shows Search
//

type SearchShowsResponse struct {
	SearchResponse[SearchShowsResult]
}

type SearchShowsResult struct {
	SearchResult
	Adult            bool     `json:"adult"`
	BackdropPath     string   `json:"backdrop_path"`
	GenreIds         []int    `json:"genre_ids"`
	OriginCountry    []string `json:"origin_country"`
	OriginalLanguage string   `json:"original_language"`
	OriginalName     string   `json:"original_name"`
	Popularity       float64  `json:"popularity"`
	FirstAirDate     string   `json:"first_air_date"`
	Name             string   `json:"name"`
}

func (t *SearchShowsResult) AsMedia() domain.Media {
	m := t.SearchResult.AsMedia()
	m.Name = t.Name
	m.OriginalTitle = t.OriginalName
	m.OriginalLanguage = t.OriginalLanguage
	m.ExtBackdropPath = t.BackdropPath
	if releaseDate, err := time.Parse("2006-01-02", t.FirstAirDate); err == nil {
		m.ReleaseDate = releaseDate
	} else {
		slog.Error("AsMedia: Failed to parse release date", "name", m.Name, "error", err)
	}
	return m
}

//
// People Search
//

type SearchPeopleResult struct {
	SearchResult
	Adult              bool    `json:"adult"`
	Gender             int     `json:"gender"`
	KnownForDepartment string  `json:"known_for_department"`
	Name               string  `json:"name"`
	OriginalName       string  `json:"original_name"`
	Popularity         float64 `json:"popularity"`
	ProfilePath        string  `json:"profile_path"`
	KnownFor           []struct {
		Adult            bool    `json:"adult"`
		BackdropPath     string  `json:"backdrop_path"`
		ID               int     `json:"id"`
		Title            string  `json:"title"`
		OriginalLanguage string  `json:"original_language"`
		OriginalTitle    string  `json:"original_title"`
		Overview         string  `json:"overview"`
		PosterPath       string  `json:"poster_path"`
		MediaType        string  `json:"media_type"`
		GenreIds         []int   `json:"genre_ids"`
		Popularity       float64 `json:"popularity"`
		ReleaseDate      string  `json:"release_date"`
		Video            bool    `json:"video"`
		VoteAverage      float64 `json:"vote_average"`
		VoteCount        int     `json:"vote_count"`
	} `json:"known_for"`
}

func (t *SearchPeopleResult) AsMedia() domain.Media {
	m := t.SearchResult.AsMedia()
	m.Name = t.Name
	m.ExtPosterPath = t.ProfilePath
	return m
}

type SearchPeopleResponse struct {
	SearchResponse[SearchPeopleResult]
}

//
// Search By External ID
//

type FindByExternalIdResponse struct {
	// These are all a SearchMultiResult so our search func can easily
	// combine all of them into one []SearchMultiResult for response
	// to client (seems not easy to convert to SearchMultiResult for
	// concatenation after unmarshalling to correct type).
	MovieResults     []SearchMultiResult `json:"movie_results"`
	PersonResults    []SearchMultiResult `json:"person_results"`
	TvResults        []SearchMultiResult `json:"tv_results"`
	TvSeasonResults  []SearchMultiResult `json:"tv_season_results"`
	TvEpisodeResults []SearchMultiResult `json:"tv_episode_results"`
}

//
// Content Details
// A base for details structs.
//

type ContentDetails struct {
	ID           int    `json:"id"`
	PosterPath   string `json:"poster_path"`
	BackdropPath string `json:"backdrop_path"`
	Genres       []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
	Homepage            string  `json:"homepage"`
	Popularity          float32 `json:"popularity"`
	Overview            string  `json:"overview"`
	OriginalLanguage    string  `json:"original_language"`
	ProductionCompanies []struct {
		ID            int    `json:"id"`
		LogoPath      string `json:"logo_path"`
		Name          string `json:"name"`
		OriginCountry string `json:"origin_country"`
	} `json:"production_companies"`
	ProductionCountries []struct {
		Iso31661 string `json:"iso_3166_1"`
		Name     string `json:"name"`
	} `json:"production_countries"`
	Tagline         string  `json:"tagline"`
	VoteAverage     float32 `json:"vote_average"`
	VoteCount       uint32  `json:"vote_count"`
	SpokenLanguages []struct {
		EnglishName string `json:"english_name"`
		Iso6391     string `json:"iso_639_1"`
		Name        string `json:"name"`
	} `json:"spoken_languages"`

	// Extra items because we use `append_to_response` on the request
	Videos ContentVideos `json:"videos"`
	// Raw watched providers object from tmdb
	WatchProviders interface{} `json:"watch/providers"`
	// Watched providers but after we apply our transformations to it.
	WatchProvidersTransformed WatchProviders `json:"-"`
}

// Adds the base items to a Media struct, which can be used in the
// structs that embed SearchResult to simplify and reduce duplication.
func (t *ContentDetails) AsMedia() domain.Media {
	m := domain.Media{
		IDs: domain.MediaIDs{
			TMDB: t.ID,
		},
		Summary:          t.Overview,
		Tagline:          t.Tagline,
		OriginalLanguage: t.OriginalLanguage,
		ExtPosterPath:    t.PosterPath,
		ExtBackdropPath:  t.BackdropPath,
		Rating:           uint(t.VoteAverage * 10),
		RatingCount:      uint(t.VoteCount),
		Homepage:         t.Homepage,
	}
	// Genres
	for _, g := range t.Genres {
		m.Genres = append(m.Genres, domain.MediaGenre{
			ID:   uint(g.ID),
			Name: g.Name,
		})
	}
	// Production companies
	for _, pc := range t.ProductionCompanies {
		if pc.Name == "" {
			continue
		}
		m.ProductionCompanies = append(m.ProductionCompanies, domain.MediaNetwork{
			ID:            pc.ID,
			Name:          pc.Name,
			LogoPath:      pc.LogoPath,
			OriginCountry: pc.OriginCountry,
		})
	}
	// Videos (Trailers, Teasers, Clips, Featurettes, Behind the Scenes, Opening Credits, Bloopers)
	hasBest := false
	for i := range t.Videos.Results {
		v := &t.Videos.Results[i]
		if v.Key == "" || (v.Site != "" && !strings.EqualFold(v.Site, "YouTube")) {
			continue
		}
		vTypeLower := strings.ToLower(strings.TrimSpace(v.Type))
		var mediaVType domain.MediaVideoType
		switch vTypeLower {
		case "trailer":
			mediaVType = domain.MediaVideoTypeTrailer
		case "teaser":
			mediaVType = domain.MediaVideoTypeTeaser
		case "clip":
			mediaVType = domain.MediaVideoTypeClip
		case "featurette":
			mediaVType = domain.MediaVideoTypeFeaturette
		case "behind the scenes":
			mediaVType = domain.MediaVideoTypeBehindTheScenes
		case "bloopers":
			mediaVType = domain.MediaVideoTypeBloopers
		case "opening credits":
			mediaVType = domain.MediaVideoTypeOpeningCredits
		default:
			mediaVType = domain.MediaVideoTypeOther
		}
		isBest := false
		if mediaVType == domain.MediaVideoTypeTrailer && v.Official && strings.Contains(strings.ToLower(v.Name), "official trailer") && !hasBest {
			isBest = true
			hasBest = true
		}
		m.Videos = append(m.Videos, domain.MediaVideo{
			ID:   v.Key,
			Name: v.Name,
			Type: mediaVType,
			Best: isBest,
		})
	}
	if !hasBest {
		for i := range m.Videos {
			if m.Videos[i].Type == domain.MediaVideoTypeTrailer {
				m.Videos[i].Best = true
				break
			}
		}
	}
	// Watch providers (Flatrate/Sub, Free, Ads, Rent, Buy)
	seenProviders := make(map[string]bool)
	addProviders := func(list []WatchProvider, pType domain.MediaProviderType) {
		for _, v := range list {
			key := strings.ToLower(v.ProviderName) + "|" + string(pType)
			if seenProviders[key] {
				continue
			}
			seenProviders[key] = true
			m.Providers = append(m.Providers, domain.MediaProvider{
				Name:     v.ProviderName,
				Type:     pType,
				LogoPath: v.LogoPath,
			})
		}
	}
	addProviders(t.WatchProvidersTransformed.Flatrate, domain.MediaProviderTypeSub)
	addProviders(t.WatchProvidersTransformed.Free, domain.MediaProviderTypeFree)
	addProviders(t.WatchProvidersTransformed.Ads, domain.MediaProviderTypeAds)
	addProviders(t.WatchProvidersTransformed.Rent, domain.MediaProviderTypeRent)
	addProviders(t.WatchProvidersTransformed.Buy, domain.MediaProviderTypeBuy)
	m.ProvidersFullListLink = t.WatchProvidersTransformed.Link
	return m
}

//
//
// Movie Details
//

type TMDBBelongsToCollection struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	PosterPath   string `json:"poster_path"`
	BackdropPath string `json:"backdrop_path"`
}

type MovieReleaseDates struct {
	Results []struct {
		Iso31661     string `json:"iso_3166_1"`
		ReleaseDates []struct {
			Certification string `json:"certification"`
			Type          int    `json:"type"`
			ReleaseDate   string `json:"release_date"`
		} `json:"release_dates"`
	} `json:"results"`
}

func (r *MovieReleaseDates) ExtractCertification(preferredCountry string) string {
	candidates := []string{strings.ToUpper(preferredCountry), "PT", "US", "GB", "BR"}
	for _, country := range candidates {
		if country == "" {
			continue
		}
		for _, res := range r.Results {
			if strings.EqualFold(res.Iso31661, country) {
				for _, rd := range res.ReleaseDates {
					if cert := strings.TrimSpace(rd.Certification); cert != "" {
						return cert
					}
				}
			}
		}
	}
	return ""
}

type MovieDetails struct {
	ContentDetails
	Adult               bool                      `json:"adult"`
	BelongsToCollection *TMDBBelongsToCollection `json:"belongs_to_collection"`
	Budget              uint32                    `json:"budget"`
	ImdbID              string                    `json:"imdb_id"`
	OriginalTitle       string                    `json:"original_title"`
	ReleaseDate         string                    `json:"release_date"`
	Revenue             uint32                    `json:"revenue"`
	Runtime             uint32                    `json:"runtime"`
	Title               string                    `json:"title"`
	Video               bool                      `json:"video"`

	Status MovieStatus `json:"status"`

	// Extra items because we use `append_to_response` on the request
	ExternalIds     ExternalIdsMovie  `json:"external_ids"`
	Similar         MovieSimilar      `json:"similar"`
	Recommendations MovieSimilar      `json:"recommendations"`
	ReleaseDates    MovieReleaseDates `json:"release_dates"`
}

func (t *MovieDetails) AsMedia() domain.Media {
	m := t.ContentDetails.AsMedia()
	m.Type = domain.MediaTypeTMDBMovie
	m.Name = t.Title
	m.OriginalTitle = t.OriginalTitle
	m.Runtime = uint(t.Runtime)
	m.Budget = t.Budget
	m.Revenue = t.Revenue
	m.Status = string(t.Status)
	m.Certification = t.ReleaseDates.ExtractCertification("")
	if releaseDate, err := time.Parse("2006-01-02", t.ReleaseDate); err == nil {
		m.ReleaseDate = releaseDate
	} else {
		slog.Error("AsMedia: Failed to parse release date", "name", m.Name, "error", err)
	}
	// IDS
	m.IDs.IMDB = t.ExternalIds.ImdbID
	if m.IDs.IMDB == "" {
		m.IDs.IMDB = t.ImdbID
	}
	m.IDs.Wikidata = t.ExternalIds.WikidataID
	if t.BelongsToCollection != nil && t.BelongsToCollection.ID > 0 {
		m.Collection = &domain.MediaCollection{
			ID:           t.BelongsToCollection.ID,
			Name:         t.BelongsToCollection.Name,
			PosterPath:   t.BelongsToCollection.PosterPath,
			BackdropPath: t.BelongsToCollection.BackdropPath,
		}
	}
	// Combine Recommendations (higher quality) + Similar, deduplicated
	seenSimilar := make(map[int]bool)
	for i := range t.Recommendations.Results {
		recMedia := t.Recommendations.Results[i].AsMedia()
		m.Recommendations = append(m.Recommendations, recMedia)
		if !seenSimilar[recMedia.IDs.TMDB] {
			seenSimilar[recMedia.IDs.TMDB] = true
			m.Similar = append(m.Similar, recMedia)
		}
	}
	for i := range t.Similar.Results {
		simMedia := t.Similar.Results[i].AsMedia()
		if !seenSimilar[simMedia.IDs.TMDB] {
			seenSimilar[simMedia.IDs.TMDB] = true
			m.Similar = append(m.Similar, simMedia)
		}
	}
	return m
}

//
// Collection Details
//

type TMDBCollectionPart struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	OriginalTitle string  `json:"original_title"`
	Overview      string  `json:"overview"`
	ReleaseDate   string  `json:"release_date"`
	PosterPath    string  `json:"poster_path"`
	BackdropPath  string  `json:"backdrop_path"`
	VoteAverage   float64 `json:"vote_average"`
	VoteCount     int     `json:"vote_count"`
	Popularity    float64 `json:"popularity"`
	Adult         bool    `json:"adult"`
}

func (p TMDBCollectionPart) AsMedia() domain.Media {
	m := domain.Media{
		Type:            domain.MediaTypeTMDBMovie,
		IDs:             domain.MediaIDs{TMDB: p.ID},
		Name:            p.Title,
		Summary:         p.Overview,
		ExtPosterPath:   p.PosterPath,
		ExtBackdropPath: p.BackdropPath,
		Rating:          uint(math.Round(p.VoteAverage * 10)),
		RatingCount:     uint(p.VoteCount),
	}
	if releaseDate, err := time.Parse("2006-01-02", p.ReleaseDate); err == nil {
		m.ReleaseDate = releaseDate
	}
	return m
}

type TMDBCollection struct {
	ID           int                  `json:"id"`
	Name         string               `json:"name"`
	Overview     string               `json:"overview"`
	PosterPath   string               `json:"poster_path"`
	BackdropPath string               `json:"backdrop_path"`
	Parts        []TMDBCollectionPart `json:"parts"`
}

func (c *TMDBCollection) AsMediaCollection() domain.MediaCollection {
	mc := domain.MediaCollection{
		ID:           c.ID,
		Name:         c.Name,
		Overview:     c.Overview,
		PosterPath:   c.PosterPath,
		BackdropPath: c.BackdropPath,
		Parts:        make([]domain.Media, 0, len(c.Parts)),
	}
	for _, p := range c.Parts {
		mc.Parts = append(mc.Parts, p.AsMedia())
	}
	return mc
}

//
// Movie Details Similar
//

type MovieSimilar struct {
	SearchResponse[MovieSimilarResult]
}

type MovieSimilarResult struct {
	ID               int     `json:"id"`
	Title            string  `json:"title"`
	Adult            bool    `json:"adult"`
	BackdropPath     string  `json:"backdrop_path"`
	GenreIds         []int   `json:"genre_ids"`
	OriginalLanguage string  `json:"original_language"`
	OriginalTitle    string  `json:"original_title"`
	Overview         string  `json:"overview"`
	Popularity       float64 `json:"popularity"`
	PosterPath       string  `json:"poster_path"`
	ReleaseDate      string  `json:"release_date"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        uint32  `json:"vote_count"`
}

func (t *MovieSimilarResult) AsMedia() domain.Media {
	m := domain.Media{
		IDs: domain.MediaIDs{
			TMDB: t.ID,
		},
		Type:             domain.MediaTypeTMDBMovie,
		Name:             t.Title,
		OriginalTitle:    t.OriginalTitle,
		OriginalLanguage: t.OriginalLanguage,
		Summary:          t.Overview,
		ExtPosterPath:    t.PosterPath,
		ExtBackdropPath:  t.BackdropPath,
		Rating:           uint(t.VoteAverage * 10),
		RatingCount:      uint(t.VoteCount),
	}
	if releaseDate, err := time.Parse("2006-01-02", t.ReleaseDate); err == nil {
		m.ReleaseDate = releaseDate
	} else {
		slog.Error("AsMedia: Failed to parse release date", "name", m.Name, "error", err)
	}
	return m
}

//
// Show Details
//

type ShowContentRatings struct {
	Results []struct {
		Iso31661 string `json:"iso_3166_1"`
		Rating   string `json:"rating"`
	} `json:"results"`
}

func (r *ShowContentRatings) ExtractCertification(preferredCountry string) string {
	candidates := []string{strings.ToUpper(preferredCountry), "PT", "US", "GB", "BR"}
	for _, country := range candidates {
		if country == "" {
			continue
		}
		for _, res := range r.Results {
			if strings.EqualFold(res.Iso31661, country) {
				if cert := strings.TrimSpace(res.Rating); cert != "" {
					return cert
				}
			}
		}
	}
	return ""
}

type TMDBEpisodeBasic struct {
	AirDate        string  `json:"air_date"`
	EpisodeNumber  int     `json:"episode_number"`
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	Overview       string  `json:"overview"`
	ProductionCode string  `json:"production_code"`
	Runtime        int     `json:"runtime"`
	SeasonNumber   int     `json:"season_number"`
	StillPath      string  `json:"still_path"`
	VoteAverage    float64 `json:"vote_average"`
	VoteCount      uint32  `json:"vote_count"`
}

type ShowDetails struct {
	ContentDetails
	CreatedBy []struct {
		ID          int    `json:"id"`
		CreditID    string `json:"credit_id"`
		Name        string `json:"name"`
		Gender      int    `json:"gender"`
		ProfilePath string `json:"profile_path"`
	} `json:"created_by"`
	EpisodeRunTime   []int    `json:"episode_run_time"`
	FirstAirDate     string   `json:"first_air_date"`
	InProduction     bool     `json:"in_production"`
	Languages        []string `json:"languages"`
	LastAirDate      string   `json:"last_air_date"`
	LastEpisodeToAir TMDBEpisodeBasic  `json:"last_episode_to_air"`
	Name             string            `json:"name"`
	NextEpisodeToAir *TMDBEpisodeBasic `json:"next_episode_to_air"`
	Networks         []struct {
		Name          string `json:"name"`
		ID            int    `json:"id"`
		LogoPath      string `json:"logo_path"`
		OriginCountry string `json:"origin_country"`
	} `json:"networks"`
	NumberOfEpisodes uint32   `json:"number_of_episodes"`
	NumberOfSeasons  uint32   `json:"number_of_seasons"`
	OriginCountry    []string `json:"origin_country"`
	OriginalName     string   `json:"original_name"`
	Seasons          []struct {
		AirDate      string `json:"air_date"`
		EpisodeCount int    `json:"episode_count"`
		ID           int    `json:"id"`
		Name         string `json:"name"`
		Overview     string `json:"overview"`
		PosterPath   string `json:"poster_path"`
		SeasonNumber int    `json:"season_number"`
	} `json:"seasons"`
	Type string `json:"type"`

	Status ShowStatus `json:"status"`

	// Extra items because we use `append_to_response` on the request
	ExternalIds     ExternalIdsShow    `json:"external_ids"`
	Keywords        Keywords           `json:"keywords"`
	Similar         ShowSimilar        `json:"similar"`
	Recommendations ShowSimilar        `json:"recommendations"`
	ContentRatings  ShowContentRatings `json:"content_ratings"`
}

func (t *ShowDetails) EffectiveRuntime() uint32 {
	if len(t.EpisodeRunTime) > 0 && t.EpisodeRunTime[0] > 0 {
		return uint32(t.EpisodeRunTime[0])
	}
	if t.LastEpisodeToAir.Runtime > 0 {
		return uint32(t.LastEpisodeToAir.Runtime)
	}
	if t.NextEpisodeToAir != nil && t.NextEpisodeToAir.Runtime > 0 {
		return uint32(t.NextEpisodeToAir.Runtime)
	}
	return 0
}

func (t *ShowDetails) AsMedia() domain.Media {
	m := t.ContentDetails.AsMedia()
	m.Type = domain.MediaTypeTMDBShow
	m.Name = t.Name
	m.OriginalTitle = t.OriginalName
	m.Status = string(t.Status)
	m.Runtime = uint(t.EffectiveRuntime())
	m.NumberOfEpisodes = t.NumberOfEpisodes
	m.NumberOfSeasons = t.NumberOfSeasons
	m.Certification = t.ContentRatings.ExtractCertification("")
	if c := domain.GetCliffhangerInfo(t.ID); c != nil {
		m.Cancellation = c
	}
	if releaseDate, err := time.Parse("2006-01-02", t.FirstAirDate); err == nil {
		m.ReleaseDate = releaseDate
	} else {
		slog.Error("AsMedia: Failed to parse release date", "name", m.Name, "error", err)
	}
	if releaseDateLast, err := time.Parse("2006-01-02", t.LastAirDate); err == nil {
		m.ReleaseDateLast = releaseDateLast
	} else {
		slog.Error("AsMedia: Failed to parse release date last",
			"name", m.Name,
			"error", err)
	}
	// Networks & Creators
	for _, n := range t.Networks {
		if n.Name == "" {
			continue
		}
		m.Networks = append(m.Networks, domain.MediaNetwork{
			ID:            n.ID,
			Name:          n.Name,
			LogoPath:      n.LogoPath,
			OriginCountry: n.OriginCountry,
		})
	}
	for _, cb := range t.CreatedBy {
		if cb.Name == "" {
			continue
		}
		m.CreatedBy = append(m.CreatedBy, domain.MediaCreator{
			ID:          cb.ID,
			Name:        cb.Name,
			ProfilePath: cb.ProfilePath,
		})
	}
	if t.LastEpisodeToAir.ID > 0 {
		m.LastEpisode = &domain.MediaEpisodeSummary{
			SeasonNumber:  t.LastEpisodeToAir.SeasonNumber,
			EpisodeNumber: t.LastEpisodeToAir.EpisodeNumber,
			Name:          t.LastEpisodeToAir.Name,
			AirDate:       t.LastEpisodeToAir.AirDate,
			Runtime:       t.LastEpisodeToAir.Runtime,
			Overview:      t.LastEpisodeToAir.Overview,
		}
	}
	if t.NextEpisodeToAir != nil && t.NextEpisodeToAir.ID > 0 {
		m.NextEpisode = &domain.MediaEpisodeSummary{
			SeasonNumber:  t.NextEpisodeToAir.SeasonNumber,
			EpisodeNumber: t.NextEpisodeToAir.EpisodeNumber,
			Name:          t.NextEpisodeToAir.Name,
			AirDate:       t.NextEpisodeToAir.AirDate,
			Runtime:       t.NextEpisodeToAir.Runtime,
			Overview:      t.NextEpisodeToAir.Overview,
		}
	}
	// IDS
	m.IDs.IMDB = t.ExternalIds.ImdbID
	m.IDs.Wikidata = t.ExternalIds.WikidataID
	m.IDs.TVDB = t.ExternalIds.TvdbID
	// Combine Recommendations (higher quality) + Similar, deduplicated
	seenSimilar := make(map[int]bool)
	for i := range t.Recommendations.Results {
		recMedia := t.Recommendations.Results[i].AsMedia()
		m.Recommendations = append(m.Recommendations, recMedia)
		if !seenSimilar[recMedia.IDs.TMDB] {
			seenSimilar[recMedia.IDs.TMDB] = true
			m.Similar = append(m.Similar, recMedia)
		}
	}
	for i := range t.Similar.Results {
		simMedia := t.Similar.Results[i].AsMedia()
		if !seenSimilar[simMedia.IDs.TMDB] {
			seenSimilar[simMedia.IDs.TMDB] = true
			m.Similar = append(m.Similar, simMedia)
		}
	}
	// Seasons
	for _, v := range t.Seasons {
		ms := domain.MediaSeason{
			Number:       v.SeasonNumber,
			Name:         v.Name,
			EpisodeCount: v.EpisodeCount,
		}
		if releaseDate, err := time.Parse("2006-01-02", v.AirDate); err == nil {
			ms.ReleaseDate = releaseDate
		} else {
			slog.Error("AsMedia: Failed to parse release date", "name", m.Name, "error", err)
		}
		m.Seasons = append(m.Seasons, ms)
	}
	// Is show anime
	for _, v := range t.Keywords.Results {
		// 210024 is the id of "anime" keyword on tmdb.
		if v.ID == 210024 {
			m.IsShowAnime = true
			break
		}
	}
	if !m.IsShowAnime && t.OriginalLanguage == "ja" {
		for _, g := range t.Genres {
			if g.ID == 16 {
				m.IsShowAnime = true
				break
			}
		}
	}
	return m
}

type SeasonDetails struct {
	ID       string `json:"_id"`
	AirDate  string `json:"air_date"`
	Episodes []struct {
		AirDate        string  `json:"air_date"`
		EpisodeNumber  int     `json:"episode_number"`
		ID             int     `json:"id"`
		Name           string  `json:"name"`
		Overview       string  `json:"overview"`
		ProductionCode string  `json:"production_code"`
		Runtime        int     `json:"runtime"`
		SeasonNumber   int     `json:"season_number"`
		ShowID         int     `json:"show_id"`
		StillPath      string  `json:"still_path"`
		VoteAverage    float64 `json:"vote_average"`
		VoteCount      int     `json:"vote_count"`
		Crew           []struct {
			Department         string  `json:"department"`
			Job                string  `json:"job"`
			CreditID           string  `json:"credit_id"`
			Adult              bool    `json:"adult"`
			Gender             int     `json:"gender"`
			ID                 int     `json:"id"`
			KnownForDepartment string  `json:"known_for_department"`
			Name               string  `json:"name"`
			OriginalName       string  `json:"original_name"`
			Popularity         float64 `json:"popularity"`
			ProfilePath        string  `json:"profile_path"`
		} `json:"crew"`
		GuestStars []struct {
			Character          string  `json:"character"`
			CreditID           string  `json:"credit_id"`
			Order              int     `json:"order"`
			Adult              bool    `json:"adult"`
			Gender             int     `json:"gender"`
			ID                 int     `json:"id"`
			KnownForDepartment string  `json:"known_for_department"`
			Name               string  `json:"name"`
			OriginalName       string  `json:"original_name"`
			Popularity         float64 `json:"popularity"`
			ProfilePath        string  `json:"profile_path"`
		} `json:"guest_stars"`
	} `json:"episodes"`
	Name         string `json:"name"`
	Overview     string `json:"overview"`
	ID0          int    `json:"id"`
	PosterPath   string `json:"poster_path"`
	SeasonNumber int    `json:"season_number"`
}

//
// Show Details Similar
//

type ShowSimilar struct {
	SearchResponse[ShowSimilarResult]
}

type ShowSimilarResult struct {
	ID               int      `json:"id"`
	Name             string   `json:"name"`
	Adult            bool     `json:"adult"`
	BackdropPath     string   `json:"backdrop_path"`
	GenreIds         []int    `json:"genre_ids"`
	OriginCountry    []string `json:"origin_country"`
	OriginalLanguage string   `json:"original_language"`
	OriginalName     string   `json:"original_name"`
	Overview         string   `json:"overview"`
	Popularity       float64  `json:"popularity"`
	PosterPath       string   `json:"poster_path"`
	FirstAirDate     string   `json:"first_air_date"`
	VoteAverage      float64  `json:"vote_average"`
	VoteCount        uint32   `json:"vote_count"`
}

func (t *ShowSimilarResult) AsMedia() domain.Media {
	m := domain.Media{
		IDs: domain.MediaIDs{
			TMDB: t.ID,
		},
		Type:             domain.MediaTypeTMDBShow,
		Name:             t.Name,
		OriginalTitle:    t.OriginalName,
		OriginalLanguage: t.OriginalLanguage,
		Summary:          t.Overview,
		ExtPosterPath:    t.PosterPath,
		ExtBackdropPath:  t.BackdropPath,
		Rating:           uint(t.VoteAverage * 10),
		RatingCount:      uint(t.VoteCount),
	}
	if releaseDate, err := time.Parse("2006-01-02", t.FirstAirDate); err == nil {
		m.ReleaseDate = releaseDate
	} else {
		slog.Error("AsMedia: Failed to parse release date", "name", m.Name, "error", err)
	}
	return m
}

//
// Person Details
//

type PersonDetails struct {
	ID                 int      `json:"id"`
	Name               string   `json:"name"`
	Birthday           string   `json:"birthday"`
	Deathday           string   `json:"deathday"`
	PlaceOfBirth       string   `json:"place_of_birth"`
	KnownForDepartment string   `json:"known_for_department"`
	Biography          string   `json:"biography"`
	AlsoKnownAs        []string `json:"also_known_as"`
	Popularity         float32  `json:"popularity"`
	ProfilePath        string   `json:"profile_path"`
	ImdbID             string   `json:"imdb_id"`
	Homepage           string   `json:"homepage"`
}

func (t *PersonDetails) AsPersonDetailsResponse() domain.PersonDetailsResponse {
	m := domain.PersonDetailsResponse{
		Name:               t.Name,
		PlaceOfBirth:       t.PlaceOfBirth,
		KnownForDepartment: t.KnownForDepartment,
		Biography:          t.Biography,
		ExtPosterPath:      t.ProfilePath,
		Homepage:           t.Homepage,
	}
	// Dates
	if date, err := time.Parse("2006-01-02", t.Birthday); err == nil {
		m.Birthday = date
	} else {
		slog.Error("AsPersonDetailsResponse: Failed to parse birthdate", "name", m.Name, "error", err)
	}
	if date, err := time.Parse("2006-01-02", t.Deathday); err == nil {
		m.Deathday = date
	} else {
		slog.Error("AsPersonDetailsResponse: Failed to parse birthdate", "name", m.Name, "error", err)
	}
	// Age
	m.Age = util.GetAge(m.Birthday, m.Deathday)
	return m
}

//
// Person Combined Credits
//

type PersonCombinedCredits struct {
	ID   int                               `json:"id"`
	Cast []PersonCombinedCreditsCastResult `json:"cast"`
	// crew PersonCombinedCreditsCrew
}

type PersonCombinedCreditsCastResult struct {
	ID               int      `json:"id"`
	OriginalLanguage string   `json:"original_language"`
	EpisodeCount     int      `json:"episode_count"`
	Overview         string   `json:"overview"`
	OriginCountry    []string `json:"origin_country"`
	OriginalName     string   `json:"original_name"`
	GenreIDs         []int    `json:"genre_ids"`
	Name             string   `json:"name"`
	MediaType        string   `json:"media_type"`
	PosterPath       string   `json:"poster_path"`
	FirstAirDate     string   `json:"first_air_date"`
	VoteAverage      float64  `json:"vote_average"`
	VoteCount        uint32   `json:"vote_count"`
	Character        string   `json:"character"`
	BackdropPath     string   `json:"backdrop_path"`
	Popularity       float64  `json:"popularity"`
	CreditID         string   `json:"credit_id"`
	OriginalTitle    string   `json:"original_title"`
	Video            bool     `json:"video"`
	ReleaseDate      string   `json:"release_date"`
	Title            string   `json:"title"`
	Adult            bool     `json:"adult"`
}

func (t *PersonCombinedCreditsCastResult) AsMedia() domain.Media {
	m := domain.Media{
		IDs: domain.MediaIDs{
			TMDB: t.ID,
		},
		Summary:         t.Overview,
		ExtPosterPath:   t.PosterPath,
		ExtBackdropPath: t.BackdropPath,
		Rating:          uint(t.VoteAverage * 10),
		RatingCount:     uint(t.VoteCount),
	}

	m.Name = t.Title
	if t.Name != "" {
		m.Name = t.Name
	}

	var tmdbReleaseDate string
	switch t.MediaType {
	case "movie":
		m.Type = domain.MediaTypeTMDBMovie
		tmdbReleaseDate = t.ReleaseDate
	case "tv":
		m.Type = domain.MediaTypeTMDBShow
		tmdbReleaseDate = t.FirstAirDate
	}
	if releaseDate, err := time.Parse("2006-01-02", tmdbReleaseDate); err == nil {
		m.ReleaseDate = releaseDate
	} else {
		slog.Error("AsMedia: Failed to parse release date", "name", m.Name, "error", err)
	}
	return m
}

//
// Content Credits
//

type ContentCredits struct {
	ID   int `json:"id"`
	Cast []struct {
		Adult              bool    `json:"adult"`
		Gender             int     `json:"gender"`
		ID                 int     `json:"id"`
		KnownForDepartment string  `json:"known_for_department"`
		Name               string  `json:"name"`
		OriginalName       string  `json:"original_name"`
		Popularity         float64 `json:"popularity"`
		ProfilePath        string  `json:"profile_path"`
		CastID             int     `json:"cast_id"`
		Character          string  `json:"character"`
		CreditID           string  `json:"credit_id"`
		Order              int     `json:"order"`
	} `json:"cast"`
	Crew []struct {
		Adult              bool    `json:"adult"`
		Gender             int     `json:"gender"`
		ID                 int     `json:"id"`
		KnownForDepartment string  `json:"known_for_department"`
		Name               string  `json:"name"`
		OriginalName       string  `json:"original_name"`
		Popularity         float64 `json:"popularity"`
		ProfilePath        string  `json:"profile_path"`
		CreditID           string  `json:"credit_id"`
		Department         string  `json:"department"`
		Job                string  `json:"job"`
	} `json:"crew"`
}

type AggregateCredits struct {
	ID   int `json:"id"`
	Cast []struct {
		Adult              bool    `json:"adult"`
		Gender             int     `json:"gender"`
		ID                 int     `json:"id"`
		KnownForDepartment string  `json:"known_for_department"`
		Name               string  `json:"name"`
		OriginalName       string  `json:"original_name"`
		Popularity         float64 `json:"popularity"`
		ProfilePath        string  `json:"profile_path"`
		TotalEpisodeCount  int     `json:"total_episode_count"`
		Order              int     `json:"order"`
		Roles              []struct {
			CreditID     string `json:"credit_id"`
			Character    string `json:"character"`
			EpisodeCount int    `json:"episode_count"`
		} `json:"roles"`
	} `json:"cast"`
	Crew []struct {
		Adult              bool    `json:"adult"`
		Gender             int     `json:"gender"`
		ID                 int     `json:"id"`
		KnownForDepartment string  `json:"known_for_department"`
		Name               string  `json:"name"`
		OriginalName       string  `json:"original_name"`
		Popularity         float64 `json:"popularity"`
		ProfilePath        string  `json:"profile_path"`
		Department         string  `json:"department"`
		TotalEpisodeCount  int     `json:"total_episode_count"`
		Jobs               []struct {
			CreditID     string `json:"credit_id"`
			Job          string `json:"job"`
			EpisodeCount int    `json:"episode_count"`
		} `json:"jobs"`
	} `json:"crew"`
}

func (a *AggregateCredits) ToContentCredits() ContentCredits {
	cc := ContentCredits{ID: a.ID}
	for _, c := range a.Cast {
		var character, creditID string
		if len(c.Roles) > 0 {
			creditID = c.Roles[0].CreditID
			character = c.Roles[0].Character
		}
		cc.Cast = append(cc.Cast, struct {
			Adult              bool    `json:"adult"`
			Gender             int     `json:"gender"`
			ID                 int     `json:"id"`
			KnownForDepartment string  `json:"known_for_department"`
			Name               string  `json:"name"`
			OriginalName       string  `json:"original_name"`
			Popularity         float64 `json:"popularity"`
			ProfilePath        string  `json:"profile_path"`
			CastID             int     `json:"cast_id"`
			Character          string  `json:"character"`
			CreditID           string  `json:"credit_id"`
			Order              int     `json:"order"`
		}{
			Adult:              c.Adult,
			Gender:             c.Gender,
			ID:                 c.ID,
			KnownForDepartment: c.KnownForDepartment,
			Name:               c.Name,
			OriginalName:       c.OriginalName,
			Popularity:         c.Popularity,
			ProfilePath:        c.ProfilePath,
			Character:          character,
			CreditID:           creditID,
			Order:              c.Order,
		})
	}
	for _, cr := range a.Crew {
		for _, j := range cr.Jobs {
			cc.Crew = append(cc.Crew, struct {
				Adult              bool    `json:"adult"`
				Gender             int     `json:"gender"`
				ID                 int     `json:"id"`
				KnownForDepartment string  `json:"known_for_department"`
				Name               string  `json:"name"`
				OriginalName       string  `json:"original_name"`
				Popularity         float64 `json:"popularity"`
				ProfilePath        string  `json:"profile_path"`
				CreditID           string  `json:"credit_id"`
				Department         string  `json:"department"`
				Job                string  `json:"job"`
			}{
				Adult:              cr.Adult,
				Gender:             cr.Gender,
				ID:                 cr.ID,
				KnownForDepartment: cr.KnownForDepartment,
				Name:               cr.Name,
				OriginalName:       cr.OriginalName,
				Popularity:         cr.Popularity,
				ProfilePath:        cr.ProfilePath,
				CreditID:           j.CreditID,
				Department:         cr.Department,
				Job:                j.Job,
			})
		}
	}
	return cc
}

//
// Discover All Trending
//

type TrendingType string

const (
	TrendingTypeAll    TrendingType = "all"
	TrendingTypeMovie  TrendingType = "movie"
	TrendingTypeShow   TrendingType = "tv"
	TrendingTypePerson TrendingType = "person"
)

type TrendingCombined struct {
	SearchResponse[TrendingCombinedResult]
}

type TrendingCombinedResult struct {
	SearchResult
	Adult            bool     `json:"adult"`
	BackdropPath     string   `json:"backdrop_path"`
	Title            string   `json:"title,omitempty"`
	Name             string   `json:"name,omitempty"`
	OriginalLanguage string   `json:"original_language"`
	OriginalTitle    string   `json:"original_title,omitempty"`
	GenreIds         []int    `json:"genre_ids"`
	Popularity       float64  `json:"popularity"`
	ReleaseDate      string   `json:"release_date,omitempty"`
	Video            bool     `json:"video,omitempty"`
	OriginalName     string   `json:"original_name,omitempty"`
	FirstAirDate     string   `json:"first_air_date,omitempty"`
	OriginCountry    []string `json:"origin_country,omitempty"`
	ProfilePath      string   `json:"profile_path"`
}

func (t *TrendingCombinedResult) AsMedia() domain.Media {
	m := t.SearchResult.AsMedia()
	m.ExtBackdropPath = t.BackdropPath
	m.OriginalLanguage = t.OriginalLanguage

	m.Name = t.Title
	m.OriginalTitle = t.OriginalTitle
	if t.Name != "" {
		m.Name = t.Name
	}
	if t.OriginalName != "" {
		m.OriginalTitle = t.OriginalName
	}

	var tmdbReleaseDate string
	switch t.MediaType {
	case "movie":
		tmdbReleaseDate = t.ReleaseDate
	case "tv":
		tmdbReleaseDate = t.FirstAirDate
	case "person":
		m.ExtPosterPath = t.ProfilePath
	}
	if tmdbReleaseDate != "" {
		if releaseDate, err := time.Parse("2006-01-02", tmdbReleaseDate); err == nil {
			m.ReleaseDate = releaseDate
		} else {
			slog.Error("AsMedia: Failed to parse release date", "name", m.Name, "error", err)
		}
	}
	return m
}

//
// Discover
//

type DiscoverOptions struct {
	// Release date greater than.
	ReleaseDateMin time.Time
	// Release date less than.
	ReleaseDateMax time.Time
	// With release type.
	// Release types are listed on this page:
	// https://developer.themoviedb.org/reference/movie-release-dates
	WithReleaseType string
	// Language for metadata.
	Language string
}

//
// Discover Movies
//

type DiscoverMovies struct {
	SearchResponse[DiscoverMoviesResult]
}

type DiscoverMoviesResult struct {
	Adult            bool    `json:"adult"`
	BackdropPath     string  `json:"backdrop_path"`
	GenreIds         []int   `json:"genre_ids"`
	ID               int     `json:"id"`
	OriginalLanguage string  `json:"original_language"`
	OriginalTitle    string  `json:"original_title"`
	Overview         string  `json:"overview"`
	Popularity       float64 `json:"popularity"`
	PosterPath       string  `json:"poster_path"`
	ReleaseDate      string  `json:"release_date"`
	Title            string  `json:"title"`
	Video            bool    `json:"video"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
}

func (t *DiscoverMoviesResult) AsMedia() domain.Media {
	m := domain.Media{
		IDs: domain.MediaIDs{
			TMDB: t.ID,
		},
		Type:             domain.MediaTypeTMDBMovie,
		Name:             t.Title,
		OriginalTitle:    t.OriginalTitle,
		OriginalLanguage: t.OriginalLanguage,
		Summary:          t.Overview,
		ExtPosterPath:    t.PosterPath,
		ExtBackdropPath:  t.BackdropPath,
		Rating:           uint(t.VoteAverage * 10),
		RatingCount:      uint(t.VoteCount),
	}
	if releaseDate, err := time.Parse("2006-01-02", t.ReleaseDate); err == nil {
		m.ReleaseDate = releaseDate
	} else {
		slog.Error("AsMedia: Failed to parse release date", "name", m.Name, "error", err)
	}
	return m
}

//
// Discover Shows
//

type DiscoverShows struct {
	SearchResponse[DiscoverShowsResult]
}

type DiscoverShowsResult struct {
	BackdropPath     string   `json:"backdrop_path"`
	FirstAirDate     string   `json:"first_air_date"`
	GenreIds         []int    `json:"genre_ids"`
	ID               int      `json:"id"`
	Name             string   `json:"name"`
	OriginCountry    []string `json:"origin_country"`
	OriginalLanguage string   `json:"original_language"`
	OriginalName     string   `json:"original_name"`
	Overview         string   `json:"overview"`
	Popularity       float64  `json:"popularity"`
	PosterPath       string   `json:"poster_path"`
	VoteAverage      float32  `json:"vote_average"`
	VoteCount        int      `json:"vote_count"`
}

func (t *DiscoverShowsResult) AsMedia() domain.Media {
	m := domain.Media{
		IDs: domain.MediaIDs{
			TMDB: t.ID,
		},
		Type:             domain.MediaTypeTMDBShow,
		Name:             t.Name,
		OriginalTitle:    t.OriginalName,
		OriginalLanguage: t.OriginalLanguage,
		Summary:          t.Overview,
		ExtPosterPath:    t.PosterPath,
		ExtBackdropPath:  t.BackdropPath,
		Rating:           uint(t.VoteAverage * 10),
		RatingCount:      uint(t.VoteCount),
	}
	if releaseDate, err := time.Parse("2006-01-02", t.FirstAirDate); err == nil {
		m.ReleaseDate = releaseDate
	} else {
		slog.Error("AsMedia: Failed to parse release date", "name", m.Name, "error", err)
	}
	return m
}

//
// Discover Shows
//

type PopularPeople struct {
	SearchResponse[PopularPeopleResult]
}

type PopularPeopleResult struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ProfilePath string `json:"profile_path"`
}

func (t *PopularPeopleResult) AsMedia() domain.Media {
	m := domain.Media{
		IDs: domain.MediaIDs{
			TMDB: t.ID,
		},
		Type:          domain.MediaTypeTMDBPerson,
		Name:          t.Name,
		ExtPosterPath: t.ProfilePath,
	}
	return m
}

//
//
//

type MovieStatus string

const (
	MovieStatusRumored        MovieStatus = "Rumored"
	MovieStatusPlanned        MovieStatus = "Planned"
	MovieStatusInProduction   MovieStatus = "In Production"
	MovieStatusPostProduction MovieStatus = "Post Production"
	MovieStatusReleased       MovieStatus = "Released"
	MovieStatusCanceled       MovieStatus = "Canceled"
)

type ShowStatus string

const (
	ShowStatusReturningSeries ShowStatus = "Returning Series"
	ShowStatusPlanned         ShowStatus = "Planned"
	ShowStatusInProduction    ShowStatus = "In Production"
	ShowStatusEnded           ShowStatus = "Ended"
	ShowStatusCanceled        ShowStatus = "Canceled"
	ShowStatusPilot           ShowStatus = "Pilot"
)

// Watch providers in one country
type WatchProviders struct {
	// Subscription services
	Flatrate []WatchProvider `json:"flatrate"`
	// Free providers
	Free []WatchProvider `json:"free"`
	// Ad-supported free providers
	Ads []WatchProvider `json:"ads"`
	// Rental providers
	Rent []WatchProvider `json:"rent"`
	// Digital purchase providers
	Buy []WatchProvider `json:"buy"`
	// Link to view all streaming options on tmdb
	Link string `json:"link"`
}

type WatchProvider struct {
	ProviderID      int    `json:"provider_id"`
	ProviderName    string `json:"provider_name"`
	LogoPath        string `json:"logo_path"`
	DisplayPriority int    `json:"display_priority"`
}

type ContentVideos struct {
	ID      int `json:"id"`
	Results []struct {
		Iso6391     string    `json:"iso_639_1"`
		Iso31661    string    `json:"iso_3166_1"`
		Name        string    `json:"name"`
		Key         string    `json:"key"`
		Site        string    `json:"site"`
		Size        int       `json:"size"`
		Type        string    `json:"type"`
		Official    bool      `json:"official"`
		PublishedAt time.Time `json:"published_at"`
		ID          string    `json:"id"`
	} `json:"results"`
}

type ExternalIds struct {
	ImdbID      string `json:"imdb_id"`
	WikidataID  string `json:"wikidata_id"`
	FacebookID  string `json:"facebook_id"`
	InstagramID string `json:"instagram_id"`
	TwitterID   string `json:"twitter_id"`
}

type ExternalIdsMovie struct {
	ExternalIds
}

type ExternalIdsShow struct {
	ExternalIds
	FreebaseMid string `json:"freebase_mid"`
	FreebaseID  string `json:"freebase_id"`
	TvdbID      int    `json:"tvdb_id"`
	TvrageID    int    `json:"tvrage_id"`
}

type Keywords struct {
	// ID      int `json:"id"`
	Results []struct {
		Name string `json:"name"`
		ID   int    `json:"id"`
	} `json:"results"`
}

type Regions struct {
	Results []struct {
		ISO3166_1    string `json:"iso_3166_1"`
		English_Name string `json:"english_name"`
		Native_Name  string `json:"native_name"`
	} `json:"results"`
}
