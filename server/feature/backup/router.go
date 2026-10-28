package backup

import (
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
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
	g := r.br.Router.Group("/backup").
		Use(
			authmiddleware.AuthRequired(r.br.Cfg),
			usermiddleware.WithUser(r.br.DB),
			authmiddleware.AdminRequired(),
		)

	// List all backups and get config
	g.GET("/", r.GetBackups)
	g.GET("/config", r.GetConfig)
	g.POST("/config", r.UpdateConfig)

	// Create backup
	g.POST("/", r.CreateBackup)

	// Download backup
	g.GET("/download/:filename", r.DownloadBackup)

	// Delete backup
	g.DELETE("/:filename", r.DeleteBackup)

	// Restore backup
	g.POST("/restore/:filename", r.RestoreBackup)

	// Upload backup archive
	g.POST("/upload", r.UploadBackup)
}

func (r *Router) GetBackups(c *gin.Context) {
	backups, err := r.s.ListBackups()
	if err != nil {
		slog.Error("GetBackups: Failed to list backups", "error", err)
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: "failed to list backups"})
		return
	}
	cfg := r.s.GetConfig()
	c.JSON(http.StatusOK, gin.H{
		"backups": backups,
		"config":  cfg,
	})
}

func (r *Router) GetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, r.s.GetConfig())
}

func (r *Router) UpdateConfig(c *gin.Context) {
	var req UpdateBackupConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "invalid config payload"})
		return
	}
	if err := r.s.UpdateConfig(req); err != nil {
		slog.Error("UpdateConfig: Failed to update backup config", "error", err)
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: "failed to update backup config"})
		return
	}
	c.JSON(http.StatusOK, r.s.GetConfig())
}

func (r *Router) CreateBackup(c *gin.Context) {
	var req CreateBackupRequest
	_ = c.ShouldBindJSON(&req)

	info, err := r.s.CreateBackup(false, req.Format)
	if err != nil {
		slog.Error("CreateBackup: Failed to create backup", "error", err)
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

func (r *Router) DownloadBackup(c *gin.Context) {
	filename := c.Param("filename")
	path, err := r.s.DownloadBackup(filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.FileAttachment(path, filepath.Base(path))
}

func (r *Router) DeleteBackup(c *gin.Context) {
	filename := c.Param("filename")
	if err := r.s.DeleteBackup(filename); err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (r *Router) RestoreBackup(c *gin.Context) {
	filename := c.Param("filename")
	if err := r.s.RestoreBackup(filename); err != nil {
		slog.Error("RestoreBackup: Failed to restore backup", "file", filename, "error", err)
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"restored": true})
}

func (r *Router) UploadBackup(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: "no file uploaded"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, router.ErrorResponse{Error: "failed to open uploaded file"})
		return
	}
	defer src.Close()

	info, err := r.s.SaveUploadedBackup(src, file.Filename)
	if err != nil {
		c.JSON(http.StatusBadRequest, router.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}

