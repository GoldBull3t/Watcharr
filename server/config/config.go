package config

import (
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"os"
	"path"
	"strings"
	"sync"
	"time"
	"uuid"

	"github.com/sbondCo/Watcharr/config/cfgmodel"
	"github.com/sbondCo/Watcharr/logging"
	"github.com/sbondCo/Watcharr/media/igdb"
	"github.com/sbondCo/Watcharr/util"
)

var DataPath = func() string {
	path := os.Getenv("WATCHARR_DATA")
	if path == "" {
		path = "./data"
	}
	return path
}()

type TrustedHeaderAuthSetting struct {
	// Required: Should header auth be enabled?
	// This bool exists so header auth can be toggled
	// easily without having to remove configuration.
	// To be actually enabled, HEADER_NAME must also
	// be set.
	Enabled bool `json:"enabled"`
	// Required: What is the name of the trusted header
	// that will contain the logged in users username?
	HeaderName string `json:"headerName"`
	// Should the frontend attempt auto login if
	// trusted header auth is enabled.
	AutoLogin bool `json:"autoLogin"`
	// Where can we redirect the user to logout
	// of the auth service?
	LogoutUrl string `json:"logoutUrl"`
}

// OIDCSetting holds OpenID Connect / OAuth2 single sign-on configuration (#461).
type OIDCSetting struct {
	Enabled      bool   `json:"enabled"`
	ProviderName string `json:"providerName"`
	IssuerURL    string `json:"issuerUrl"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	RedirectURI  string `json:"redirectUri"`
	AutoRegister bool   `json:"autoRegister"`
}

var cfgMu sync.RWMutex

type ServerConfig struct {
	// Used to sign JWT tokens. Make sure to make
	// it strong, just like a very long, complicated password.
	JWT_SECRET string `json:",omitempty"`

	// Default country for new users. This is used to set the default
	// region to get correct content streaming providers.
	// TODO Enforce iso_3166_1 validity (same as tmdb)
	DEFAULT_COUNTRY string `json:",omitempty"`

	// Default language for TMDB metadata and new users (e.g. en-US, de-DE, tr-TR, pt-BR).
	DEFAULT_LANGUAGE string `json:",omitempty"`

	// Default permissions bitmask for newly registered users (#674).
	DEFAULT_USER_PERMISSIONS int `json:",omitempty"`

	// Optional: Point to your Jellyfin install
	// to enable it as an auth provider.
	JELLYFIN_HOST string `json:",omitempty"`

	// Optional: Enable jellyfin webhook endpoint.
	JELLYFIN_WEBHOOK_ENABLED bool `json:",omitempty"`

	// Optional: Jellyfin webhook endpoint auth key (UUID).
	// Auto generated for the user when the webhook is enabled.
	JELLYFIN_WEBHOOK_KEY uuid.UUID `json:",omitzero"`

	// Optional: Use Emby instead of Jellyfin branding in the ui.
	USE_EMBY bool

	// Enable/disable signup functionality.
	// Set to `false` to disable registering an account.
	SIGNUP_ENABLED bool

	// Optional: Provide your own TMDB API Key.
	// If unprovided, the default Watcharr API key will be used.
	TMDB_KEY string `json:",omitempty"`

	// Optional: Provide your own TheTVDB API Key (v4) (#555, #1103).
	TVDB_KEY string `json:",omitempty"`

	// Primary TV metadata provider ("tmdb" or "tvdb") (#555, #1103).
	METADATA_PROVIDER string `json:",omitempty"`

	// Optional: OMDb API Key for IMDb, Rotten Tomatoes, Metacritic & Awards.
	OMDB_KEY string `json:",omitempty"`

	// Optional: MDBList API Key for Letterboxd, Trakt, Rotten Tomatoes, IMDb & Metacritic scores.
	MDBLIST_KEY string `json:",omitempty"`

	// Optional: Steam Web API Key for syncing gaming playtime and stats.
	STEAM_KEY string `json:",omitempty"`

	// Optional: Point to Plex install to enable plex features.
	PLEX_HOST string `json:",omitempty"`

	// Optional: Machine identifier of your Plex server.
	// This is used to ensure only users of your Plex server
	// can use this Watcharr instance.
	// Will be fetched automatically when PLEX_HOST is provided via web ui.
	PLEX_MACHINE_ID string `json:",omitempty"`

	// Optional: Trusted header authentication configuration.
	// VERY DANGEROUS if access is not controlled correctly!
	HEADER_AUTH TrustedHeaderAuthSetting `json:",omitempty"`

	// Optional: OpenID Connect / OAuth2 authentication configuration (#461).
	OIDC OIDCSetting `json:",omitempty"`

	// Optional: Point to Jellyseerr or Overseerr install to enable media requests.
	OVERSEERR_HOST string `json:",omitempty"`
	OVERSEERR_KEY  string `json:",omitempty"`

	SONARR []cfgmodel.SonarrSettings `json:",omitempty"`
	RADARR []cfgmodel.RadarrSettings `json:",omitempty"`
	TWITCH igdb.IGDB                 `json:",omitzero"`

	// Optional: Schedule for tasks.
	TASK_SCHEDULE map[string]int `json:",omitempty"`

	// Automated disaster recovery backup configuration
	BACKUP_ENABLED   *bool `json:",omitempty"`
	BACKUP_RETENTION int   `json:",omitempty"`

	// Enable/disable debug logging. Useful for when trying
	// to figure out exactly what the server is doing at a point
	// of failure.
	// Set to `true` to enable.
	DEBUG bool `json:",omitempty"`
}

// ServerConfig, but with JWT_SECRET removed from json.
// Used for returning to user from get config api request.
//
// Technically only admins will have access to that api route,
// but I feel more comfortable removing it anyways (+ this is
// not editable on frontend, so not needed).
func (c *ServerConfig) GetSafe() ServerConfig {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	return ServerConfig{
		SIGNUP_ENABLED:           c.SIGNUP_ENABLED,
		DEFAULT_COUNTRY:          c.DEFAULT_COUNTRY,
		DEFAULT_LANGUAGE:         c.DEFAULT_LANGUAGE,
		DEFAULT_USER_PERMISSIONS: c.GetDefaultUserPermissions(),
		JELLYFIN_HOST:            c.JELLYFIN_HOST,
		JELLYFIN_WEBHOOK_ENABLED: c.JELLYFIN_WEBHOOK_ENABLED,
		JELLYFIN_WEBHOOK_KEY:     c.JELLYFIN_WEBHOOK_KEY,
		USE_EMBY:                 c.USE_EMBY,
		TMDB_KEY:                 c.TMDB_KEY,
		TVDB_KEY:                 c.TVDB_KEY,
		METADATA_PROVIDER:        c.GetMetadataProvider(),
		OMDB_KEY:                 c.OMDB_KEY,
		MDBLIST_KEY:              c.MDBLIST_KEY,
		STEAM_KEY:                c.STEAM_KEY,
		PLEX_HOST:                c.PLEX_HOST,
		PLEX_MACHINE_ID:          c.PLEX_MACHINE_ID,
		OVERSEERR_HOST:           c.OVERSEERR_HOST,
		OVERSEERR_KEY:            c.OVERSEERR_KEY,
		BACKUP_ENABLED:           func() *bool { b := c.GetBackupEnabled(); return &b }(),
		BACKUP_RETENTION:         c.GetBackupRetention(),
		DEBUG:                    c.DEBUG,
		OIDC:                     c.OIDC,
		SONARR:                   c.SONARR, // Dont act safe, this contains sonarr api key, needed for config
		RADARR:                   c.RADARR, // Dont act safe, this contains radarr api key, needed for config
		TWITCH: igdb.IGDB{
			ClientID:     c.TWITCH.ClientID,
			ClientSecret: c.TWITCH.ClientSecret,
		}, // Dont act safe, this contains twitch secrets, needed for config
	}
}

func (c *ServerConfig) GetDefaultUserPermissions() int {
	if c.DEFAULT_USER_PERMISSIONS <= 0 {
		return 1 // entity.PERM_NONE
	}
	return c.DEFAULT_USER_PERMISSIONS
}

func (c *ServerConfig) GetTVDBKey() string {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	if c.TVDB_KEY != "" {
		return c.TVDB_KEY
	}
	return strings.TrimSpace(os.Getenv("TVDB_API_KEY"))
}

func (c *ServerConfig) GetMetadataProvider() string {
	if strings.EqualFold(c.METADATA_PROVIDER, "tvdb") && (c.TVDB_KEY != "" || os.Getenv("TVDB_API_KEY") != "") {
		return "tvdb"
	}
	return "tmdb"
}

func (c *ServerConfig) GetDefaultLanguage() string {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	if c.DEFAULT_LANGUAGE != "" {
		return c.DEFAULT_LANGUAGE
	}
	return "en-US"
}

func (c *ServerConfig) GetOMDBKey() string {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	if c.OMDB_KEY != "" {
		return c.OMDB_KEY
	}
	return strings.TrimSpace(os.Getenv("OMDB_API_KEY"))
}

func (c *ServerConfig) GetMDBListKey() string {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	if c.MDBLIST_KEY != "" {
		return c.MDBLIST_KEY
	}
	return strings.TrimSpace(os.Getenv("MDBLIST_API_KEY"))
}

func (c *ServerConfig) GetSteamKey() string {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	if c.STEAM_KEY != "" {
		return c.STEAM_KEY
	}
	if env := strings.TrimSpace(os.Getenv("STEAM_API_KEY")); env != "" {
		return env
	}
	return strings.TrimSpace(os.Getenv("STEAM_KEY"))
}

func (c *ServerConfig) GetBackupEnabled() bool {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	if c.BACKUP_ENABLED == nil {
		return true
	}
	return *c.BACKUP_ENABLED
}

func (c *ServerConfig) SetBackupEnabled(b bool) {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	c.BACKUP_ENABLED = &b
}

func (c *ServerConfig) GetBackupRetention() int {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	if c.BACKUP_RETENTION > 0 {
		return c.BACKUP_RETENTION
	}
	return 7
}

func (c *ServerConfig) SetBackupRetention(r int) {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	if r <= 0 {
		r = 7
	}
	c.BACKUP_RETENTION = r
}

func (c *ServerConfig) GetTaskSchedule(name string) int {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	if c.TASK_SCHEDULE == nil {
		return 0
	}
	return c.TASK_SCHEDULE[name]
}

func (c *ServerConfig) SetTaskSchedule(name string, seconds int) {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	if c.TASK_SCHEDULE == nil {
		c.TASK_SCHEDULE = map[string]int{}
	}
	c.TASK_SCHEDULE[name] = seconds
}

type ServerConfigGetByName struct {
	Value any `json:"value"`
}

// Get config item by name.
func (c *ServerConfig) Get(s string) (ServerConfigGetByName, error) {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	switch s {
	case "DEFAULT_COUNTRY":
		return ServerConfigGetByName{Value: c.DEFAULT_COUNTRY}, nil
	case "DEFAULT_LANGUAGE":
		return ServerConfigGetByName{Value: c.DEFAULT_LANGUAGE}, nil
	case "DEFAULT_USER_PERMISSIONS":
		return ServerConfigGetByName{Value: c.GetDefaultUserPermissions()}, nil
	case "JELLYFIN_HOST":
		return ServerConfigGetByName{Value: c.JELLYFIN_HOST}, nil
	case "JELLYFIN_WEBHOOK_ENABLED":
		return ServerConfigGetByName{Value: c.JELLYFIN_WEBHOOK_ENABLED}, nil
	case "JELLYFIN_WEBHOOK_KEY":
		return ServerConfigGetByName{Value: c.JELLYFIN_WEBHOOK_KEY}, nil
	case "USE_EMBY":
		return ServerConfigGetByName{Value: c.USE_EMBY}, nil
	case "SIGNUP_ENABLED":
		return ServerConfigGetByName{Value: c.SIGNUP_ENABLED}, nil
	case "TMDB_KEY":
		return ServerConfigGetByName{Value: c.TMDB_KEY}, nil
	case "TVDB_KEY":
		return ServerConfigGetByName{Value: c.TVDB_KEY}, nil
	case "METADATA_PROVIDER":
		return ServerConfigGetByName{Value: c.GetMetadataProvider()}, nil
	case "OMDB_KEY":
		return ServerConfigGetByName{Value: c.OMDB_KEY}, nil
	case "MDBLIST_KEY":
		return ServerConfigGetByName{Value: c.MDBLIST_KEY}, nil
	case "STEAM_KEY":
		return ServerConfigGetByName{Value: c.STEAM_KEY}, nil
	case "PLEX_HOST":
		return ServerConfigGetByName{Value: c.PLEX_HOST}, nil
	case "PLEX_MACHINE_ID":
		return ServerConfigGetByName{Value: c.PLEX_MACHINE_ID}, nil
	case "HEADER_AUTH":
		return ServerConfigGetByName{Value: c.HEADER_AUTH}, nil
	case "OIDC":
		return ServerConfigGetByName{Value: c.OIDC}, nil
	case "DEBUG":
		return ServerConfigGetByName{Value: c.DEBUG}, nil
	case "OVERSEERR_HOST":
		return ServerConfigGetByName{Value: c.OVERSEERR_HOST}, nil
	case "OVERSEERR_KEY":
		return ServerConfigGetByName{Value: c.OVERSEERR_KEY}, nil
	case "BACKUP_ENABLED":
		return ServerConfigGetByName{Value: c.GetBackupEnabled()}, nil
	case "BACKUP_RETENTION":
		return ServerConfigGetByName{Value: c.GetBackupRetention()}, nil
	}
	return ServerConfigGetByName{}, errors.New("invalid setting")
}

// SetOIDCSetting saves OpenID Connect / OAuth2 settings (#461).
func (c *ServerConfig) SetOIDCSetting(oidc OIDCSetting) error {
	cfgMu.Lock()
	oidc.IssuerURL = strings.TrimRight(strings.TrimSpace(oidc.IssuerURL), "/")
	oidc.ClientID = strings.TrimSpace(oidc.ClientID)
	oidc.ClientSecret = strings.TrimSpace(oidc.ClientSecret)
	oidc.RedirectURI = strings.TrimSpace(oidc.RedirectURI)
	oidc.ProviderName = strings.TrimSpace(oidc.ProviderName)
	c.OIDC = oidc
	cfgMu.Unlock()
	if err := c.Write(); err != nil {
		slog.Error("SetOIDCSetting: Failed to write config!", "error", err)
		return errors.New("failed to write config")
	}
	return nil
}

// Update server config property
func (c *ServerConfig) UpdateConfig(k string, v any) error {
	slog.Debug("updateConfig", "k", k, "v", v)
	if v == nil {
		return errors.New("invalid value")
	}
	cfgMu.Lock()
	switch k {
	case "JELLYFIN_HOST":
		str, ok := v.(string)
		if !ok {
			cfgMu.Unlock()
			return errors.New("invalid value")
		}
		c.JELLYFIN_HOST = str
	case "USE_EMBY":
		b, ok := v.(bool)
		if !ok {
			cfgMu.Unlock()
			return errors.New("invalid value")
		}
		c.USE_EMBY = b
	case "SIGNUP_ENABLED":
		b, ok := v.(bool)
		if !ok {
			cfgMu.Unlock()
			return errors.New("invalid value")
		}
		c.SIGNUP_ENABLED = b
	case "DEFAULT_USER_PERMISSIONS":
		var perms int
		switch val := v.(type) {
		case float64:
			perms = int(val)
		case int:
			perms = val
		default:
			cfgMu.Unlock()
			return errors.New("invalid value")
		}
		if perms <= 0 {
			perms = 1
		}
		// Never allow default user permissions to include PERM_ADMIN (2) automatically
		perms &^= 2
		if perms == 0 {
			perms = 1
		}
		c.DEFAULT_USER_PERMISSIONS = perms
	case "TMDB_KEY":
		str, ok := v.(string)
		if !ok {
			cfgMu.Unlock()
			return errors.New("invalid value")
		}
		c.TMDB_KEY = strings.TrimSpace(str)
	case "TVDB_KEY":
		str, ok := v.(string)
		if !ok {
			cfgMu.Unlock()
			return errors.New("invalid value")
		}
		c.TVDB_KEY = strings.TrimSpace(str)
	case "METADATA_PROVIDER":
		str, ok := v.(string)
		if !ok {
			cfgMu.Unlock()
			return errors.New("invalid value")
		}
		p := strings.ToLower(strings.TrimSpace(str))
		if p != "tvdb" {
			p = "tmdb"
		}
		c.METADATA_PROVIDER = p
	case "OMDB_KEY":
		str, ok := v.(string)
		if !ok {
			cfgMu.Unlock()
			return errors.New("invalid value")
		}
		c.OMDB_KEY = strings.TrimSpace(str)
	case "MDBLIST_KEY":
		str, ok := v.(string)
		if !ok {
			cfgMu.Unlock()
			return errors.New("invalid value")
		}
		c.MDBLIST_KEY = strings.TrimSpace(str)
	case "STEAM_KEY":
		str, ok := v.(string)
		if !ok {
			cfgMu.Unlock()
			return errors.New("invalid value")
		}
		c.STEAM_KEY = strings.TrimSpace(str)
	case "DEBUG":
		b, ok := v.(bool)
		if !ok {
			cfgMu.Unlock()
			return errors.New("invalid value")
		}
		c.DEBUG = b
		logging.Level(c.DEBUG)
	case "DEFAULT_COUNTRY":
		str, ok := v.(string)
		if !ok {
			cfgMu.Unlock()
			return errors.New("invalid value")
		}
		c.DEFAULT_COUNTRY = str
	case "DEFAULT_LANGUAGE":
		str, ok := v.(string)
		if !ok {
			cfgMu.Unlock()
			return errors.New("invalid value")
		}
		c.DEFAULT_LANGUAGE = str
	case "OVERSEERR_HOST":
		str, ok := v.(string)
		if !ok {
			cfgMu.Unlock()
			return errors.New("invalid value")
		}
		c.OVERSEERR_HOST = strings.TrimRight(strings.TrimSpace(str), "/")
	case "BACKUP_ENABLED":
		b, ok := v.(bool)
		if !ok {
			cfgMu.Unlock()
			return errors.New("invalid value")
		}
		c.BACKUP_ENABLED = &b
	case "BACKUP_RETENTION":
		var r int
		switch val := v.(type) {
		case float64:
			r = int(val)
		case int:
			r = val
		default:
			cfgMu.Unlock()
			return errors.New("invalid value")
		}
		if r <= 0 {
			r = 7
		}
		c.BACKUP_RETENTION = r
	case "OVERSEERR_KEY":
		str, ok := v.(string)
		if !ok {
			cfgMu.Unlock()
			return errors.New("invalid value")
		}
		c.OVERSEERR_KEY = strings.TrimSpace(str)
	default:
		cfgMu.Unlock()
		return errors.New("invalid setting")
	}
	cfgMu.Unlock()

	err := c.Write()
	if err != nil {
		slog.Error("updateConfig: Failed to write updated config!", "error", err)
		return errors.New("failed to write config")
	}
	return nil
}

type SetJellyfinWebhookEnabledResponse struct {
	JELLYFIN_WEBHOOK_KEY uuid.UUID
}

// Update JELLYFIN_WEBHOOK_ENABLED.
// Will set a JELLYFIN_WEBHOOK_KEY if one doesn't exist and returns it.
func (c *ServerConfig) SetJellyfinWebhookEnabled(
	v bool,
) (SetJellyfinWebhookEnabledResponse, error) {
	cfgMu.Lock()
	c.JELLYFIN_WEBHOOK_ENABLED = v
	if c.JELLYFIN_WEBHOOK_ENABLED && c.JELLYFIN_WEBHOOK_KEY == uuid.Nil() {
		// If we are enabling the webhook and we don't already have a key,
		// create one. We don't re-gen the key everytime so users can freely
		// enable/disable the webhook support without having to reconfigure
		// jellyfin with the new key when they re-enable.
		c.JELLYFIN_WEBHOOK_KEY = uuid.New()
	}
	key := c.JELLYFIN_WEBHOOK_KEY
	cfgMu.Unlock()

	if err := c.Write(); err != nil {
		slog.Error("SetJellyfinWebhookEnabled: Failed to write config!",
			"err", err)
		return SetJellyfinWebhookEnabledResponse{},
			errors.New("failed to write config")
	}

	return SetJellyfinWebhookEnabledResponse{
		JELLYFIN_WEBHOOK_KEY: key,
	}, nil
}

// Write current Config to file atomically via a temporary file.
func (c *ServerConfig) Write() error {
	cfgMu.RLock()
	barej, err := json.MarshalIndent(c, "", "\t")
	cfgMu.RUnlock()
	if err != nil {
		return err
	}
	cfgPath := path.Join(DataPath, "watcharr.json")
	tmpPath := cfgPath + ".tmp"
	if err := os.WriteFile(tmpPath, barej, 0600); err != nil {
		return err
	}
	return os.Rename(tmpPath, cfgPath)
}

func (c *ServerConfig) SaveTwitchConfig(newt igdb.IGDB) error {
	// If existing client id and secret are same.. just return here
	if (c.TWITCH.ClientID != nil && newt.ClientID != nil && c.TWITCH.ClientSecret != nil && newt.ClientSecret != nil) &&
		*c.TWITCH.ClientID == *newt.ClientID && *c.TWITCH.ClientSecret == *newt.ClientSecret {
		slog.Info("SaveTwitchConfig: New ClientID and ClientSecret match old ClientID and ClientSecret.. ignoring request to update.")
		return nil
	}
	// Update our config
	c.TWITCH.ClientID = newt.ClientID
	c.TWITCH.ClientSecret = newt.ClientSecret
	c.TWITCH.AccessToken = ""
	c.TWITCH.AccessTokenExpires = time.Time{}
	// Try to init again
	err := c.TWITCH.Init()
	if err != nil {
		slog.Error("SaveTwitchConfig failed to initialize TWITCH", "error", err)
		return errors.New("initialization with credentials failed")
	}
	err = c.Write()
	if err != nil {
		slog.Error("SaveTwitchConfig failed to write config", "error", err)
		return errors.New("failed to save config")
	}
	return nil
}

func (c *ServerConfig) TwitchEnabled() bool {
	if c.TWITCH.ClientID != nil && c.TWITCH.ClientSecret != nil {
		return true
	}
	return false
}

// Read config file
// Calls generateConfig if file doesn't exist
func read() (*ServerConfig, error) {
	cfgFile, err := os.Open(path.Join(DataPath, "watcharr.json"))
	if err != nil {
		if os.IsNotExist(err) {
			slog.Info("Config file doesn't exist... generating.")
			if genCfg, err := generateConfig(); err == nil {
				return genCfg, nil
			}
		}
		return nil, err
	}
	defer cfgFile.Close()

	c := new(ServerConfig)
	dec := json.NewDecoder(cfgFile)
	if err = dec.Decode(c); err != nil {
		return nil, err
	}

	initFromConfig(c)

	return c, nil
}

// Ensure required config is provided
func initFromConfig(c *ServerConfig) {
	if c.JWT_SECRET == "" {
		log.Fatal("JWT_SECRET missing from config!")
	}
	if c.DEFAULT_LANGUAGE == "" {
		c.DEFAULT_LANGUAGE = "en-US"
	}
}

// Generate new barebones watcharr.json config file.
// Generates a JWT_SECRET and set default config.
func generateConfig() (*ServerConfig, error) {
	key, err := util.GenerateString(64)
	if err != nil {
		return nil, err
	}
	cfg := ServerConfig{
		JWT_SECRET: key,
		// Other defaults..
		DEFAULT_COUNTRY:  "US",
		DEFAULT_LANGUAGE: "en-US",
		SIGNUP_ENABLED:   true,
	}
	barej, err := json.MarshalIndent(cfg, "", "\t")
	if err != nil {
		return nil, err
	}
	return &cfg, os.WriteFile(path.Join(DataPath, "watcharr.json"), barej, 0600)
}

// Get server config.
// Reads from config file.
func Get() (*ServerConfig, error) {
	cfg, err := read()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *ServerConfig) OverseerrEnabled() bool {
	cfgMu.RLock()
	defer cfgMu.RUnlock()
	return c.OVERSEERR_HOST != "" && c.OVERSEERR_KEY != ""
}
