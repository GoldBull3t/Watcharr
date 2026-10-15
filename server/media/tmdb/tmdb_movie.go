package tmdb

import (
	"errors"
	"log/slog"
	"time"

	"github.com/sbondCo/Watcharr/cache"
)

type MovieDetailsOptions struct {
	// TMDB ID
	ID string
	// Country (currently used for watch providers)
	Country string
	// Request params map.
	Params map[string]string

	// If CacheContentMovie should be ran or not.
	// If the caller wants to do its own caching to the db, it can use this
	// to avoid multiple calls to CacheContentMovie.
	DontRunDBCache bool
}

func (t *TMDB) MovieDetails(o MovieDetailsOptions) (MovieDetails, error) {
	resp := new(MovieDetails)
	cacheKey := cache.CreateCacheKey(
		"MovieDetails",
		o.ID,
		o.Country,
		o.Params)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("MovieDetails: Returning cache.")
		return *resp, nil
	}
	err := t.req("/movie/"+o.ID, o.Params, &resp)
	if err != nil {
		slog.Error("MovieDetails: Request failed!", "error", err)
		return MovieDetails{}, errors.New("request failed")
	}
	reqLang := t.GetDefaultLanguage()
	if l, ok := o.Params["language"]; ok && l != "" {
		reqLang = l
	}
	if reqLang != "en-US" && resp.Overview == "" {
		fbParams := make(map[string]string)
		for k, v := range o.Params {
			fbParams[k] = v
		}
		fbParams["language"] = "en-US"
		fallbackResp := new(MovieDetails)
		if err := t.req("/movie/"+o.ID, fbParams, &fallbackResp); err == nil {
			if resp.Overview == "" {
				resp.Overview = fallbackResp.Overview
			}
			if resp.Tagline == "" {
				resp.Tagline = fallbackResp.Tagline
			}
			if resp.Title == "" {
				resp.Title = fallbackResp.Title
			}
		}
	}
	resp.WatchProvidersTransformed = transformProviders(
		&resp.WatchProviders,
		o.Country)
	// We don't want this to linger around (in cache) since we have the
	// transformed version now..
	resp.WatchProviders = nil
	if !o.DontRunDBCache {
		go t.contentProvider.CacheContentMovie(*resp, true)
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

func (t *TMDB) MovieCredits(id string, optParams ...map[string]string) (ContentCredits, error) {
	params := map[string]string{}
	if len(optParams) > 0 && optParams[0] != nil {
		for k, v := range optParams[0] {
			params[k] = v
		}
	}
	resp := new(ContentCredits)
	err := t.req("/movie/"+id+"/credits", params, &resp)
	if err != nil {
		slog.Error("MovieCredits: Request failed!", "error", err)
		return ContentCredits{}, errors.New("request failed")
	}
	return *resp, nil
}

func (t *TMDB) Collection(id string, optParams ...map[string]string) (TMDBCollection, error) {
	params := map[string]string{}
	if len(optParams) > 0 && optParams[0] != nil {
		for k, v := range optParams[0] {
			params[k] = v
		}
	}
	resp := new(TMDBCollection)
	cacheKey := cache.CreateCacheKey("Collection", id, params)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("Collection: Returning cache.")
		return *resp, nil
	}
	err := t.req("/collection/"+id, params, &resp)
	if err != nil {
		slog.Error("Collection: Request failed!", "error", err)
		return TMDBCollection{}, errors.New("request failed")
	}
	reqLang := t.GetDefaultLanguage()
	if l, ok := params["language"]; ok && l != "" {
		reqLang = l
	}
	if reqLang != "en-US" && (resp.Overview == "" || resp.Name == "") {
		fbParams := make(map[string]string)
		for k, v := range params {
			fbParams[k] = v
		}
		fbParams["language"] = "en-US"
		fallbackResp := new(TMDBCollection)
		if err := t.req("/collection/"+id, fbParams, &fallbackResp); err == nil {
			if resp.Overview == "" {
				resp.Overview = fallbackResp.Overview
			}
			if resp.Name == "" {
				resp.Name = fallbackResp.Name
			}
		}
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}
