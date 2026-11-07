// This router simply serves the images stored in the server data folder
// under the `img` folder.
// Note: The `img` folder contains user uploaded content (eg profile pictures).

package img

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/router"
)

type Router struct {
	br *router.BaseRouter
}

func NewRouter(br *router.BaseRouter) *Router {
	return &Router{
		br,
	}
}

func isValidTMDBImagePath(s string) bool {
	if len(s) <= 1 {
		return false
	}
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '/' || ch == '_' || ch == '.' || ch == '-') {
			return false
		}
	}
	return true
}

func (r *Router) AddRoutes() {
	imgPath := path.Join(config.DataPath, "img")
	img := r.br.Router.Group("/img").
		Use(func(c *gin.Context) {
			// The two following headers are preventative since this group
			// hosts content which can potentially include untrusted data.
			c.Header("X-Content-Type-Options", "nosniff")
			c.Header("Content-Security-Policy", "default-src 'none'; sandbox")
			c.Header("Cache-Control", "public, max-age=2592000, immutable")
			c.Next()
		})

	// Serve locally if exists, otherwise fallback redirect to TMDB CDN seamlessly
	img.GET("/*filepath", func(c *gin.Context) {
		fp := c.Param("filepath")
		if fp == "" || fp == "/" || strings.Contains(fp, "\\") || strings.Contains(fp, "..") || strings.ContainsRune(fp, 0) {
			c.Status(http.StatusBadRequest)
			return
		}
		absImgDir, err := filepath.Abs(imgPath)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		localFile := filepath.Join(absImgDir, filepath.FromSlash(fp))
		if !strings.HasPrefix(localFile, absImgDir+string(filepath.Separator)) {
			c.Status(http.StatusBadRequest)
			return
		}
		if info, err := os.Stat(localFile); err == nil && !info.IsDir() {
			c.File(localFile)
			return
		}
		if !strings.HasPrefix(fp, "/") || strings.HasPrefix(fp, "//") || !isValidTMDBImagePath(fp) {
			c.Status(http.StatusBadRequest)
			return
		}
		// Fallback to TMDB CDN so posters render immediately even while downloading
		c.Redirect(http.StatusTemporaryRedirect, "https://image.tmdb.org/t/p/w500"+fp)
	})
}
