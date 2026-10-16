package calendar

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

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
)

var httpClient = &http.Client{
	Timeout: 15 * time.Second,
}

// Discord Webhook Payload
type discordEmbedField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type discordEmbedThumbnail struct {
	URL string `json:"url"`
}

type discordEmbedFooter struct {
	Text string `json:"text"`
}

type discordEmbed struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	URL         string                 `json:"url,omitempty"`
	Color       int                    `json:"color"`
	Fields      []discordEmbedField    `json:"fields,omitempty"`
	Thumbnail   *discordEmbedThumbnail `json:"thumbnail,omitempty"`
	Footer      *discordEmbedFooter    `json:"footer,omitempty"`
	Timestamp   string                 `json:"timestamp,omitempty"`
}

type discordWebhookPayload struct {
	Username  string         `json:"username"`
	AvatarURL string         `json:"avatar_url,omitempty"`
	Content   string         `json:"content,omitempty"`
	Embeds    []discordEmbed `json:"embeds"`
}

// Slack Webhook Payload
type slackWebhookPayload struct {
	Text string `json:"text"`
}

// Generic Webhook Payload
type genericWebhookPayload struct {
	Event    string                `json:"event"`
	Date     string                `json:"date"`
	IsTest   bool                  `json:"isTest"`
	Count    int                   `json:"count"`
	Releases []domain.CalendarEvent `json:"releases"`
}

// TestWebhook sends a test notification to either a provided URL or the user's configured webhook.
func (s *Service) TestWebhook(userId uint, customUrl string) (string, error) {
	targetUrl := strings.TrimSpace(customUrl)
	if targetUrl == "" {
		var user entity.User
		if err := s.db.Where("id = ?", userId).Take(&user).Error; err != nil {
			return "", errors.New("user not found")
		}
		if user.CalendarWebhookUrl == nil || strings.TrimSpace(*user.CalendarWebhookUrl) == "" {
			return "", errors.New("no webhook URL configured")
		}
		targetUrl = strings.TrimSpace(*user.CalendarWebhookUrl)
	}

	today := time.Now().Format("2006-01-02")
	sampleEvents := []domain.CalendarEvent{
		{
			ID:            "test-tv-1",
			MediaID:       100088,
			MediaType:     "tv",
			Title:         "Watcharr - Teste de Notificação",
			ReleaseDate:   today,
			Overview:      "O teu webhook de notificações do calendário do Watcharr está configurado e pronto a alertar sobre as tuas estreias!",
			InLibrary:     true,
			Status:        "WATCHING",
			SeasonNumber:  1,
			EpisodeNumber: 1,
			EpisodeName:   "Conexão Estabelecida com Sucesso",
		},
	}

	if err := s.SendWebhook(targetUrl, sampleEvents, true); err != nil {
		return "", err
	}

	return "Notificação de teste enviada com sucesso!", nil
}

// CheckAndNotifyUser checks releases for today and sends webhook notifications if applicable.
func (s *Service) CheckAndNotifyUser(userId uint, forceSend bool) (int, error) {
	var user entity.User
	if err := s.db.Where("id = ?", userId).Take(&user).Error; err != nil {
		return 0, errors.New("user not found")
	}

	if user.CalendarWebhookUrl == nil || strings.TrimSpace(*user.CalendarWebhookUrl) == "" {
		return 0, errors.New("webhook URL is not configured")
	}

	targetUrl := strings.TrimSpace(*user.CalendarWebhookUrl)
	if !forceSend {
		if user.CalendarWebhookEnabled == nil || !*user.CalendarWebhookEnabled {
			return 0, nil
		}
	}

	today := time.Now().Format("2006-01-02")
	cacheKey := fmt.Sprintf("CalendarNotified_%d_%s", user.ID, today)

	if !forceSend {
		if _, found := CalendarStore.Get(cacheKey); found {
			slog.Debug("Calendar: User already notified today", "userId", user.ID, "date", today)
			return 0, nil
		}
	}

	filterType := domain.CalendarTypeAll
	if user.CalendarWebhookFilter != nil && *user.CalendarWebhookFilter != "" {
		filterType = domain.CalendarType(*user.CalendarWebhookFilter)
	}

	events, err := s.GetEvents(user.ID, domain.CalendarRequest{
		Start:  today,
		End:    today,
		Filter: domain.CalendarFilterLibrary,
		Type:   filterType,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to fetch today's calendar events: %w", err)
	}

	if len(events) == 0 {
		return 0, nil
	}

	if err := s.SendWebhook(targetUrl, events, false); err != nil {
		return 0, err
	}

	CalendarStore.Set(cacheKey, true, 24*time.Hour)
	return len(events), nil
}

// CheckAndNotifyAllUsers runs for all users who have calendar notifications enabled.
func (s *Service) CheckAndNotifyAllUsers() error {
	var users []entity.User
	err := s.db.Where("calendar_webhook_enabled = ? AND calendar_webhook_url != ?", true, "").Find(&users).Error
	if err != nil {
		slog.Error("Calendar: Failed to query users for release notifications", "error", err)
		return err
	}

	slog.Info("Calendar: Checking release notifications for users", "count", len(users))
	for _, u := range users {
		count, err := s.CheckAndNotifyUser(u.ID, false)
		if err != nil {
			slog.Error("Calendar: Failed to notify user of releases", "userId", u.ID, "error", err)
		} else if count > 0 {
			slog.Info("Calendar: Sent release notification to user", "userId", u.ID, "releases", count)
		}
	}

	return nil
}

// SendWebhook dispatches formatted payloads to the webhook URL.
func (s *Service) SendWebhook(webhookUrl string, events []domain.CalendarEvent, isTest bool) error {
	u := strings.TrimSpace(webhookUrl)
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		return errors.New("webhook URL must start with http:// or https://")
	}

	isDiscord := strings.Contains(u, "discord.com/api/webhooks") || strings.Contains(u, "discordapp.com/api/webhooks")
	isSlack := strings.Contains(u, "hooks.slack.com")

	if isDiscord {
		return s.sendDiscordWebhook(u, events, isTest)
	} else if isSlack {
		return s.sendSlackWebhook(u, events, isTest)
	}

	return s.sendGenericWebhook(u, events, isTest)
}

func (s *Service) sendDiscordWebhook(url string, events []domain.CalendarEvent, isTest bool) error {
	headline := "🍿 **Estreias de Hoje na Tua Lista do Watcharr!**"
	if isTest {
		headline = "🔔 **Notificação de Teste do Watcharr Calendar**"
	}

	// Discord accepts up to 10 embeds per message. Batch in chunks of 10 if necessary.
	for i := 0; i < len(events); i += 10 {
		end := i + 10
		if end > len(events) {
			end = len(events)
		}
		chunk := events[i:end]

		var embeds []discordEmbed
		for _, ev := range chunk {
			title := ev.Title
			if ev.MediaType == "tv" && ev.SeasonNumber > 0 {
				epName := ev.EpisodeName
				if epName == "" {
					epName = fmt.Sprintf("Episódio %d", ev.EpisodeNumber)
				}
				title = fmt.Sprintf("%s — S%02dE%02d: %s", ev.Title, ev.SeasonNumber, ev.EpisodeNumber, epName)
			} else if ev.MediaType == "movie" {
				title = fmt.Sprintf("%s (%s)", ev.Title, ev.ReleaseDate)
			}

			desc := ev.Overview
			if len(desc) > 300 {
				desc = desc[:297] + "..."
			}

			fields := []discordEmbedField{
				{
					Name:   "Tipo",
					Value:  strings.ToUpper(ev.MediaType),
					Inline: true,
				},
				{
					Name:   "Data de Estreia",
					Value:  ev.ReleaseDate,
					Inline: true,
				},
			}
			if ev.Status != "" {
				fields = append(fields, discordEmbedField{
					Name:   "Na tua lista",
					Value:  ev.Status,
					Inline: true,
				})
			}

			emb := discordEmbed{
				Title:       title,
				Description: desc,
				Color:       3586712, // Hex #36BA98
				Fields:      fields,
				Footer: &discordEmbedFooter{
					Text: "Watcharr Release Radar",
				},
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			}

			if ev.PosterPath != "" {
				emb.Thumbnail = &discordEmbedThumbnail{
					URL: fmt.Sprintf("https://image.tmdb.org/t/p/w500%s", ev.PosterPath),
				}
			}

			embeds = append(embeds, emb)
		}

		payload := discordWebhookPayload{
			Username:  "Watcharr",
			AvatarURL: "https://raw.githubusercontent.com/sbondCo/Watcharr/main/ui/static/favicon.png",
			Content:   headline,
			Embeds:    embeds,
		}

		body, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to encode discord payload: %w", err)
		}

		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("discord request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			respBody, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("discord webhook returned %d: %s", resp.StatusCode, string(respBody))
		}
	}

	return nil
}

func (s *Service) sendSlackWebhook(url string, events []domain.CalendarEvent, isTest bool) error {
	var sb strings.Builder
	if isTest {
		sb.WriteString("🔔 *Notificação de Teste do Watcharr Calendar*\nO teu webhook está configurado e funcional!\n")
	} else {
		sb.WriteString(fmt.Sprintf("🍿 *Estreias de Hoje no Watcharr (%s)*\n\n", time.Now().Format("2006-01-02")))
		for _, ev := range events {
			if ev.MediaType == "tv" {
				sb.WriteString(fmt.Sprintf("• *%s* (S%02dE%02d) - %s\n", ev.Title, ev.SeasonNumber, ev.EpisodeNumber, ev.EpisodeName))
			} else {
				sb.WriteString(fmt.Sprintf("• *%s* (Filme)\n", ev.Title))
			}
		}
	}

	payload := slackWebhookPayload{
		Text: sb.String(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("slack request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("slack webhook returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (s *Service) sendGenericWebhook(url string, events []domain.CalendarEvent, isTest bool) error {
	payload := genericWebhookPayload{
		Event:    "calendar.releases",
		Date:     time.Now().Format("2006-01-02"),
		IsTest:   isTest,
		Count:    len(events),
		Releases: events,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Watcharr-Calendar-Webhook/1.0")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

