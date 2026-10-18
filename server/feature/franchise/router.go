package franchise

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/feature/auth/authmiddleware"
	"github.com/sbondCo/Watcharr/feature/user/usermiddleware"
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

func (r *Router) AddRoutes() {
	franchise := r.br.Router.Group("/franchise")
	franchise.Use(authmiddleware.AuthRequired(r.br.Cfg))
	franchise.Use(usermiddleware.WithUser(r.br.DB))
	franchise.Use(router.LanguageRequired(r.br.Cfg))

	franchise.GET("", r.GetFranchises)
	franchise.GET("/:id", r.GetFranchiseRoadmap)
	franchise.GET("/check/:type/:id", r.FindFranchiseForContent)
}

func (r *Router) GetFranchises(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	summaries, err := r.s.GetFranchises(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, summaries)
}

func (r *Router) GetFranchiseRoadmap(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "id da franquia não fornecido"})
		return
	}
	order := c.DefaultQuery("order", "timeline")
	userLang := "en-US"
	if l, exists := c.Get("userLanguage"); exists {
		if s, ok := l.(string); ok && s != "" {
			userLang = s
		}
	}

	roadmap, err := r.s.GetFranchiseRoadmap(userId, id, order, userLang)
	if err != nil {
		c.JSON(http.StatusNotFound, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, roadmap)
}

func (r *Router) FindFranchiseForContent(c *gin.Context) {
	contentType := c.Param("type")
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "id inválido"})
		return
	}

	match, err := r.s.FindFranchiseForContent(contentType, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, match)
}

