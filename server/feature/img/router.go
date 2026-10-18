// This router simply serves the images stored in the server data folder
// under the `img` folder.
// Note: The `img` folder contains user uploaded content (eg profile pictures).

package img

import (
	"net/http"
	"os"
	"path"

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
		localFile := path.Join(imgPath, fp)
		if _, err := os.Stat(localFile); err == nil {
			c.File(localFile)
			return
		}
		// Fallback to TMDB CDN so posters render immediately even while downloading
		c.Redirect(http.StatusTemporaryRedirect, "https://image.tmdb.org/t/p/w500"+fp)
	})
}
