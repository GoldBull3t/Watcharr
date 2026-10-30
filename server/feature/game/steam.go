package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database/entity"
	"gorm.io/gorm"
)

type SteamService struct {
	db     *gorm.DB
	cfg    *config.ServerConfig
	client *http.Client
}

func NewSteamService(db *gorm.DB, cfg *config.ServerConfig) *SteamService {
	return &SteamService{
		db:  db,
		cfg: cfg,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Request and Response structs
type SteamConnectRequest struct {
	SteamID string `json:"steamId" binding:"required"`
	ApiKey  string `json:"apiKey"` // Optional personal key; falls back to server config / env
}

type SteamStatusResponse struct {
	Connected   bool   `json:"connected"`
	SteamID     string `json:"steamId,omitempty"`
	PersonaName string `json:"personaName,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
	TotalGames  int    `json:"totalGames"`
	HasCustomKey bool  `json:"hasCustomKey"`
}

type SteamSyncResponse struct {
	SyncedCount     int      `json:"syncedCount"`
	TotalSteamGames int      `json:"totalSteamGames"`
	UpdatedGames    []string `json:"updatedGames"`
	Message         string   `json:"message"`
}

type SteamOwnedGame struct {
	AppID                  int    `json:"appid"`
	Name                   string `json:"name"`
	PlaytimeForever        int    `json:"playtime_forever"` // In minutes
	Playtime2Weeks         int    `json:"playtime_2weeks"`  // In minutes
	ImgIconURL             string `json:"img_icon_url"`
	RtimeLastPlayed        int64  `json:"rtime_last_played"`
}

type steamGetOwnedGamesResponse struct {
	Response struct {
		GameCount int              `json:"game_count"`
		Games     []SteamOwnedGame `json:"games"`
	} `json:"response"`
}

type steamPlayerSummary struct {
	SteamID      string `json:"steamid"`
	PersonaName  string `json:"personaname"`
	ProfileURL   string `json:"profileurl"`
	Avatar       string `json:"avatar"`
	AvatarMedium string `json:"avatarmedium"`
	AvatarFull   string `json:"avatarfull"`
}

type steamGetPlayerSummariesResponse struct {
	Response struct {
		Players []steamPlayerSummary `json:"players"`
	} `json:"response"`
}

type steamResolveVanityURLResponse struct {
	Response struct {
		SteamID string `json:"steamid"`
		Success int    `json:"success"`
		Message string `json:"message"`
	} `json:"response"`
}

var steamPunctuationRegex = regexp.MustCompile(`[^a-z0-9]+`)

// NormalizeTitle strips special editions, trademarks, and punctuation for robust comparison.
func NormalizeTitle(t string) string {
	t = strings.ToLower(t)

	// Strip common edition/release markers
	markers := []string{
		"®", "™", "©",
		"game of the year edition", "goty edition", "goty",
		"definitive edition", "enhanced edition", "directors cut",
		"director's cut", "standard edition", "deluxe edition",
		"remastered", "anniversary edition", "complete edition",
		"digital deluxe edition",
	}
	for _, m := range markers {
		t = strings.ReplaceAll(t, m, "")
	}

	// Replace all non-alphanumeric chars with space
	t = steamPunctuationRegex.ReplaceAllString(t, " ")
	return strings.TrimSpace(t)
}

// TitlesMatch checks if two game names refer to the same title.
func TitlesMatch(a, b string) bool {
	normA := NormalizeTitle(a)
	normB := NormalizeTitle(b)

	if normA == "" || normB == "" {
		return false
	}
	if normA == normB {
		return true
	}

	// Substring / prefix check if string is sufficiently long (prevent short false matches)
	if len(normA) >= 5 && len(normB) >= 5 {
		var shorter, longer string
		if len(normA) < len(normB) {
			shorter, longer = normA, normB
		} else {
			shorter, longer = normB, normA
		}

		if strings.HasPrefix(longer, shorter) {
			remainder := strings.TrimSpace(strings.TrimPrefix(longer, shorter))
			// If remainder starts with a sequel indicator (number or roman numeral), they are NOT the same game!
			sequelIndicators := []string{
				"2", "3", "4", "5", "6", "7", "8", "9",
				"ii", "iii", "iv", "v", "vi", "vii", "viii", "ix", "x",
			}
			parts := strings.Fields(remainder)
			if len(parts) > 0 {
				firstExtra := parts[0]
				for _, seq := range sequelIndicators {
					if firstExtra == seq {
						return false
					}
				}
			}
			return true
		}
	}

	return false
}

// ResolveSteamID cleans up input (URL, vanity name or SteamID64).
func (s *SteamService) ResolveSteamID(input string, apiKey string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", errors.New("steam id or profile is required")
	}

	// Handle full profile URLs like https://steamcommunity.com/profiles/76561198... or /id/customname
	if strings.Contains(trimmed, "steamcommunity.com") {
		u, err := url.Parse(trimmed)
		if err == nil {
			parts := strings.Split(strings.Trim(u.Path, "/"), "/")
			if len(parts) >= 2 {
				trimmed = parts[1]
			}
		}
	}

	// 17-digit numeric SteamID64
	numericRegex := regexp.MustCompile(`^[0-9]{17}$`)
	if numericRegex.MatchString(trimmed) {
		return trimmed, nil
	}

	// Otherwise, treat as vanity URL name
	resolvedKey := s.resolveApiKey(apiKey)
	if resolvedKey == "" {
		return "", errors.New("steam api key is required to resolve vanity urls")
	}

	reqURL := fmt.Sprintf("https://api.steampowered.com/ISteamUser/ResolveVanityURL/v0001/?key=%s&vanityurl=%s",
		url.QueryEscape(resolvedKey), url.QueryEscape(trimmed))

	resp, err := s.client.Get(reqURL)
	if err != nil {
		return "", fmt.Errorf("failed contacting steam api: %w", err)
	}
	defer resp.Body.Close()

	var data steamResolveVanityURLResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", errors.New("failed decoding steam response")
	}

	if data.Response.Success != 1 || data.Response.SteamID == "" {
		return "", fmt.Errorf("steam user not found for '%s'", trimmed)
	}

	return data.Response.SteamID, nil
}

func (s *SteamService) resolveApiKey(userKey string) string {
	if strings.TrimSpace(userKey) != "" {
		return strings.TrimSpace(userKey)
	}
	return s.cfg.GetSteamKey()
}

// GetPlayerSummary fetches profile name and avatar from Steam.
func (s *SteamService) GetPlayerSummary(steamID string, apiKey string) (string, string, error) {
	key := s.resolveApiKey(apiKey)
	if key == "" {
		return "", "", nil
	}

	reqURL := fmt.Sprintf("https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s",
		url.QueryEscape(key), url.QueryEscape(steamID))

	resp, err := s.client.Get(reqURL)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var data steamGetPlayerSummariesResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", "", err
	}

	if len(data.Response.Players) > 0 {
		p := data.Response.Players[0]
		avatar := p.AvatarFull
		if avatar == "" {
			avatar = p.AvatarMedium
		}
		return p.PersonaName, avatar, nil
	}

	return "", "", nil
}

// GetOwnedGames fetches all games and playtimes for the given Steam user.
func (s *SteamService) GetOwnedGames(steamID string, apiKey string) ([]SteamOwnedGame, error) {
	key := s.resolveApiKey(apiKey)
	if key == "" {
		return nil, errors.New("no steam api key configured (set in Watcharr settings or server STEAM_API_KEY)")
	}

	reqURL := fmt.Sprintf("https://api.steampowered.com/IPlayerService/GetOwnedGames/v0001/?key=%s&steamid=%s&include_appinfo=1&include_played_free_games=1&format=json",
		url.QueryEscape(key), url.QueryEscape(steamID))

	resp, err := s.client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed contacting steam api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("steam api returned status %d: %s", resp.StatusCode, string(body))
	}

	var data steamGetOwnedGamesResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed decoding steam games response: %w", err)
	}

	return data.Response.Games, nil
}

// ConnectUser links the user's account with Steam.
func (s *SteamService) ConnectUser(userId uint, req SteamConnectRequest) (SteamStatusResponse, error) {
	steamID, err := s.ResolveSteamID(req.SteamID, req.ApiKey)
	if err != nil {
		return SteamStatusResponse{}, err
	}

	personaName, avatar, _ := s.GetPlayerSummary(steamID, req.ApiKey)

	// Fetch owned games count
	var gameCount int
	if games, err := s.GetOwnedGames(steamID, req.ApiKey); err == nil {
		gameCount = len(games)
	}

	// Save or update in user_services
	var svc entity.UserServices
	res := s.db.Where("user_id = ? AND name = ?", userId, entity.UserServiceNameSteam).First(&svc)
	if res.Error == nil {
		svc.ClientID = steamID
		svc.AuthToken = strings.TrimSpace(req.ApiKey)
		svc.AuthToken2 = personaName
		s.db.Save(&svc)
	} else {
		svc = entity.UserServices{
			UserID:     userId,
			Name:       entity.UserServiceNameSteam,
			ClientID:   steamID,
			AuthToken:  strings.TrimSpace(req.ApiKey),
			AuthToken2: personaName,
		}
		if err := s.db.Create(&svc).Error; err != nil {
			slog.Error("ConnectUser: Failed saving steam service", "error", err)
			return SteamStatusResponse{}, errors.New("failed to save steam integration")
		}
	}

	return SteamStatusResponse{
		Connected:    true,
		SteamID:      steamID,
		PersonaName:  personaName,
		Avatar:       avatar,
		TotalGames:   gameCount,
		HasCustomKey: req.ApiKey != "",
	}, nil
}

// GetStatus checks current user's Steam integration status.
func (s *SteamService) GetStatus(userId uint) (SteamStatusResponse, error) {
	var svc entity.UserServices
	res := s.db.Where("user_id = ? AND name = ?", userId, entity.UserServiceNameSteam).First(&svc)
	if res.Error != nil {
		return SteamStatusResponse{Connected: false}, nil
	}

	personaName, avatar, _ := s.GetPlayerSummary(svc.ClientID, svc.AuthToken)
	if personaName == "" && svc.AuthToken2 != "" {
		personaName = svc.AuthToken2
	}

	var gameCount int
	if games, err := s.GetOwnedGames(svc.ClientID, svc.AuthToken); err == nil {
		gameCount = len(games)
	}

	return SteamStatusResponse{
		Connected:    true,
		SteamID:      svc.ClientID,
		PersonaName:  personaName,
		Avatar:       avatar,
		TotalGames:   gameCount,
		HasCustomKey: svc.AuthToken != "",
	}, nil
}

// DisconnectUser removes user's Steam connection.
func (s *SteamService) DisconnectUser(userId uint) error {
	res := s.db.Where("user_id = ? AND name = ?", userId, entity.UserServiceNameSteam).Delete(&entity.UserServices{})
	return res.Error
}

// SyncUserPlaytime matches tracked games and updates hoursPlayed, platform, and progression.
func (s *SteamService) SyncUserPlaytime(userId uint) (SteamSyncResponse, error) {
	var svc entity.UserServices
	res := s.db.Where("user_id = ? AND name = ?", userId, entity.UserServiceNameSteam).First(&svc)
	if res.Error != nil {
		return SteamSyncResponse{}, errors.New("steam is not connected for this account")
	}

	games, err := s.GetOwnedGames(svc.ClientID, svc.AuthToken)
	if err != nil {
		return SteamSyncResponse{}, fmt.Errorf("failed fetching steam games: %w", err)
	}

	// Fetch all watched items for this user that are games
	var watchedList []entity.Watched
	err = s.db.
		Where("user_id = ? AND game_id IS NOT NULL", userId).
		Preload("Game").
		Find(&watchedList).Error
	if err != nil {
		return SteamSyncResponse{}, fmt.Errorf("failed fetching user games: %w", err)
	}

	if len(watchedList) == 0 {
		return SteamSyncResponse{
			SyncedCount:     0,
			TotalSteamGames: len(games),
			UpdatedGames:    []string{},
			Message:         "No games in Watcharr library to sync with Steam.",
		}, nil
	}

	syncedCount := 0
	var updatedTitles []string

	for _, w := range watchedList {
		if w.Game == nil || w.Game.Name == "" {
			continue
		}

		for _, sg := range games {
			if !TitlesMatch(w.Game.Name, sg.Name) {
				continue
			}

			// Hours from minutes
			hours := math.Round((float64(sg.PlaytimeForever)/60.0)*10) / 10
			needsSave := false

			// If Steam has recorded hours and it's higher than our current value (or current is nil)
			if hours > 0 && (w.HoursPlayed == nil || hours > *w.HoursPlayed) {
				w.HoursPlayed = &hours
				needsSave = true
			}

			// Assign platform if blank
			if w.Platform == "" {
				w.Platform = "PC / Steam"
				needsSave = true
			}

			// Promote to WATCHING if played in the last 2 weeks and currently PLANNED or HOLD
			if sg.Playtime2Weeks > 0 && (w.Status == entity.PLANNED || w.Status == entity.HOLD) {
				w.Status = entity.WATCHING
				needsSave = true
			}

			if needsSave {
				w.UpdatedAt = time.Now()
				if err := s.db.Save(&w).Error; err == nil {
					syncedCount++
					updatedTitles = append(updatedTitles, fmt.Sprintf("%s (%.1fh)", w.Game.Name, hours))
				}
			}
			break
		}
	}

	msg := fmt.Sprintf("Sincronizados %d jogos da Steam com sucesso.", syncedCount)
	if syncedCount == 0 {
		msg = "Nenhum jogo novo ou horas pendentes para atualizar."
	}

	return SteamSyncResponse{
		SyncedCount:     syncedCount,
		TotalSteamGames: len(games),
		UpdatedGames:    updatedTitles,
		Message:         msg,
	}, nil
}
