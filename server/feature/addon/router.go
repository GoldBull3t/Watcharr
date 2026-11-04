package addon

import (
	"fmt"
	"net/http"
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
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
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

		// 4. Scrobble / Playback hook
		addonGroup.POST("/scrobble", authMw, r.HandleScrobble)

		// 5. Simkl-compatible scrobble endpoint for NuvioTV native fallback
		addonGroup.POST("/scrobble/stop", authMw, r.HandleSimklScrobble)
		addonGroup.POST("/scrobble/start", authMw, r.HandleSimklScrobble)
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

	resp, err := r.s.GetCatalog(userId, mediaType, catalogId)
	if err != nil {
		c.JSON(http.StatusOK, CatalogResponse{Metas: []MetaPreview{}})
		return
	}

	c.JSON(http.StatusOK, resp)
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

// HandleSimklScrobble handles Simkl-format scrobbles if NuvioTV is configured to direct to this host
func (r *Router) HandleSimklScrobble(c *gin.Context) {
	userId := c.MustGet("userId").(uint)

	// Simkl scrobble body structure
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
		// Non-blocking response
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

	if payload.Progress >= 80.0 {
		_ = r.s.HandleScrobble(userId, payload)
	}

	c.JSON(http.StatusOK, gin.H{"result": "OK", "action": "scrobble_recorded"})
}

