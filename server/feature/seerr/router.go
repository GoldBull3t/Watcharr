package seerr

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/feature/auth/authmiddleware"
	"github.com/sbondCo/Watcharr/feature/user/usermiddleware"
	"github.com/sbondCo/Watcharr/router"
	"github.com/sbondCo/Watcharr/seerr"
)

type Router struct {
	br *router.BaseRouter
}

func NewRouter(br *router.BaseRouter) *Router {
	return &Router{
		br: br,
	}
}

func (r *Router) getClient() (*seerr.Client, bool) {
	if !r.br.Cfg.OverseerrEnabled() {
		return nil, false
	}
	return seerr.NewClient(r.br.Cfg.OVERSEERR_HOST, r.br.Cfg.OVERSEERR_KEY), true
}

func (r *Router) AddRoutes() {
	g := r.br.Router.Group("/seerr").
		Use(
			authmiddleware.AuthRequired(r.br.Cfg),
			usermiddleware.WithUser(r.br.DB),
		)

	// Admin test endpoint (can test arbitrary host/key or currently saved config)
	g.POST("/test", authmiddleware.AdminRequired(), r.TestConnection)

	// Status check for media item
	g.GET("/status/:type/:tmdbId", authmiddleware.PermRequired(entity.PERM_REQUEST_CONTENT), r.GetStatus)

	// Submit request to Overseerr / Jellyseerr
	g.POST("/request", authmiddleware.PermRequired(entity.PERM_REQUEST_CONTENT), r.CreateRequest)
}

func (r *Router) TestConnection(c *gin.Context) {
	var tr seerr.TestRequest
	_ = c.ShouldBindJSON(&tr)

	host := tr.Host
	key := tr.Key

	if host == "" {
		host = r.br.Cfg.OVERSEERR_HOST
	}
	if key == "" {
		key = r.br.Cfg.OVERSEERR_KEY
	}

	if host == "" || key == "" {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "host and api key are required"})
		return
	}

	client := seerr.NewClient(host, key)
	status, err := client.Test()
	if err != nil {
		slog.Error("TestConnection failed for Seerr", "host", host, "error", err)
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, status)
}

func (r *Router) GetStatus(c *gin.Context) {
	mediaType := c.Param("type")
	tmdbIdStr := c.Param("tmdbId")
	tmdbId, err := strconv.Atoi(tmdbIdStr)
	if err != nil || tmdbId <= 0 {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "invalid tmdbId"})
		return
	}

	client, configured := r.getClient()
	if !configured {
		c.JSON(http.StatusOK, seerr.ContentStatusResponse{
			Configured: false,
			Status:     seerr.MediaStatusUnknown,
			StatusName: "UNKNOWN",
		})
		return
	}

	media, err := client.GetMediaStatus(mediaType, tmdbId)
	if err != nil {
		slog.Error("Seerr GetMediaStatus error", "type", mediaType, "tmdbId", tmdbId, "error", err)
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: "failed to fetch status from Overseerr"})
		return
	}

	resp := seerr.ContentStatusResponse{
		Configured:   true,
		Status:       media.Status,
		StatusName:   media.Status.String(),
		IsAvailable:  media.Status == seerr.MediaStatusAvailable,
		IsProcessing: media.Status == seerr.MediaStatusProcessing,
		IsPending:    media.Status == seerr.MediaStatusPendingApproval,
		MediaInfo:    media,
	}

	c.JSON(http.StatusOK, resp)
}

func (r *Router) CreateRequest(c *gin.Context) {
	client, configured := r.getClient()
	if !configured {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "Jellyseerr/Overseerr is not configured on this server"})
		return
	}

	var payload seerr.RequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "invalid request payload"})
		return
	}

	if payload.MediaID <= 0 || (payload.MediaType != "movie" && payload.MediaType != "tv") {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "mediaId and valid mediaType ('movie' or 'tv') are required"})
		return
	}

	if payload.MediaType == "tv" && payload.Seasons == nil {
		payload.Seasons = "all"
	}

	resp, err := client.CreateRequest(payload)
	if err != nil {
		slog.Error("Seerr CreateRequest error", "payload", payload, "error", err)
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}
