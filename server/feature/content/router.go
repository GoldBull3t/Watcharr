package content

import (
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/auth/authmiddleware"
	"github.com/sbondCo/Watcharr/feature/user/usermiddleware"
	"github.com/sbondCo/Watcharr/feature/watched/addedtocontent"
	"github.com/sbondCo/Watcharr/media/extratings"
	"github.com/sbondCo/Watcharr/media/tmdb"
	"github.com/sbondCo/Watcharr/router"
	"github.com/sbondCo/Watcharr/util"
)

type WatchedProvider interface {
	UpdateWatchedLastViewedSeason(userId uint, id uint, seasonNum int) error
	GetWatchedItemBySupportedMediaId(userId uint, id uint, t util.SupportedMedia) (entity.Watched, error)
	GetWatchedItemsBySupportedMediaIds(userId uint, c []addedtocontent.IdToTypePair) ([]entity.Watched, error)
}

type Router struct {
	br   *router.BaseRouter
	cs   *Service
	wp   WatchedProvider
	tmdb *tmdb.TMDB
}

func NewRouter(
	br *router.BaseRouter,
	cs *Service,
	wp WatchedProvider,
	tmdb *tmdb.TMDB,
) *Router {
	return &Router{
		br:   br,
		cs:   cs,
		wp:   wp,
		tmdb: tmdb,
	}
}

func (r *Router) AddRoutes() {
	content := r.br.Router.Group("/content").Use(authmiddleware.AuthRequired(r.br.Cfg))

	// NOTE: Some routes previously used `cache.CachePage`, but with per-user language
	// settings, TMDB client methods handle their own caching via ContentStore.

	// Get movie details (for movie page)
	content.GET("/movie/:id", usermiddleware.WithUser(r.br.DB), router.WhereaboutsRequired(r.br.Cfg), router.LanguageRequired(r.br.Cfg), r.GetMovieDetails)
	// Get movie cast
	content.GET("/movie/:id/credits", usermiddleware.WithUser(r.br.DB), router.LanguageRequired(r.br.Cfg), r.GetMovieCredits)
	// Get tv details (for tv page)
	content.GET("/tv/:id", usermiddleware.WithUser(r.br.DB), router.WhereaboutsRequired(r.br.Cfg), router.LanguageRequired(r.br.Cfg), r.GetTvDetails)
	// Get tv cast
	content.GET("/tv/:id/credits", usermiddleware.WithUser(r.br.DB), router.LanguageRequired(r.br.Cfg), r.GetTvCredits)
	// Get season details
	// Supports `watchedId` query parameter for saving the requested season as `LastViewedSeason`.
	content.GET("/tv/:id/season/:num", usermiddleware.WithUser(r.br.DB), router.LanguageRequired(r.br.Cfg), r.GetSeasonDetails)
	// Get person details
	content.GET("/person/:id", usermiddleware.WithUser(r.br.DB), router.LanguageRequired(r.br.Cfg), r.GetPerson)
	// Get person credits
	content.GET("/person/:id/credits", usermiddleware.WithUser(r.br.DB), router.LanguageRequired(r.br.Cfg), r.GetPersonCredits)
	// Available regions for watch providers
	content.GET("/regions", r.GetRegions)
	// Available languages for TMDB metadata
	content.GET("/languages", r.GetLanguages)
	// Get movie collection / series details
	content.GET("/collection/:id", usermiddleware.WithUser(r.br.DB), router.LanguageRequired(r.br.Cfg), r.GetCollection)
	// Update or set ending type override for tv show
	content.POST("/tv/:id/ending-type", usermiddleware.WithUser(r.br.DB), r.UpdateTvEndingType)
}

func (r *Router) GetMovieDetails(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	if c.Param("id") == "" {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "an id was not provided"})
		return
	}
	userCountry := c.MustGet("userCountry").(string)
	content, err := r.tmdb.MovieDetails(tmdb.MovieDetailsOptions{
		ID:      c.Param("id"),
		Country: userCountry,
		Params: map[string]string{
			"append_to_response": "videos,watch/providers,similar,recommendations,release_dates,external_ids",
			"language":           c.MustGet("userLanguage").(string),
		},
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	contentAsMedia := content.AsMedia()
	if cert := content.ReleaseDates.ExtractCertification(userCountry); cert != "" {
		contentAsMedia.Certification = cert
	}

	isAnime := content.OriginalLanguage == "ja"
	if isAnime {
		isAnime = false
		for _, g := range content.Genres {
			if g.ID == 16 || strings.EqualFold(g.Name, "Animation") || strings.EqualFold(g.Name, "Animação") {
				isAnime = true
				break
			}
		}
	}
	extRatings, fallbackCert := extratings.FetchEnrichments(extratings.FetchOptions{
		MediaType:  "movie",
		TMDBID:     content.ID,
		ImdbID:     contentAsMedia.IDs.IMDB,
		Title:      content.Title,
		OrigTitle:  content.OriginalTitle,
		IsAnime:    isAnime,
		OMDBKey:    r.br.Cfg.GetOMDBKey(),
		MDBListKey: r.br.Cfg.GetMDBListKey(),
	})
	contentAsMedia.ExternalRatings = extRatings
	if contentAsMedia.Certification == "" && fallbackCert != "" {
		contentAsMedia.Certification = fallbackCert
	}

	if err := addedtocontent.AddSingularAndList(
		r.wp,
		userId,
		contentAsMedia,
		func(w *entity.Watched) {
			contentAsMedia.Watched = domain.NewWatchedDtoForContentPage(w)
		},
		[]*addedtocontent.AddListCall[domain.Media]{
			addedtocontent.NewAddListCall(
				contentAsMedia.Similar,
				func(i int, w *entity.Watched) {
					contentAsMedia.Similar[i].Watched = domain.NewWatchedDtoForLists(w)
				},
			),
			addedtocontent.NewAddListCall(
				contentAsMedia.Recommendations,
				func(i int, w *entity.Watched) {
					contentAsMedia.Recommendations[i].Watched = domain.NewWatchedDtoForLists(w)
				},
			),
		},
	); err != nil {
		slog.Error("GetMovieDetails: Failed to add watched to content!", "error", err)
		c.JSON(
			http.StatusInternalServerError,
			router.ErrorResponse{Error: "failed to add watched data to response"},
		)
		return
	}
	c.JSON(http.StatusOK, contentAsMedia)
}

func (r *Router) GetMovieCredits(c *gin.Context) {
	if c.Param("id") == "" {
		c.Status(400)
		return
	}
	content, err := r.tmdb.MovieCredits(c.Param("id"), map[string]string{
		"language": c.MustGet("userLanguage").(string),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, content)
}

func (r *Router) GetTvDetails(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	if c.Param("id") == "" {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "an id was not provided"})
		return
	}
	userCountry := c.MustGet("userCountry").(string)
	// 1. Get details
	content, err := r.tmdb.ShowDetails(tmdb.ShowDetailsOptions{
		ID:      c.Param("id"),
		Country: userCountry,
		Params: map[string]string{
			"append_to_response": "videos,watch/providers,similar,recommendations,content_ratings,external_ids,keywords",
			"language":           c.MustGet("userLanguage").(string),
		},
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	contentAsMedia := content.AsMedia()
	if cert := content.ContentRatings.ExtractCertification(userCountry); cert != "" {
		contentAsMedia.Certification = cert
	}

	// Cancellation Guard: Evaluate show status, cliffhanger database, prolonged hiatus, and database overrides
	var dbEndingType string
	var cachedContent entity.Content
	if err := r.br.DB.Where("tmdb_id = ? AND type = ?", content.ID, entity.SHOW).Select("ending_type").First(&cachedContent).Error; err == nil {
		dbEndingType = cachedContent.EndingType
	}
	cancInfo := EvaluateShowCancellation(content, dbEndingType)
	contentAsMedia.Cancellation = &cancInfo

	extRatings, fallbackCert := extratings.FetchEnrichments(extratings.FetchOptions{
		MediaType:  "tv",
		TMDBID:     content.ID,
		ImdbID:     contentAsMedia.IDs.IMDB,
		TvdbID:     contentAsMedia.IDs.TVDB,
		Title:      content.Name,
		OrigTitle:  content.OriginalName,
		IsAnime:    contentAsMedia.IsShowAnime,
		OMDBKey:    r.br.Cfg.GetOMDBKey(),
		MDBListKey: r.br.Cfg.GetMDBListKey(),
	})
	contentAsMedia.ExternalRatings = extRatings
	if contentAsMedia.Certification == "" && fallbackCert != "" {
		contentAsMedia.Certification = fallbackCert
	}

	if err := addedtocontent.AddSingularAndList(
		r.wp,
		userId,
		contentAsMedia,
		func(w *entity.Watched) {
			contentAsMedia.Watched = domain.NewWatchedDtoForContentPage(w)
		},
		[]*addedtocontent.AddListCall[domain.Media]{
			addedtocontent.NewAddListCall(
				contentAsMedia.Similar,
				func(i int, w *entity.Watched) {
					contentAsMedia.Similar[i].Watched = domain.NewWatchedDtoForLists(w)
				},
			),
			addedtocontent.NewAddListCall(
				contentAsMedia.Recommendations,
				func(i int, w *entity.Watched) {
					contentAsMedia.Recommendations[i].Watched = domain.NewWatchedDtoForLists(w)
				},
			),
		},
	); err != nil {
		slog.Error("GetTvDetails: Failed to add watched to content!", "error", err)
		c.JSON(
			http.StatusInternalServerError,
			router.ErrorResponse{Error: "failed to add watched data to response"},
		)
		return
	}
	c.JSON(http.StatusOK, contentAsMedia)
}

type UpdateEndingTypeRequest struct {
	EndingType string `json:"endingType" binding:"required"`
}

func (r *Router) UpdateTvEndingType(c *gin.Context) {
	idStr := c.Param("id")
	tmdbId, err := strconv.Atoi(idStr)
	if err != nil || tmdbId <= 0 {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "invalid show id"})
		return
	}
	var req UpdateEndingTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "endingType is required"})
		return
	}
	endingType := strings.ToLower(strings.TrimSpace(req.EndingType))
	res := r.br.DB.Model(&entity.Content{}).
		Where("tmdb_id = ? AND type = ?", tmdbId, entity.SHOW).
		Update("ending_type", endingType)
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: "failed to update ending type"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"tmdbId":     tmdbId,
		"endingType": endingType,
	})
}

func (r *Router) GetTvCredits(c *gin.Context) {
	if c.Param("id") == "" {
		c.Status(400)
		return
	}
	content, err := r.tmdb.ShowCredits(c.Param("id"), map[string]string{
		"language": c.MustGet("userLanguage").(string),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, content)
}

// Get season details
// Supports `watchedId` query parameter for saving the requested season as `LastViewedSeason`.
func (r *Router) GetSeasonDetails(c *gin.Context) {
	if c.Param("id") == "" || c.Param("num") == "" {
		c.Status(400)
		return
	}
	content, err := r.tmdb.SeasonDetails(c.Param("id"), c.Param("num"), map[string]string{
		"language": c.MustGet("userLanguage").(string),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	// If a `watchedId` is passed, we should update it with this season
	// number, so the LastViewedSeason field is up to date (this seemed
	// better than making a new request for just saving this).
	// We will attach a `watcharr-lastviewedseason-saved` header if
	// this part succeeds so the client can decide on showing an error.
	if watchedIdQ := c.Query("watchedId"); watchedIdQ != "" {
		userId := c.MustGet("userId").(uint)
		watchedId, err := strconv.ParseUint(watchedIdQ, 10, 64)
		if err != nil {
			slog.Error("get season details route: Processing watchedId param failed", "error", err.Error(), "id", watchedIdQ)
		} else {
			if seasonNum, err := strconv.ParseInt(c.Param("num"), 10, 64); err == nil {
				if err = r.wp.UpdateWatchedLastViewedSeason(
					userId, uint(watchedId), int(seasonNum),
				); err == nil {
					c.Header("watcharr-lastviewedseason-saved", "1")
				}
			} else {
				slog.Error("get season details route: Parsing season number as int failed", "error", err.Error(), "season_num", c.Param("num"))
			}
		}
	} else {
		slog.Debug("get season details route: No watchedId parameter found.. not doing anything.")
	}
	c.JSON(http.StatusOK, content)
}

func (r *Router) GetPerson(c *gin.Context) {
	if c.Param("id") == "" {
		c.Status(400)
		return
	}
	content, err := r.tmdb.PersonDetails(c.Param("id"), map[string]string{
		"language": c.MustGet("userLanguage").(string),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, content.AsPersonDetailsResponse())
}

func (r *Router) GetPersonCredits(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	if c.Param("id") == "" {
		c.Status(400)
		return
	}
	content, err := r.tmdb.PersonCredits(c.Param("id"), map[string]string{
		"language": c.MustGet("userLanguage").(string),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	// Add content into response struct then add watched entries
	resp := domain.PersonCreditsResponse{}
	for i := range content.Cast {
		resp.Credits = append(resp.Credits, content.Cast[i].AsMedia())
	}
	if err := addedtocontent.AddList(
		r.wp,
		userId,
		resp.Credits,
		func(i int, w *entity.Watched) {
			resp.Credits[i].Watched = domain.NewWatchedDtoForLists(w)
		},
	); err != nil {
		slog.Error("GetPersonCredits: Failed to add watched to content!", "error", err)
		c.JSON(
			http.StatusInternalServerError,
			router.ErrorResponse{Error: "failed to add watched data to response"},
		)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (r *Router) GetRegions(c *gin.Context) {
	re, err := r.tmdb.Regions()
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, re)
}

func (r *Router) GetLanguages(c *gin.Context) {
	langs, err := r.tmdb.PrimaryTranslations()
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, langs)
}

func (r *Router) GetCollection(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	if c.Param("id") == "" {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "an id was not provided"})
		return
	}
	collectionData, err := r.tmdb.Collection(c.Param("id"), map[string]string{
		"language": c.MustGet("userLanguage").(string),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	// Sort parts chronologically by release date
	sort.Slice(collectionData.Parts, func(i, j int) bool {
		return collectionData.Parts[i].ReleaseDate < collectionData.Parts[j].ReleaseDate
	})

	collectionMedia := collectionData.AsMediaCollection()
	// Add user's watched status to each part
	if len(collectionMedia.Parts) > 0 {
		if err := addedtocontent.AddList(
			r.wp,
			userId,
			collectionMedia.Parts,
			func(i int, w *entity.Watched) {
				collectionMedia.Parts[i].Watched = domain.NewWatchedDtoForLists(w)
			},
		); err != nil {
			slog.Error("GetCollection: Failed to add watched to parts!", "error", err)
		}
	}
	c.JSON(http.StatusOK, collectionMedia)
}
