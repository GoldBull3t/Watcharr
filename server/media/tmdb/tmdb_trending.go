package tmdb

import (
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/sbondCo/Watcharr/cache"
)

func (t *TMDB) Trending(
	ttype TrendingType,
	pageNum int,
	region string,
	language ...string,
) (TrendingCombined, error) {
	resp := new(TrendingCombined)
	if ttype != TrendingTypeAll &&
		ttype != TrendingTypeMovie &&
		ttype != TrendingTypeShow &&
		ttype != TrendingTypePerson {
		slog.Error("Trending: Invalid type provided", "provided_t", t)
		return *resp, errors.New("invalid type")
	}
	if pageNum <= 0 {
		pageNum = 1
	}
	lang := ""
	if len(language) > 0 {
		lang = language[0]
	}
	cacheKey := cache.CreateCacheKey(
		"Trending",
		string(ttype),
		region,
		pageNum,
		lang)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("Trending: Returning cache.")
		return *resp, nil
	}
	p := map[string]string{
		"page":   strconv.Itoa(pageNum),
		"region": region,
	}
	if lang != "" {
		p["language"] = lang
	}
	err := t.req(
		"/trending/"+string(ttype)+"/day",
		p,
		&resp)
	if err != nil {
		slog.Error("Trending: Request failed!", "error", err)
		return *resp, errors.New("request failed")
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}
