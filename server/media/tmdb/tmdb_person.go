package tmdb

import (
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/sbondCo/Watcharr/cache"
)

func (t *TMDB) PersonDetails(id string, optParams ...map[string]string) (PersonDetails, error) {
	params := map[string]string{}
	if len(optParams) > 0 && optParams[0] != nil {
		for k, v := range optParams[0] {
			params[k] = v
		}
	}
	cacheKey := cache.CreateCacheKey("PersonDetails", id, params)
	resp := new(PersonDetails)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("PersonDetails: Returning cache.")
		return *resp, nil
	}
	err := t.req("/person/"+id, params, &resp)
	if err != nil {
		slog.Error("PersonDetails: Request failed!", "error", err)
		return PersonDetails{}, errors.New("request failed")
	}
	reqLang := t.GetDefaultLanguage()
	if l, ok := params["language"]; ok && l != "" {
		reqLang = l
	}
	if reqLang != "en-US" && resp.Biography == "" {
		fbParams := map[string]string{"language": "en-US"}
		fallbackResp := new(PersonDetails)
		if err := t.req("/person/"+id, fbParams, &fallbackResp); err == nil {
			if resp.Biography == "" {
				resp.Biography = fallbackResp.Biography
			}
		}
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

func (t *TMDB) PersonCredits(id string, optParams ...map[string]string) (PersonCombinedCredits, error) {
	params := map[string]string{}
	if len(optParams) > 0 && optParams[0] != nil {
		for k, v := range optParams[0] {
			params[k] = v
		}
	}
	cacheKey := cache.CreateCacheKey("PersonCredits", id, params)
	resp := new(PersonCombinedCredits)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("PersonCredits: Returning cache.")
		return *resp, nil
	}
	err := t.req(
		"/person/"+id+"/combined_credits",
		params,
		&resp)
	if err != nil {
		slog.Error("PersonCredits: Request failed!", "error", err)
		return PersonCombinedCredits{}, errors.New("request failed")
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}

func (t *TMDB) PopularPeople(pageNum int, language ...string) (PopularPeople, error) {
	params := map[string]string{"page": strconv.Itoa(pageNum)}
	if len(language) > 0 && language[0] != "" {
		params["language"] = language[0]
	}
	cacheKey := cache.CreateCacheKey("PopularPeople", pageNum, params)
	resp := new(PopularPeople)
	if cache.GetCache(ContentStore, cacheKey, &resp) {
		slog.Debug("PopularPeople: Returning cache.")
		return *resp, nil
	}
	err := t.req(
		"/person/popular",
		params,
		&resp)
	if err != nil {
		slog.Error("PopularPeople: Request failed!", "error", err)
		return PopularPeople{}, errors.New("request failed")
	}
	ContentStore.Set(cacheKey, resp, time.Hour*24)
	return *resp, nil
}
