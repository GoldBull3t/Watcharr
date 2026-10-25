package seerr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

var httpClient = &http.Client{
	Timeout: 15 * time.Second,
}

type Client struct {
	Host string
	Key  string
}

func NewClient(host, key string) *Client {
	return &Client{
		Host: strings.TrimRight(strings.TrimSpace(host), "/"),
		Key:  strings.TrimSpace(key),
	}
}

func (c *Client) request(method, path string, body any, target any) (int, error) {
	if c.Host == "" {
		return 0, errors.New("overseerr host not configured")
	}

	url := fmt.Sprintf("%s%s", c.Host, path)

	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return 0, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBytes)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return 0, fmt.Errorf("failed to create http request: %w", err)
	}

	req.Header.Set("X-Api-Key", c.Key)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	if target != nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
		}
		if len(respBytes) > 0 {
			if err := json.Unmarshal(respBytes, target); err != nil {
				return resp.StatusCode, fmt.Errorf("failed to unmarshal response: %w", err)
			}
		}
	}

	return resp.StatusCode, nil
}

// Test checks connectivity and authentication with Jellyseerr/Overseerr.
func (c *Client) Test() (*StatusResponse, error) {
	var resp StatusResponse
	code, err := c.request(http.MethodGet, "/api/v1/status", nil, &resp)
	if err != nil {
		return nil, err
	}
	if code == http.StatusUnauthorized || code == http.StatusForbidden {
		return nil, errors.New("unauthorized: invalid api key")
	}
	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("status check failed with code %d", code)
	}
	return &resp, nil
}

// GetMediaStatus queries media status by TMDB ID.
func (c *Client) GetMediaStatus(mediaType string, tmdbId int) (*MediaInfo, error) {
	endpointType := "movie"
	if mediaType == "tv" || mediaType == "show" {
		endpointType = "tv"
	}

	path := fmt.Sprintf("/api/v1/%s/%d", endpointType, tmdbId)

	var resp MediaResponse
	code, err := c.request(http.MethodGet, path, nil, &resp)
	if err != nil {
		slog.Error("GetMediaStatus request failed", "type", mediaType, "tmdbId", tmdbId, "error", err)
		return nil, err
	}

	if code == http.StatusNotFound {
		// Not found in Overseerr database means it hasn't been requested or indexed yet
		return &MediaInfo{
			TmdbID: tmdbId,
			Status: MediaStatusUnknown,
		}, nil
	}

	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("seerr returned status code %d", code)
	}

	if resp.MediaInfo == nil {
		return &MediaInfo{
			TmdbID: tmdbId,
			Status: MediaStatusUnknown,
		}, nil
	}

	return resp.MediaInfo, nil
}

// CreateRequest submits a request to Overseerr.
func (c *Client) CreateRequest(payload RequestPayload) (*RequestResponse, error) {
	var resp RequestResponse
	code, err := c.request(http.MethodPost, "/api/v1/request", payload, &resp)
	if err != nil {
		slog.Error("CreateRequest failed", "payload", payload, "error", err)
		return nil, err
	}

	if code == http.StatusConflict {
		return nil, errors.New("media has already been requested or is available")
	}

	if code < 200 || code >= 300 {
		return nil, fmt.Errorf("seerr request returned status code %d", code)
	}

	return &resp, nil
}

