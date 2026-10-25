package push

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/feature/auth/authmiddleware"
	"github.com/sbondCo/Watcharr/feature/user/usermiddleware"
	"github.com/sbondCo/Watcharr/router"
)

type Router struct {
	br *router.BaseRouter
	s  *Service
}

func NewRouter(br *router.BaseRouter, s *Service) *Router {
	return &Router{
		br: br,
		s:  s,
	}
}

func (r *Router) AddRoutes() {
	g := r.br.Router.Group("/push").
		Use(
			authmiddleware.AuthRequired(r.br.Cfg),
			usermiddleware.WithUser(r.br.DB),
		)

	// Get VAPID Public Key needed by browser to subscribe
	g.GET("/key", r.GetKey)

	// Subscribe device to push notifications
	g.POST("/subscribe", r.Subscribe)

	// Unsubscribe device from push notifications
	g.DELETE("/subscribe", r.Unsubscribe)

	// Send a test push notification to the logged-in user
	g.POST("/test", r.SendTest)
}

func (r *Router) GetKey(c *gin.Context) {
	pub, err := r.s.GetPublicKey()
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: "failed to get public key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"publicKey": pub,
	})
}

func (r *Router) Subscribe(c *gin.Context) {
	user := c.MustGet("user").(entity.User)

	var req SubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "invalid subscription payload"})
		return
	}

	sub, err := r.s.Subscribe(user.ID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: "failed to save subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"id":      sub.ID,
	})
}

func (r *Router) Unsubscribe(c *gin.Context) {
	user := c.MustGet("user").(entity.User)

	var req UnsubscribeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "invalid unsubscribe payload"})
		return
	}

	if err := r.s.Unsubscribe(user.ID, req.Endpoint); err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: "failed to unsubscribe"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

func (r *Router) SendTest(c *gin.Context) {
	user := c.MustGet("user").(entity.User)

	sent, err := r.s.SendTest(user.ID)
	if err != nil {
		slog.Error("Push: Test notification error", "user_id", user.ID, "error", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"sent":    sent,
	})
}

