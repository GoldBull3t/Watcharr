// Package tvdb is a minimal TheTVDB API v4 client used to provide
// TV show metadata when TVDB is selected as the metadata provider (#555, #1103).
package tvdb

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"
)

const baseURL = "https://api4.thetvdb.com/v4"

// Tokens are valid for a month; refresh a bit earlier.
const tokenTTL = 27 * 24 * time.Hour

var httpClient = &http.Client{Timeout: 15 * time.Second}

type Client struct {
	mu       sync.Mutex
	apiKey   string
	token    string
	tokenExp time.Time
}

// Translation holds localized text for a series.
type Translation struct {
	Name     string `json:"name"`
	Overview string `json:"overview"`
}

type apiResponse[T any] struct {
	Status string `json:"status"`
	Data   T      `json:"data"`
}

// New returns a client for the given API key.
func New(apiKey string) *Client {
	return &Client{apiKey: strings.TrimSpace(apiKey)}
}

// langToISO3 maps the app's locale codes (eg. pt-BR) to the 3 letter codes TVDB uses.
var langToISO3 = map[string]string{
	"en": "eng", "pt": "por", "es": "spa", "fr": "fra", "de": "deu", "it": "ita",
	"ja": "jpn", "ko": "kor", "zh": "zho", "ru": "rus", "tr": "tur", "nl": "nld",
	"pl": "pol", "sv": "swe", "da": "dan", "fi": "fin", "no": "nor", "cs": "ces",
	"hu": "hun", "ar": "ara", "he": "heb", "el": "ell", "uk": "ukr", "ro": "ron",
}

// ISO3 converts a locale like "pt-BR" to "por". Falls back to "eng".
func ISO3(locale string) string {
	base := strings.ToLower(strings.SplitN(locale, "-", 2)[0])
	if v, ok := langToISO3[base]; ok {
		return v
	}
	return "eng"
}

func (c *Client) login() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token != "" && time.Now().Before(c.tokenExp) {
		return c.token, nil
	}
	if c.apiKey == "" {
		return "", errors.New("tvdb api key not configured")
	}
	body, _ := json.Marshal(map[string]string{"apikey": c.apiKey})
	resp, err := httpClient.Post(baseURL+"/login", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("tvdb login request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("tvdb login failed with status %d (check your TVDB API key)", resp.StatusCode)
	}
	var out apiResponse[struct {
		Token string `json:"token"`
	}]
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil || out.Data.Token == "" {
		return "", errors.New("tvdb login returned an invalid response")
	}
	c.token = out.Data.Token
	c.tokenExp = time.Now().Add(tokenTTL)
	return c.token, nil
}

func get[T any](c *Client, path string) (T, error) {
	var zero T
	tok, err := c.login()
	if err != nil {
		return zero, err
	}
	req, err := http.NewRequest(http.MethodGet, baseURL+path, nil)
	if err != nil {
		return zero, err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		c.mu.Lock()
		c.token = ""
		c.mu.Unlock()
		return zero, errors.New("tvdb authorization failed")
	}
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		return zero, fmt.Errorf("tvdb returned status %d for %s", resp.StatusCode, path)
	}
	var out apiResponse[T]
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return zero, fmt.Errorf("tvdb response decode failed: %w", err)
	}
	return out.Data, nil
}

// Validate checks that the API key can authenticate.
func (c *Client) Validate() error {
	_, err := c.login()
	return err
}

// SeriesTranslation returns the localized name/overview for a series.
func (c *Client) SeriesTranslation(tvdbID int, locale string) (Translation, error) {
	t, err := get[Translation](c, fmt.Sprintf("/series/%d/translations/%s", tvdbID, ISO3(locale)))
	if err != nil {
		slog.Debug("tvdb: SeriesTranslation failed", "tvdb_id", tvdbID, "error", err)
		return Translation{}, err
	}
	return t, nil
}

