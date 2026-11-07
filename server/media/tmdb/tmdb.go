package tmdb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	gocache "github.com/robfig/go-cache"
	"github.com/sbondCo/Watcharr/database/entity"
)

type TMDBError struct {
	StatusCode int    `json:"status_code"`
	HTTPStatus int    `json:"http_status"`
	Message    string `json:"status_message"`
}

func (e *TMDBError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("TMDB error (status %d)", e.HTTPStatus)
}

var ContentStore = gocache.New(time.Hour*24, time.Minute)

// TokenBucket provides a thread-safe token bucket rate limiter.
type TokenBucket struct {
	capacity   float64
	tokens     float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

func NewTokenBucket(capacity float64, refillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) Wait() {
	if tb == nil {
		return
	}
	for {
		tb.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(tb.lastRefill).Seconds()
		tb.tokens += elapsed * tb.refillRate
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now

		if tb.tokens >= 1.0 {
			tb.tokens -= 1.0
			tb.mu.Unlock()
			return
		}

		missing := 1.0 - tb.tokens
		waitTime := time.Duration((missing / tb.refillRate) * float64(time.Second))
		tb.mu.Unlock()

		if waitTime > 0 {
			time.Sleep(waitTime)
		}
	}
}

type ContentProvider interface {
	CacheContentShow(content ShowDetails, onlyUpdate bool) (entity.Content, error)
	CacheContentMovie(content MovieDetails, onlyUpdate bool) (entity.Content, error)
}

type TMDB struct {
	Key             string
	DefaultLanguage func() string
	contentProvider ContentProvider
	client          *http.Client
	limiter         *TokenBucket
}

func NewTMDB(key string) *TMDB {
	return &TMDB{
		Key: key,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		limiter: NewTokenBucket(40, 4.0), // 40 max burst, 4 tokens/sec refill rate
	}
}

func (t *TMDB) SetClient(c *http.Client) {
	t.client = c
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

	// Wait for token from rate limiter
	if t.limiter != nil {
		t.limiter.Wait()
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
	// If custom key failed with 401 Unauthorized, retry with default Watcharr key
	defaultKey := "d047fa61d926371f277e7a83c9c4ff2c"
	if res.StatusCode == http.StatusUnauthorized && t.GetKey() != defaultKey {
		slog.Warn("TMDB returned 401 with custom TMDB_KEY; retrying with default Watcharr key")
		params.Set("api_key", defaultKey)
		base.RawQuery = params.Encode()
		if t.limiter != nil {
			t.limiter.Wait()
		}
		resRetry, errRetry := client.Get(base.String())
		if errRetry == nil {
			defer resRetry.Body.Close()
			if resRetry.StatusCode == 200 {
				bodyRetry, errRetryBody := io.ReadAll(resRetry.Body)
				if errRetryBody == nil {
					return bodyRetry, nil
				}
			} else {
				res = resRetry
				if b, errB := io.ReadAll(resRetry.Body); errB == nil {
					body = b
				}
			}
		}
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
		if t.limiter != nil {
			t.limiter.Wait()
		}
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
		var tmdbErr TMDBError
		if jsonErr := json.Unmarshal(body, &tmdbErr); jsonErr == nil && tmdbErr.Message != "" {
			tmdbErr.HTTPStatus = res.StatusCode
			return nil, &tmdbErr
		}
		return nil, &TMDBError{
			HTTPStatus: res.StatusCode,
			Message:    string(body),
		}
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

// InvalidateCacheForID evicts cached TMDB entries so fresh details and seasons are fetched (#1049).
func (t *TMDB) InvalidateCacheForID(id string) {
	if id == "" || ContentStore == nil {
		return
	}
	var keysToDelete []string
	ContentStore.Lock()
	cv := reflect.ValueOf(ContentStore)
	if cv.Kind() == reflect.Ptr && !cv.IsNil() {
		inner := cv.Elem().Field(0)
		if inner.Kind() == reflect.Ptr && !inner.IsNil() {
			items := inner.Elem().FieldByName("items")
			if items.IsValid() && items.Kind() == reflect.Map {
				for _, kv := range items.MapKeys() {
					k := kv.String()
					if strings.Contains(k, id) {
						keysToDelete = append(keysToDelete, k)
					}
				}
			}
		}
	}
	ContentStore.Unlock()
	for _, k := range keysToDelete {
		ContentStore.Delete(k)
	}
}
