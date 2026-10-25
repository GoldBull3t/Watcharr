package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
	"gorm.io/gorm"
)

var activityHttpClient = &http.Client{
	Timeout: 10 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
	Transport: &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil || len(ips) == 0 {
				return nil, fmt.Errorf("failed to resolve webhook host: %w", err)
			}
			allowLocal := os.Getenv("ALLOW_LOCAL_WEBHOOKS") == "true"
			var safeIP net.IP
			for _, ipAddr := range ips {
				ip := ipAddr.IP
				if isBlockedWebhookIP(ip, allowLocal) {
					return nil, errors.New("webhook destination IP is restricted for security (SSRF protection)")
				}
				if safeIP == nil {
					safeIP = ip
				}
			}
			dialer := &net.Dialer{Timeout: 8 * time.Second}
			return dialer.DialContext(ctx, network, net.JoinHostPort(safeIP.String(), port))
		},
	},
}

func isBlockedWebhookIP(ip net.IP, allowLocal bool) bool {
	if ip == nil {
		return true
	}
	if ip.IsUnspecified() || ip.IsMulticast() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	if !allowLocal && (ip.IsLoopback() || ip.IsPrivate()) {
		return true
	}
	return false
}

func validateWebhookURL(rawUrl string) (string, error) {
	trimmed := strings.TrimSpace(rawUrl)
	if trimmed == "" {
		return "", errors.New("webhook URL cannot be empty")
	}
	parsed, err := url.ParseRequestURI(trimmed)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", errors.New("invalid webhook URL scheme (must be http or https)")
	}
	if parsed.Host == "" {
		return "", errors.New("invalid webhook URL host")
	}
	return trimmed, nil
}

type ActivityWebhookPayload struct {
	Event       string  `json:"event"`
	Status      string  `json:"status,omitempty"`
	MediaType   string  `json:"mediaType"`
	Title       string  `json:"title"`
	Season      int     `json:"season,omitempty"`
	Episode     int     `json:"episode,omitempty"`
	EpisodeName string  `json:"episodeName,omitempty"`
	Rating      float64 `json:"rating,omitempty"`
	Poster      string  `json:"poster,omitempty"`
	Backdrop    string  `json:"backdrop,omitempty"`
	Overview    string  `json:"overview,omitempty"`
	User        string  `json:"user"`
	Timestamp   string  `json:"timestamp"`
	Reason      string  `json:"reason,omitempty"`
}

// DispatchActivityWebhook asynchronously sends a real-time event webhook
func DispatchActivityWebhook(db *gorm.DB, userId uint, payload ActivityWebhookPayload) {
	go func() {
		var user entity.User
		if err := db.Where("id = ?", userId).First(&user).Error; err != nil {
			return
		}

		if user.ActivityWebhookEnabled == nil || !*user.ActivityWebhookEnabled {
			return
		}
		if user.ActivityWebhookUrl == nil || strings.TrimSpace(*user.ActivityWebhookUrl) == "" {
			return
		}

		targetUrl := strings.TrimSpace(*user.ActivityWebhookUrl)

		// Filter check
		if user.ActivityWebhookFilter != nil && *user.ActivityWebhookFilter != "" && *user.ActivityWebhookFilter != "all" {
			filter := strings.ToLower(strings.TrimSpace(*user.ActivityWebhookFilter))
			switch filter {
			case "finished":
				if payload.Status != "FINISHED" {
					return
				}
			case "watching":
				if payload.Status != "WATCHING" {
					return
				}
			case "rating":
				if payload.Rating <= 0 {
					return
				}
			}
		}

		if payload.User == "" {
			payload.User = user.Username
		}
		if payload.Timestamp == "" {
			payload.Timestamp = time.Now().UTC().Format(time.RFC3339)
		}

		if err := SendActivityWebhook(targetUrl, payload, false); err != nil {
			slog.Warn("DispatchActivityWebhook: Failed to send", "userId", userId, "error", err)
		} else {
			slog.Info("DispatchActivityWebhook: Sent event successfully", "userId", userId, "event", payload.Event, "title", payload.Title)
		}
	}()
}

// SendActivityWebhook handles formatting and HTTP delivery
func SendActivityWebhook(rawUrl string, payload ActivityWebhookPayload, isTest bool) error {
	u, err := validateWebhookURL(rawUrl)
	if err != nil {
		return err
	}

	isDiscord := strings.Contains(u, "discord.com/api/webhooks") || strings.Contains(u, "discordapp.com/api/webhooks")
	isSlack := strings.Contains(u, "hooks.slack.com")

	if isDiscord {
		return sendDiscordActivityWebhook(u, payload, isTest)
	} else if isSlack {
		return sendSlackActivityWebhook(u, payload, isTest)
	}

	return sendGenericActivityWebhook(u, payload, isTest)
}

func sendDiscordActivityWebhook(webhookUrl string, p ActivityWebhookPayload, isTest bool) error {
	color := 0x3b82f6 // Blue
	headline := fmt.Sprintf("▶️ **%s** começou a ver **%s**", p.User, p.Title)

	if isTest {
		color = 0x6366f1
		headline = "🔔 **Watcharr: Notificação de Teste de Webhook em Tempo Real!**"
	} else if p.Rating > 0 {
		color = 0xfbbf24 // Gold
		headline = fmt.Sprintf("⭐ **%s** avaliou **%s** com %.1f★", p.User, p.Title, p.Rating)
	} else if p.Status == "FINISHED" {
		color = 0x10b981 // Emerald
		if p.Season > 0 && p.Episode > 0 {
			headline = fmt.Sprintf("✅ **%s** terminou o episódio S%02dE%02d de **%s**", p.User, p.Season, p.Episode, p.Title)
		} else {
			headline = fmt.Sprintf("🏆 **%s** terminou de ver **%s**!", p.User, p.Title)
		}
	} else if p.Status == "WATCHING" {
		if p.Season > 0 && p.Episode > 0 {
			headline = fmt.Sprintf("🍿 **%s** está a ver S%02dE%02d de **%s**", p.User, p.Season, p.Episode, p.Title)
		}
	}

	var fields []map[string]any
	if p.Season > 0 && p.Episode > 0 {
		epDesc := fmt.Sprintf("S%02dE%02d", p.Season, p.Episode)
		if p.EpisodeName != "" {
			epDesc += ": " + p.EpisodeName
		}
		fields = append(fields, map[string]any{
			"name":   "Episódio",
			"value":  epDesc,
			"inline": true,
		})
	}
	if p.Status != "" {
		fields = append(fields, map[string]any{
			"name":   "Status",
			"value":  p.Status,
			"inline": true,
		})
	}
	if p.Rating > 0 {
		fields = append(fields, map[string]any{
			"name":   "Avaliação",
			"value":  fmt.Sprintf("%.1f / 10", p.Rating),
			"inline": true,
		})
	}

	embed := map[string]any{
		"title":       p.Title,
		"description": p.Overview,
		"color":       color,
		"fields":      fields,
		"footer": map[string]any{
			"text": "Watcharr • Integrações Homelab",
		},
		"timestamp": p.Timestamp,
	}

	if p.Poster != "" {
		embed["thumbnail"] = map[string]any{
			"url": p.Poster,
		}
	}

	body := map[string]any{
		"content": headline,
		"embeds":  []any{embed},
	}

	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", webhookUrl, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := activityHttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("discord returned status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

func sendSlackActivityWebhook(webhookUrl string, p ActivityWebhookPayload, isTest bool) error {
	msg := fmt.Sprintf("Watcharr: %s atualizou %s (Status: %s)", p.User, p.Title, p.Status)
	body := map[string]any{
		"text": msg,
	}
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", webhookUrl, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := activityHttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// sendGenericActivityWebhook sends pure structured JSON (ideal for Home Assistant & Ntfy)
func sendGenericActivityWebhook(webhookUrl string, p ActivityWebhookPayload, isTest bool) error {
	jsonBytes, err := json.Marshal(p)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", webhookUrl, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Watcharr-Outbound-Webhook/4.2.3")
	if isTest {
		req.Header.Set("X-Watcharr-Test", "true")
	}

	resp, err := activityHttpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook endpoint returned status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

// TestActivityWebhook sends a test notification
func TestActivityWebhook(db *gorm.DB, userId uint, targetUrl string) error {
	var user entity.User
	_ = db.Where("id = ?", userId).First(&user)

	sample := ActivityWebhookPayload{
		Event:     "media.finished",
		Status:    "FINISHED",
		MediaType: "movie",
		Title:     "Avatar: Fire and Ash",
		Rating:    5.0,
		Poster:    "https://image.tmdb.org/t/p/w500/fQpJFVWX8MCdOkwoutDw7tjhvgp.jpg",
		Overview:  "Teste de conectividade de Webhook em Tempo Real do Watcharr para Home Assistant e Homelab.",
		User:      user.Username,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Reason:    "Test Webhook",
	}

	return SendActivityWebhook(targetUrl, sample, true)
}

