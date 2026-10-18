package tmdb

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	gocache "github.com/robfig/go-cache"
	"github.com/sbondCo/Watcharr/database/entity"
)

var ContentStore = gocache.New(time.Hour*24, time.Minute)

type ContentProvider interface {
	CacheContentShow(content ShowDetails, onlyUpdate bool) (entity.Content, error)
	CacheContentMovie(content MovieDetails, onlyUpdate bool) (entity.Content, error)
}

type TMDB struct {
	Key             string
	DefaultLanguage func() string
	contentProvider ContentProvider
	client          *http.Client
}

func NewTMDB(key string) *TMDB {
	return &TMDB{
		Key: key,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (t *TMDB) AddContentProvider(contentProvider ContentProvider) {
	t.contentProvider = contentProvider
}

func (t *TMDB) GetDefaultLanguage() string {
	if t.DefaultLanguage != nil {
		if l := t.DefaultLanguage(); l != "" {
			return l
		}
	}
	return "en-US"
}

func (t *TMDB) GetKey() string {
	if t.Key != "" {
		return t.Key //Config.TMDB_KEY
	}
	return "d047fa61d926371f277e7a83c9c4ff2c"
}

func (t *TMDB) apiRequest(ep string, p map[string]string) ([]byte, error) {
	slog.Debug("tmdbAPIRequest", "endpoint", ep, "params", p)
	base, err := url.Parse("https://api.themoviedb.org/3")
	if err != nil {
		return nil, errors.New("failed to parse api uri")
	}

	// Path params
	base.Path += ep

	// Query params
	params := url.Values{}
	params.Add("api_key", t.GetKey())
	lang := t.GetDefaultLanguage()
	if customLang, ok := p["language"]; ok && customLang != "" {
		lang = customLang
	}
	params.Add("language", lang)
	for k, v := range p {
		if k == "language" {
			continue
		}
		params.Add(k, v)
	}

	// Add params to url
	base.RawQuery = params.Encode()

	// Run get request
	client := t.client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	res, err := client.Get(base.String())
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}
	// Check for rate limit (429) and retry once after backing off
	if res.StatusCode == http.StatusTooManyRequests {
		retryAfter := 1 * time.Second
		if ra := res.Header.Get("Retry-After"); ra != "" {
			if s, err := strconv.Atoi(ra); err == nil && s > 0 {
				retryAfter = time.Duration(s) * time.Second
			}
		}
		slog.Warn("TMDB rate limit hit (429), retrying after delay", "retryAfter", retryAfter)
		time.Sleep(retryAfter)

		// Retry request
		resRetry, errRetry := client.Get(base.String())
		if errRetry == nil {
			defer resRetry.Body.Close()
			if resRetry.StatusCode == 200 {
				bodyRetry, errRetryBody := io.ReadAll(resRetry.Body)
				if errRetryBody == nil {
					return bodyRetry, nil
				}
			}
		}
	}

	if res.StatusCode != 200 {
		slog.Error("TMDB non 200 status code:", "status_code", res.StatusCode)
		return nil, errors.New(string(body))
	}
	return body, nil
}

func (t *TMDB) req(ep string, p map[string]string, resp interface{}) error {
	body, err := t.apiRequest(ep, p)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(body), &resp)
	if err != nil {
		return err
	}
	return nil
}

func (t *TMDB) PrimaryTranslations() ([]string, error) {
	var translations []string
	err := t.req("/configuration/primary_translations", map[string]string{}, &translations)
	if err != nil {
		return nil, err
	}
	return translations, nil
}
