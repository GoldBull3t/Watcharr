package addon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetManifest(t *testing.T) {
	s := &Service{}
	m := s.GetManifest()

	assert.Equal(t, "com.watcharr.nuvio.addon", m.ID)
	assert.Equal(t, "Watcharr", m.Name)
	assert.Contains(t, m.Resources, "catalog")
	assert.Contains(t, m.Resources, "meta")
	assert.Contains(t, m.Types, "movie")
	assert.Contains(t, m.Types, "series")
	assert.True(t, len(m.Catalogs) >= 4)
}

