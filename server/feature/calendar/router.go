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
	calendar.POST("/webhook/test", r.TestCalendarWebhook)
	calendar.POST("/webhook/send", r.SendCalendarWebhook)
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

type WebhookTestReq struct {
	Url string `json:"url"`
}

func (r *Router) TestCalendarWebhook(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	var body WebhookTestReq
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "invalid request body"})
			return
		}
	}

	msg, err := r.s.TestWebhook(userId, body.Url)
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": msg,
	})
}

func (r *Router) SendCalendarWebhook(c *gin.Context) {
	userId := c.MustGet("userId").(uint)
	count, err := r.s.CheckAndNotifyUser(userId, true)
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}

	msg := "Nenhuma estreia encontrada na tua biblioteca para o dia de hoje."
	if count > 0 {
		msg = "Notificação de lançamentos de hoje enviada com sucesso!"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"count":   count,
		"message": msg,
	})
}
