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
	assert.Equal(t, "1.1.0", m.Version)
	assert.Contains(t, m.Resources, "catalog")
	assert.Contains(t, m.Resources, "meta")
	assert.Contains(t, m.Types, "movie")
	assert.Contains(t, m.Types, "series")
	assert.True(t, len(m.Catalogs) >= 4)
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
	assert.Equal(t, "1.1.0", manifest.Version)
	assert.True(t, len(manifest.Catalogs) > 0)
}


