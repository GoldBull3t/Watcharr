package addon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/router"
	"github.com/stretchr/testify/assert"
)

func TestGetManifest(t *testing.T) {
	s := &Service{}
	m := s.GetManifest()

	assert.Equal(t, "com.watcharr.nuvio.addon", m.ID)
	assert.Equal(t, "Watcharr", m.Name)
	assert.Equal(t, "1.3.0", m.Version)
	assert.Contains(t, m.Resources, "catalog")
	assert.Contains(t, m.Resources, "meta")
	assert.Contains(t, m.Resources, "stream")
	assert.Contains(t, m.Types, "movie")
	assert.Contains(t, m.Types, "series")
	assert.Equal(t, 4, len(m.Catalogs))
	assert.NotNil(t, m.BehaviorHints)

	// Ensure each catalog strictly contains required Stremio fields
	for _, cat := range m.Catalogs {
		assert.NotEmpty(t, cat.Type)
		assert.NotEmpty(t, cat.ID)
		assert.NotEmpty(t, cat.Name)
		assert.NotEmpty(t, cat.Extra)
		assert.Contains(t, cat.ExtraSupported, "skip")
	}
}

func TestRouterRoutesRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()

	br := &router.BaseRouter{
		Router: &engine.RouterGroup,
		Cfg:    &config.ServerConfig{},
	}

	s := &Service{}
	r := NewRouter(br, s)

	// Must not panic on route registration
	assert.NotPanics(t, func() {
		r.AddRoutes()
	})

	// Test manifest endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/addon/mytoken/manifest.json", nil)
	engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var manifest ManifestResponse
	err := json.Unmarshal(w.Body.Bytes(), &manifest)
	assert.NoError(t, err)
	assert.Equal(t, "Watcharr", manifest.Name)
	assert.Equal(t, "1.3.0", manifest.Version)
	assert.Equal(t, 4, len(manifest.Catalogs))
	assert.Contains(t, manifest.Resources, "stream")
}

func TestParseStreamID(t *testing.T) {
	tmdbID, imdbID, sNum, epNum := ParseStreamID("tt0903747:1:2")
	assert.Equal(t, "tt0903747", imdbID)
	assert.Equal(t, 1, sNum)
	assert.Equal(t, 2, epNum)
	assert.Equal(t, 0, tmdbID)

	tmdbID, imdbID, sNum, epNum = ParseStreamID("tmdb:1396:2:5.json")
	assert.Equal(t, 1396, tmdbID)
	assert.Equal(t, 2, sNum)
	assert.Equal(t, 5, epNum)
	assert.Equal(t, "", imdbID)

	tmdbID, imdbID, sNum, epNum = ParseStreamID("tmdb:550")
	assert.Equal(t, 550, tmdbID)
	assert.Equal(t, 0, sNum)
	assert.Equal(t, 0, epNum)

	tmdbID, imdbID, sNum, epNum = ParseStreamID("tt0137523")
	assert.Equal(t, "tt0137523", imdbID)
	assert.Equal(t, 0, sNum)
	assert.Equal(t, 0, epNum)
}
