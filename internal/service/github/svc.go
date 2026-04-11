// Package githubsvc implements GitHub OAuth2 flow.
package githubsvc

import (
	"emotionalBeach/config"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultRedirectURI = "https://api.ymmos.com/callback"
	authURLTemplate    = "https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user"
	tokenURL           = "https://github.com/login/oauth/access_token"
	userAPIURL         = "https://api.github.com/user"
)

// Svc encapsulates GitHub OAuth configuration and HTTP transport.
type Svc struct {
	clientID     string
	clientSecret string
	redirectURI  string
	httpClient   *http.Client
}

// NewSvc constructs a Svc from application config.
func NewSvc(cfg *config.Config) *Svc {
	return &Svc{
		clientID:     cfg.Server.ClientID,
		clientSecret: cfg.Server.ClientSecret,
		redirectURI:  defaultRedirectURI,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
	}
}

// AuthURL returns the GitHub OAuth2 authorization URL.
func (s *Svc) AuthURL() string {
	return fmt.Sprintf(authURLTemplate, s.clientID, url.QueryEscape(s.redirectURI))
}

// RedirectURI returns the configured callback URI.
func (s *Svc) RedirectURI() string {
	return s.redirectURI
}

// ExchangeToken exchanges an authorization code for an access token.
func (s *Svc) ExchangeToken(code string) (string, error) {
	data := url.Values{}
	data.Set("client_id", s.clientID)
	data.Set("client_secret", s.clientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", s.redirectURI)

	req, err := http.NewRequest(http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("github token endpoint returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read token response: %w", err)
	}
	var result map[string]interface{}
	if err = json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse token response: %w", err)
	}
	token, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("access_token not found in response")
	}
	return token, nil
}

// GetUserInfo fetches the authenticated GitHub user's profile.
func (s *Svc) GetUserInfo(accessToken string) (map[string]interface{}, error) {
	req, err := http.NewRequest(http.MethodGet, userAPIURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build user request: %w", err)
	}
	req.Header.Set("Authorization", "token "+accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("user request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("github user endpoint returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read user response: %w", err)
	}
	var info map[string]interface{}
	if err = json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parse user response: %w", err)
	}
	return info, nil
}

