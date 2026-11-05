package watched

import (
	"testing"
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/util"
	"github.com/stretchr/testify/assert"
)

func TestWatchedPageCache_VersionInvalidation(t *testing.T) {
	userId := uint(9999)
	v1 := getUserCacheVersion(userId)
	assert.Equal(t, int64(0), v1)

	// Simulate caching a page
	cacheKey1 := "wp:9999:v0:1:10:dummy:<nil>"

	dummyResp := util.PaginationResponse[entity.Watched, util.None]{
		TotalResults: 5,
	}
	watchedPageCache.Set(cacheKey1, dummyResp, 5*time.Minute)

	cached, found := watchedPageCache.Get(cacheKey1)
	assert.True(t, found)
	assert.Equal(t, int64(5), cached.(util.PaginationResponse[entity.Watched, util.None]).TotalResults)

	// Invalidate user cache
	InvalidateUserWatchedCache(userId)
	v2 := getUserCacheVersion(userId)
	assert.Equal(t, int64(1), v2)

	// New cache key has new version
	cacheKey2 := "wp:9999:v1:1:10:dummy:<nil>"
	_, foundNew := watchedPageCache.Get(cacheKey2)
	assert.False(t, foundNew, "New version cache key should be a miss")
}
