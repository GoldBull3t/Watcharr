package watched

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestAutoFinishCompletedMovies(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.AutoMigrate(&entity.Content{}, &entity.Watched{}, &entity.Activity{})
	assert.NoError(t, err)

	// Create Movie 1: 90 min runtime, started 120 min ago (Should be auto-finished)
	c1 := entity.Content{
		TmdbID:  101,
		Title:   "Expired Movie",
		Type:    entity.MOVIE,
		Runtime: 90,
	}
	db.Create(&c1)

	twoHoursAgo := time.Now().UTC().Add(-120 * time.Minute)
	w1 := entity.Watched{
		UserID:    1,
		ContentID: &c1.ID,
		Status:    entity.WATCHING,
	}
	db.Create(&w1)
	db.Model(&w1).Update("updated_at", twoHoursAgo)

	// Create Movie 2: 120 min runtime, started 30 min ago (Should NOT be auto-finished)
	c2 := entity.Content{
		TmdbID:  102,
		Title:   "Playing Movie",
		Type:    entity.MOVIE,
		Runtime: 120,
	}
	db.Create(&c2)

	thirtyMinAgo := time.Now().UTC().Add(-30 * time.Minute)
	w2 := entity.Watched{
		UserID:    1,
		ContentID: &c2.ID,
		Status:    entity.WATCHING,
	}
	db.Create(&w2)
	db.Model(&w2).Update("updated_at", thirtyMinAgo)

	// Run auto finish
	count, err := AutoFinishCompletedMovies(db)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Verify w1 transitioned to FINISHED
	var checked1 entity.Watched
	db.First(&checked1, w1.ID)
	assert.Equal(t, entity.FINISHED, checked1.Status)

	// Verify w2 remains WATCHING
	var checked2 entity.Watched
	db.First(&checked2, w2.ID)
	assert.Equal(t, entity.WATCHING, checked2.Status)
}

