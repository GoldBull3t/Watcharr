package calendar

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/domain"
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

func (r *Router) AddRoutes() {
	calendar := r.br.Router.Group("/calendar").
		Use(authmiddleware.AuthRequired(r.br.Cfg))

	calendar.GET("", r.GetCalendarEvents)
}

func (r *Router) GetCalendarEvents(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	req := domain.CalendarRequest{
		Filter: domain.CalendarFilterLibrary,
		Type:   domain.CalendarTypeAll,
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "invalid query parameters"})
		return
	}

	events, err := r.s.GetEvents(userId, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: "failed to fetch calendar events"})
		return
	}

	c.JSON(http.StatusOK, events)
}

