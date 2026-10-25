package integration

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/feature/auth/authmiddleware"
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

// IntegrationAuthMiddleware extracts API key or JWT from header, query param, or path param
func (r *Router) IntegrationAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if strings.HasPrefix(token, "Bearer ") {
			token = strings.TrimPrefix(token, "Bearer ")
		}
		if token == "" {
			token = c.Query("apiKey")
		}
		if token == "" {
			token = c.Query("token")
		}
		if token == "" {
			token = c.Param("token")
		}

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, router.ErrorResponse{
				Error: "Chave de API ou token não fornecido. Use Header Authorization: Bearer <token> ou ?apiKey=...",
			})
			return
		}

		userId, err := r.s.Authenticate(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, router.ErrorResponse{
				Error: "Chave de API ou token inválido.",
			})
			return
		}

		c.Set("userId", userId)
		c.Next()
	}
}

func (r *Router) AddRoutes() {
	authMw := r.IntegrationAuthMiddleware()

	// Direct routes: /api/integrations/...
	integrations := r.br.Router.Group("/integrations")
	{
		integrations.GET("/homepage", authMw, r.GetHomepage)
		integrations.GET("/ha/sensor", authMw, r.GetHomeAssistantSensor)
		integrations.GET("/ha/now-next.svg", authMw, r.GetNowNextSVG)
		integrations.GET("/ha/library", authMw, r.GetLibrary)
		integrations.GET("/ha/upcoming", authMw, r.GetUpcoming)
		integrations.GET("/calendar.ics", authMw, r.GetCalendarICS)
	}

	// Versioned v1 routes for Dashy / standard homelab format: /api/v1/integrations/...
	v1 := r.br.Router.Group("/v1/integrations")
	{
		v1.GET("/homepage", authMw, r.GetHomepage)
		v1.GET("/ha/sensor", authMw, r.GetHomeAssistantSensor)
		v1.GET("/ha/now-next.svg", authMw, r.GetNowNextSVG)
		v1.GET("/ha/library", authMw, r.GetLibrary)
		v1.GET("/ha/upcoming", authMw, r.GetUpcoming)
		v1.GET("/calendar.ics", authMw, r.GetCalendarICS)
	}

	// Legacy / user path calendar feed: /api/user/:token/calendar.ics
	r.br.Router.GET("/user/:token/calendar.ics", authMw, r.GetCalendarICS)

	// User API Key management: /api/user/api-key
	userGroup := r.br.Router.Group("/user").Use(authmiddleware.AuthRequired(r.br.Cfg))
	{
		userGroup.GET("/api-key", r.GetUserApiKey)
		userGroup.POST("/api-key/regenerate", r.RegenerateUserApiKey)
	}
}

func (r *Router) GetHomepage(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	data, err := r.s.GetHomepageData(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: "falha ao calcular dados do widget"})
		return
	}
	c.Header("Cache-Control", "public, max-age=60")
	c.JSON(http.StatusOK, data)
}

func (r *Router) GetHomeAssistantSensor(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	sensor, err := r.s.GetHomeAssistantSensor(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: "falha ao gerar sensor do Home Assistant"})
		return
	}
	c.Header("Cache-Control", "public, max-age=60")
	c.JSON(http.StatusOK, sensor)
}

func (r *Router) GetLibrary(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	lib, err := r.s.GetLibraryData(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: "falha ao carregar biblioteca"})
		return
	}
	c.Header("Cache-Control", "public, max-age=60")
	c.JSON(http.StatusOK, lib)
}

func (r *Router) GetUpcoming(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	upcoming, err := r.s.GetUpcomingData(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: "falha ao carregar estreias"})
		return
	}
	c.Header("Cache-Control", "public, max-age=60")
	c.JSON(http.StatusOK, gin.H{
		"data": upcoming,
	})
}

func (r *Router) GetCalendarICS(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	ics, err := r.s.GetCalendarICS(userId)
	if err != nil {
		c.String(http.StatusInternalServerError, "error generating calendar feed")
		return
	}
	c.Header("Content-Type", "text/calendar; charset=utf-8")
	c.Header("Content-Disposition", "inline; filename=\"watcharr.ics\"")
	c.Header("Cache-Control", "public, max-age=300")
	c.String(http.StatusOK, ics)
}

func (r *Router) GetNowNextSVG(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	svg, err := r.s.GetNowNextSVG(userId)
	if err != nil {
		c.String(http.StatusInternalServerError, "error generating svg card")
		return
	}
	c.Header("Content-Type", "image/svg+xml; charset=utf-8")
	c.Header("Cache-Control", "public, max-age=60")
	c.String(http.StatusOK, svg)
}

func (r *Router) GetUserApiKey(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	key, err := r.s.GetOrCreateApiKey(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: "falha ao obter chave de API"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"apiKey": key,
	})
}

func (r *Router) RegenerateUserApiKey(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	key, err := r.s.RegenerateApiKey(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: "falha ao regenerar chave de API"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"apiKey": key,
	})
}

