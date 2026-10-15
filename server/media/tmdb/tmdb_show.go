package tmdb

import (
	"errors"
	"log/slog"
	"time"

	"github.com/sbondCo/Watcharr/cache"
)

type ShowDetailsOptions struct {
	// TMDB ID
	ID string
	// Country (currently used for watch providers)
	Country string
	// Request params map.
	Params map[string]string

	// If CacheContentShow should be ran or not.
	// If the caller wants to do its own caching to the db, it can use this
	// to avoid multiple calls to CacheContentShow.
	DontRunDBCache bool
}

func (t *TMDB) ShowDetails(o ShowDetailsOptions) (ShowDetails, error) {
	cacheKey := cache.CreateCacheKey(
		"ShowDetails",
		o.ID,
		o.Country,
		o.Params)
	resp := new(ShowDetails)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("ShowDetails: Returning cache.")
		return *resp, nil
	}
	err := t.req("/tv/"+o.ID, o.Params, &resp)
	if err != nil {
		slog.Error("ShowDetails: Request failed!", "error", err)
		return ShowDetails{}, errors.New("request failed")
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
		fallbackResp := new(ShowDetails)
		if err := t.req("/tv/"+o.ID, fbParams, &fallbackResp); err == nil {
			if resp.Overview == "" {
				resp.Overview = fallbackResp.Overview
			}
			if resp.Tagline == "" {
				resp.Tagline = fallbackResp.Tagline
			}
			if resp.Name == "" {
				resp.Name = fallbackResp.Name
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
		go t.contentProvider.CacheContentShow(*resp, true)
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

func (t *TMDB) ShowCredits(id string, optParams ...map[string]string) (ContentCredits, error) {
	params := map[string]string{}
	if len(optParams) > 0 && optParams[0] != nil {
		for k, v := range optParams[0] {
			params[k] = v
		}
	}
	resp := new(ContentCredits)
	err := t.req("/tv/"+id+"/credits", params, &resp)
	if err != nil {
		slog.Error("ShowCredits: Request failed!", "error", err)
		return ContentCredits{}, errors.New("request failed")
	}
	return *resp, nil
}

func (t *TMDB) SeasonDetails(
	showId string,
	seasonNumber string,
	optParams ...map[string]string,
) (SeasonDetails, error) {
	params := map[string]string{}
	if len(optParams) > 0 && optParams[0] != nil {
		for k, v := range optParams[0] {
			params[k] = v
		}
	}
	cacheKey := cache.CreateCacheKey("SeasonDetails", showId, seasonNumber, params)
	resp := new(SeasonDetails)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("SeasonDetails: Returning cache.")
		return *resp, nil
	}
	err := t.req(
		"/tv/"+showId+"/season/"+seasonNumber,
		params,
		&resp)
	if err != nil {
		slog.Error("SeasonDetails: Request failed!", "error", err)
		return SeasonDetails{}, errors.New("request failed")
	}
	reqLang := t.GetDefaultLanguage()
	if l, ok := params["language"]; ok && l != "" {
		reqLang = l
	}
	if reqLang != "en-US" {
		needFallback := resp.Overview == ""
		if !needFallback {
			for _, ep := range resp.Episodes {
				if ep.Overview == "" {
					needFallback = true
					break
				}
			}
		}
		if needFallback {
			fbParams := map[string]string{"language": "en-US"}
			fallbackResp := new(SeasonDetails)
			if err := t.req("/tv/"+showId+"/season/"+seasonNumber, fbParams, &fallbackResp); err == nil {
				if resp.Overview == "" {
					resp.Overview = fallbackResp.Overview
				}
				if resp.Name == "" {
					resp.Name = fallbackResp.Name
				}
				for i := range resp.Episodes {
					if i < len(fallbackResp.Episodes) {
						if resp.Episodes[i].Overview == "" {
							resp.Episodes[i].Overview = fallbackResp.Episodes[i].Overview
						}
						if resp.Episodes[i].Name == "" {
							resp.Episodes[i].Name = fallbackResp.Episodes[i].Name
						}
					}
				}
			}
		}
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}
