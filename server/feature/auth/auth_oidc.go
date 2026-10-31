package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
	"gorm.io/gorm"
)

type oidcDiscoveryDoc struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
}

type oidcTokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

type oidcUserInfo struct {
	Sub               string `json:"sub"`
	PreferredUsername string `json:"preferred_username"`
	Nickname          string `json:"nickname"`
	Name              string `json:"name"`
	Email             string `json:"email"`
}

type OIDCCallbackRequest struct {
	Code        string `json:"code" binding:"required"`
	State       string `json:"state" binding:"required"`
	RedirectURI string `json:"redirectUri"`
}

const oidcStateMaxAge = 10 * time.Minute

// newOIDCState creates a stateless, HMAC signed `state` value (nonce.timestamp.signature)
// used to protect the OIDC flow against CSRF.
func (s *Service) newOIDCState() (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(nonce) + "." + strconv.FormatInt(time.Now().Unix(), 10)
	return payload + "." + s.signOIDCState(payload), nil
}

func (s *Service) signOIDCState(payload string) string {
	mac := hmac.New(sha256.New, []byte(s.cfg.JWT_SECRET))
	mac.Write([]byte("oidc-state:" + payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *Service) validateOIDCState(state string) error {
	parts := strings.Split(state, ".")
	if len(parts) != 3 {
		return errors.New("invalid state")
	}
	payload := parts[0] + "." + parts[1]
	if subtle.ConstantTimeCompare([]byte(parts[2]), []byte(s.signOIDCState(payload))) != 1 {
		return errors.New("invalid state")
	}
	ts, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || time.Since(time.Unix(ts, 0)) > oidcStateMaxAge || ts > time.Now().Unix()+60 {
		return errors.New("state expired, please try signing in again")
	}
	return nil
}

type OIDCLoginURLResponse struct {
	AuthURL string `json:"authUrl"`
}

var oidcHTTPClient = &http.Client{
	Timeout: 15 * time.Second,
}

func (s *Service) discoverOIDC() (*oidcDiscoveryDoc, error) {
	issuer := strings.TrimRight(strings.TrimSpace(s.cfg.OIDC.IssuerURL), "/")
	if issuer == "" {
		return nil, errors.New("OIDC issuer URL is not configured")
	}
	wellKnownURL := issuer
	if !strings.HasSuffix(wellKnownURL, "/.well-known/openid-configuration") {
		wellKnownURL += "/.well-known/openid-configuration"
	}

	req, err := http.NewRequest("GET", wellKnownURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := oidcHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to contact OIDC issuer: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OIDC discovery returned status %d", resp.StatusCode)
	}

	var doc oidcDiscoveryDoc
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("invalid OIDC discovery document: %w", err)
	}
	if doc.AuthorizationEndpoint == "" || doc.TokenEndpoint == "" {
		return nil, errors.New("OIDC discovery document missing authorization_endpoint or token_endpoint")
	}
	return &doc, nil
}

// GetOIDCLoginURL builds the OAuth2 / OIDC authorization URL (#461).
func (s *Service) GetOIDCLoginURL(fallbackRedirectURI string) (OIDCLoginURLResponse, error) {
	if !s.cfg.OIDC.Enabled || s.cfg.OIDC.IssuerURL == "" || s.cfg.OIDC.ClientID == "" {
		return OIDCLoginURLResponse{}, errors.New("OIDC authentication is not enabled")
	}
	doc, err := s.discoverOIDC()
	if err != nil {
		slog.Error("GetOIDCLoginURL: discovery failed", "error", err)
		return OIDCLoginURLResponse{}, err
	}

	redirectURI := strings.TrimSpace(s.cfg.OIDC.RedirectURI)
	if redirectURI == "" {
		redirectURI = strings.TrimSpace(fallbackRedirectURI)
	}
	if redirectURI == "" {
		return OIDCLoginURLResponse{}, errors.New("missing redirect URI")
	}

	authURL, err := url.Parse(doc.AuthorizationEndpoint)
	if err != nil {
		return OIDCLoginURLResponse{}, errors.New("invalid authorization_endpoint URL")
	}

	q := authURL.Query()
	q.Set("client_id", s.cfg.OIDC.ClientID)
	q.Set("response_type", "code")
	q.Set("scope", "openid profile email")
	q.Set("redirect_uri", redirectURI)
	state, err := s.newOIDCState()
	if err != nil {
		return OIDCLoginURLResponse{}, errors.New("failed to generate state")
	}
	q.Set("state", state)
	authURL.RawQuery = q.Encode()

	return OIDCLoginURLResponse{AuthURL: authURL.String()}, nil
}

func parseIDTokenClaims(idToken string) (*oidcUserInfo, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) < 2 {
		return nil, errors.New("invalid id_token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var info oidcUserInfo
	if err := json.Unmarshal(payload, &info); err != nil {
		return nil, err
	}
	return &info, nil
}

// LoginOIDC exchanges an authorization code for tokens and logs in or registers the user (#461).
func (s *Service) LoginOIDC(req OIDCCallbackRequest) (AuthResponse, error) {
	if !s.cfg.OIDC.Enabled || s.cfg.OIDC.IssuerURL == "" || s.cfg.OIDC.ClientID == "" {
		return AuthResponse{}, errors.New("OIDC authentication is not enabled")
	}
	if err := s.validateOIDCState(req.State); err != nil {
		slog.Warn("LoginOIDC: state validation failed", "error", err)
		return AuthResponse{}, err
	}

	doc, err := s.discoverOIDC()
	if err != nil {
		slog.Error("LoginOIDC: discovery failed", "error", err)
		return AuthResponse{}, errors.New("failed to discover OIDC provider")
	}

	redirectURI := strings.TrimSpace(s.cfg.OIDC.RedirectURI)
	if redirectURI == "" {
		redirectURI = strings.TrimSpace(req.RedirectURI)
	}

	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", req.Code)
	form.Set("client_id", s.cfg.OIDC.ClientID)
	if s.cfg.OIDC.ClientSecret != "" {
		form.Set("client_secret", s.cfg.OIDC.ClientSecret)
	}
	if redirectURI != "" {
		form.Set("redirect_uri", redirectURI)
	}

	tokenReq, err := http.NewRequest("POST", doc.TokenEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return AuthResponse{}, errors.New("failed to build token request")
	}
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokenReq.Header.Set("Accept", "application/json")

	tokenResp, err := oidcHTTPClient.Do(tokenReq)
	if err != nil {
		slog.Error("LoginOIDC: token exchange request failed", "error", err)
		return AuthResponse{}, errors.New("failed to exchange OIDC code")
	}
	defer tokenResp.Body.Close()

	bodyBytes, _ := io.ReadAll(tokenResp.Body)
	var tok oidcTokenResponse
	if err := json.Unmarshal(bodyBytes, &tok); err != nil {
		slog.Error("LoginOIDC: failed to parse token response", "error", err, "status", tokenResp.StatusCode)
		return AuthResponse{}, errors.New("invalid token response from OIDC provider")
	}
	if tokenResp.StatusCode != http.StatusOK || tok.AccessToken == "" {
		slog.Error("LoginOIDC: non-200 or empty access token", "status", tokenResp.StatusCode, "error", tok.Error, "desc", tok.ErrorDesc)
		if tok.ErrorDesc != "" {
			return AuthResponse{}, errors.New(tok.ErrorDesc)
		}
		return AuthResponse{}, errors.New("OIDC token exchange failed")
	}

	var info oidcUserInfo
	if doc.UserinfoEndpoint != "" {
		uReq, err := http.NewRequest("GET", doc.UserinfoEndpoint, nil)
		if err == nil {
			uReq.Header.Set("Authorization", "Bearer "+tok.AccessToken)
			uReq.Header.Set("Accept", "application/json")
			if uResp, err := oidcHTTPClient.Do(uReq); err == nil {
				defer uResp.Body.Close()
				if uResp.StatusCode == http.StatusOK {
					_ = json.NewDecoder(uResp.Body).Decode(&info)
				}
			}
		}
	}

	// Fallback or supplement with id_token claims
	if (info.Sub == "" || (info.PreferredUsername == "" && info.Email == "" && info.Name == "")) && tok.IDToken != "" {
		if idClaims, err := parseIDTokenClaims(tok.IDToken); err == nil && idClaims != nil {
			if info.Sub == "" {
				info.Sub = idClaims.Sub
			}
			if info.PreferredUsername == "" {
				info.PreferredUsername = idClaims.PreferredUsername
			}
			if info.Nickname == "" {
				info.Nickname = idClaims.Nickname
			}
			if info.Name == "" {
				info.Name = idClaims.Name
			}
			if info.Email == "" {
				info.Email = idClaims.Email
			}
		}
	}

	username := strings.TrimSpace(info.PreferredUsername)
	if username == "" {
		username = strings.TrimSpace(info.Nickname)
	}
	if username == "" && info.Email != "" {
		username = strings.Split(strings.TrimSpace(info.Email), "@")[0]
	}
	if username == "" {
		username = strings.TrimSpace(info.Name)
	}
	if username == "" && info.Sub != "" {
		username = "oidc_" + info.Sub
	}
	if username == "" || info.Sub == "" {
		return AuthResponse{}, errors.New("OIDC provider did not return a valid user identifier (sub/username)")
	}

	dbUser := new(entity.User)
	userID, err := s.GetUserIDFromServiceClientId(entity.UserServiceNameOIDC, info.Sub)
	if err == nil && userID > 0 {
		if err := s.db.Where("id = ?", userID).Take(dbUser).Error; err != nil {
			return AuthResponse{}, errors.New("failed to locate linked OIDC user")
		}
	} else {
		// Check if an OIDC_USER with this username already exists
		if err := s.db.Where("username = ? AND type = ?", username, entity.OIDC_USER).Take(dbUser).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return AuthResponse{}, errors.New("database error looking up OIDC user")
			}
			if !s.cfg.OIDC.AutoRegister && !s.cfg.SIGNUP_ENABLED {
				return AuthResponse{}, errors.New("automatic registration for OIDC users is disabled")
			}
			// Determine initial permissions (first user on server gets admin, others get DEFAULT_USER_PERMISSIONS)
			var userCount int64
			s.db.Model(&entity.User{}).Count(&userCount)
			perms := s.cfg.GetDefaultUserPermissions()
			if userCount == 0 {
				perms = entity.PERM_ADMIN
			}

			dbUser = &entity.User{
				Username:    username,
				Type:        entity.OIDC_USER,
				Permissions: perms,
				UserServices: []entity.UserServices{
					{
						Name:      entity.UserServiceNameOIDC,
						ClientID:  info.Sub,
						AuthToken: tok.AccessToken,
					},
				},
			}
			dbUser.Country = &s.cfg.DEFAULT_COUNTRY
			dbUser.Language = &s.cfg.DEFAULT_LANGUAGE
			if err := s.db.Create(dbUser).Error; err != nil {
				slog.Error("LoginOIDC: Failed to create new OIDC user", "error", err)
				return AuthResponse{}, errors.New("failed to create user from OIDC")
			}
		}
	}

	jwtToken, err := s.signJWT(dbUser.ID, dbUser.Username, dbUser.Type)
	if err != nil {
		return AuthResponse{}, errors.New("failed to sign authentication token")
	}
	return AuthResponse{Token: jwtToken}, nil
}

