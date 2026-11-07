package addon

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/router"
)

type Router struct {
	br *router.BaseRouter
	s  *Service
}

func NewRouter(br *router.BaseRouter, service *Service) *Router {
	return &Router{
		br: br,
		s:  service,
	}
}

// AddonAuthMiddleware authenticates using token from path parameter or query string
func (r *Router) AddonAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		if token == "" {
			token = c.Query("token")
		}
		if token == "" {
			token = c.Query("apiKey")
		}
		if token == "" {
			headerToken := c.GetHeader("Authorization")
			if strings.HasPrefix(headerToken, "Bearer ") {
				token = strings.TrimPrefix(headerToken, "Bearer ")
			}
		}

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, router.ErrorResponse{
				Error: "Token do Addon não fornecido.",
			})
			return
		}

		userId, err := r.s.Authenticate(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, router.ErrorResponse{
				Error: "Token de autenticação do Addon inválido ou expirado.",
			})
			return
		}

		c.Set("userId", userId)
		c.Next()
	}
}

func (r *Router) AddRoutes() {
	authMw := r.AddonAuthMiddleware()

	// CORS support specifically for Stremio / NuvioTV webview and TV clients
	addonGroup := r.br.Router.Group("/addon/:token")
	addonGroup.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, If-None-Match")
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	{
		// 1. Manifest (public metadata - never 401s so NuvioTV can always read the catalog list)
		addonGroup.GET("/manifest.json", r.GetManifest)

		// 2. Catalog (supports /catalog/:type/:id.json and /catalog/:type/:id/:extra.json)
		addonGroup.GET("/catalog/:type/:id", authMw, r.GetCatalog)
		addonGroup.GET("/catalog/:type/:id/:extra", authMw, r.GetCatalog)

		// 3. Meta (matches both /meta/:type/:id and with extra)
		addonGroup.GET("/meta/:type/:id", authMw, r.GetMeta)
		addonGroup.GET("/meta/:type/:id/:extra", authMw, r.GetMeta)

		// 4. Stream - Active Scrobble playback intercept (matches both /stream/:type/:id and with extra)
		addonGroup.GET("/stream/:type/:id", authMw, r.GetStream)
		addonGroup.GET("/stream/:type/:id/:extra", authMw, r.GetStream)

		// 5. Scrobble / Playback hook
		addonGroup.POST("/scrobble", authMw, r.HandleScrobble)

		// 6. Simkl-compatible scrobble endpoint for NuvioTV native fallback
		addonGroup.POST("/scrobble/stop", authMw, r.HandleSimklScrobble)
		addonGroup.POST("/scrobble/start", authMw, r.HandleSimklScrobbleStart)
	}
}

func (r *Router) GetManifest(c *gin.Context) {
	manifest := r.s.GetManifest()
	c.JSON(http.StatusOK, manifest)
}

func (r *Router) GetCatalog(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	mediaType := c.Param("type")
	catalogId := c.Param("id")
	catalogId = strings.TrimSuffix(catalogId, ".json")
	extra := c.Param("extra")

	skip := parseSkip(extra, c.Query("skip"))

	resp, etag, err := r.s.GetCatalog(userId, mediaType, catalogId, skip)
	if err != nil {
		c.JSON(http.StatusOK, CatalogResponse{Metas: []MetaPreview{}})
		return
	}

	c.Header("ETag", etag)
	c.Header("Cache-Control", "private, no-cache")

	ifNoneMatch := c.GetHeader("If-None-Match")
	if ifNoneMatch != "" && ifNoneMatch == etag {
		c.Status(http.StatusNotModified)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func parseSkip(extra string, querySkip string) int {
	if querySkip != "" {
		if val, err := strconv.Atoi(querySkip); err == nil && val > 0 {
			return val
		}
	}
	extra = strings.TrimSuffix(extra, ".json")
	for _, part := range strings.Split(extra, "&") {
		if strings.HasPrefix(part, "skip=") {
			if val, err := strconv.Atoi(strings.TrimPrefix(part, "skip=")); err == nil && val > 0 {
				return val
			}
		}
	}
	return 0
}

func (r *Router) GetMeta(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	mediaType := c.Param("type")
	id := c.Param("id")
	id = strings.TrimSuffix(id, ".json")

	resp, err := r.s.GetMeta(userId, mediaType, id)
	if err != nil {
		c.JSON(http.StatusNotFound, router.ErrorResponse{
			Error: fmt.Sprintf("Metadados não encontrados: %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetStream intercepts Stremio stream requests to record playback start in real time,
// returning an empty stream array so standard stream providers handle media delivery.
func (r *Router) GetStream(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	mediaType := c.Param("type")
	id := c.Param("id")
	id = strings.TrimSuffix(id, ".json")

	// Intercept stream in background/non-blocking for video player
	_ = r.s.HandleStreamPlayback(userId, mediaType, id)

	c.JSON(http.StatusOK, StreamResponse{Streams: []Stream{}})
}

func (r *Router) HandleScrobble(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	var payload ScrobblePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "Payload de scrobble inválido."})
		return
	}

	if err := r.s.HandleScrobble(userId, payload); err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Progresso sincronizado com sucesso no Watcharr"})
}

// HandleSimklScrobbleStart handles Simkl-format start scrobbles by marking media as WATCHING
func (r *Router) HandleSimklScrobbleStart(c *gin.Context) {
	userId := c.MustGet("userId").(uint)

	var simklBody struct {
		Movie *struct {
			Title string `json:"title"`
			Year  int    `json:"year"`
			IDs   struct {
				TMDB int    `json:"tmdb"`
				IMDb string `json:"imdb"`
			} `json:"ids"`
		} `json:"movie"`
		Show *struct {
			Title string `json:"title"`
			Year  int    `json:"year"`
			IDs   struct {
				TMDB int    `json:"tmdb"`
				IMDb string `json:"imdb"`
			} `json:"ids"`
		} `json:"show"`
		Episode *struct {
			Season int `json:"season"`
			Number int `json:"number"`
		} `json:"episode"`
		Progress float64 `json:"progress"`
	}

	if err := c.ShouldBindJSON(&simklBody); err != nil {
		c.JSON(http.StatusOK, gin.H{"result": "OK"})
		return
	}

	if simklBody.Movie != nil {
		rawId := simklBody.Movie.IDs.IMDb
		if simklBody.Movie.IDs.TMDB > 0 {
			rawId = fmt.Sprintf("tmdb:%d", simklBody.Movie.IDs.TMDB)
		}
		if rawId != "" {
			_ = r.s.HandleStreamPlayback(userId, "movie", rawId)
		}
	} else if simklBody.Show != nil {
		rawId := simklBody.Show.IDs.IMDb
		if simklBody.Show.IDs.TMDB > 0 {
			rawId = fmt.Sprintf("tmdb:%d", simklBody.Show.IDs.TMDB)
		}
		if rawId != "" {
			if simklBody.Episode != nil && simklBody.Episode.Season > 0 && simklBody.Episode.Number > 0 {
				rawId = fmt.Sprintf("%s:%d:%d", rawId, simklBody.Episode.Season, simklBody.Episode.Number)
			}
			_ = r.s.HandleStreamPlayback(userId, "series", rawId)
		}
	}

	c.JSON(http.StatusOK, gin.H{"result": "OK", "action": "scrobble_started"})
}

// HandleSimklScrobble handles Simkl-format scrobbles if NuvioTV is configured to direct to this host
func (r *Router) HandleSimklScrobble(c *gin.Context) {
	if strings.HasSuffix(c.Request.URL.Path, "/start") {
		r.HandleSimklScrobbleStart(c)
		return
	}

	userId := c.MustGet("userId").(uint)

	var simklBody struct {
		Movie *struct {
			Title string `json:"title"`
			Year  int    `json:"year"`
			IDs   struct {
				TMDB int    `json:"tmdb"`
				IMDb string `json:"imdb"`
			} `json:"ids"`
		} `json:"movie"`
		Show *struct {
			Title string `json:"title"`
			Year  int    `json:"year"`
			IDs   struct {
				TMDB int    `json:"tmdb"`
				IMDb string `json:"imdb"`
			} `json:"ids"`
		} `json:"show"`
		Episode *struct {
			Season int `json:"season"`
			Number int `json:"number"`
		} `json:"episode"`
		Progress float64 `json:"progress"`
	}

	if err := c.ShouldBindJSON(&simklBody); err != nil {
		c.JSON(http.StatusOK, gin.H{"result": "OK"})
		return
	}

	payload := ScrobblePayload{
		Event:    "stop",
		Progress: simklBody.Progress,
	}

	if simklBody.Movie != nil {
		payload.MediaType = "movie"
		payload.TmdbID = simklBody.Movie.IDs.TMDB
		payload.ImdbID = simklBody.Movie.IDs.IMDb
	} else if simklBody.Show != nil {
		payload.MediaType = "series"
		payload.TmdbID = simklBody.Show.IDs.TMDB
		payload.ImdbID = simklBody.Show.IDs.IMDb
		if simklBody.Episode != nil {
			payload.SeasonNumber = simklBody.Episode.Season
			payload.EpisodeNumber = simklBody.Episode.Number
		}
	}

	progress := payload.Progress
	if progress >= 0.8 && progress <= 1.0 {
		progress = progress * 100.0
		payload.Progress = progress
	}

	if payload.Progress >= 80.0 {
		_ = r.s.HandleScrobble(userId, payload)
	}

	c.JSON(http.StatusOK, gin.H{"result": "OK", "action": "scrobble_recorded"})
}
