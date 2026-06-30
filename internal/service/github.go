package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"emotionalBeach/internal/common"
	ebmetrics "emotionalBeach/internal/infra"
	"emotionalBeach/internal/models"
)

const (
	defaultRedirectURI = "https://api.ymmos.com/callback"
	authURLTemplate    = "https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user"
	tokenURL           = "https://github.com/login/oauth/access_token"
	userAPIURL         = "https://api.github.com/user"
)

// AuthURL returns the GitHub OAuth2 authorization URL.
func (s *Service) AuthURL() string {
	return fmt.Sprintf(authURLTemplate, s.clientID, url.QueryEscape(s.redirectURI))
}

// RedirectURI returns the configured callback URI.
func (s *Service) RedirectURI() string {
	return s.redirectURI
}

// ExchangeToken exchanges an authorization code for an access token.
func (s *Service) ExchangeToken(code string) (string, error) {
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
func (s *Service) GetUserInfo(accessToken string) (map[string]interface{}, error) {
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

// LoginWithGitHub resolves a GitHub profile to a local user, creating one when
// necessary, then refreshes the user's identity token and returns the user so
// the caller can issue a JWT. GitHub OAuth2 does not require a captcha.
func (s *Service) LoginWithGitHub(info map[string]interface{}) (*models.UserBasic, error) {
	ghID, err := toInt64(info["id"])
	if err != nil || ghID == 0 {
		return nil, errors.New("invalid github user id")
	}

	user, derr := s.dao.FindUserByGitHubID(ghID)
	if derr != nil || user == nil {
		created, cerr := s.createGitHubUser(ghID, info)
		if cerr != nil {
			ebmetrics.UserLoginsTotal.WithLabelValues("github_create_error").Inc()
			return nil, cerr
		}
		user = created
		ebmetrics.UserRegistrationsTotal.Inc()
	}

	now := time.Now()
	if user.LoginTime == nil {
		user.LoginTime = &now
	}
	identity := common.Md5encoder(strconv.Itoa(int(time.Now().Unix())))
	if err := s.dao.UpdateIdentity(user.ID, identity); err != nil {
		ebmetrics.UserLoginsTotal.WithLabelValues("github_token_error").Inc()
		return nil, err
	}
	user.Identity = identity

	// Sync avatar/login from GitHub when available.
	if name := toString(info["login"]); name != "" && user.Name == "" {
		user.Name = name
	}
	if avatar := toString(info["avatar_url"]); avatar != "" {
		user.Avatar = avatar
	}

	ebmetrics.UserLoginsTotal.WithLabelValues("github_success").Inc()
	return user, nil
}

// createGitHubUser creates a local user record linked to the GitHub identity.
func (s *Service) createGitHubUser(ghID int64, info map[string]interface{}) (*models.UserBasic, error) {
	login := toString(info["login"])
	name := login
	if name == "" {
		name = "gh_" + strconv.FormatInt(ghID, 10)
	}
	// Ensure the username is unique; append the numeric GitHub id if necessary.
	if s.dao.UserNameExists(name) {
		name = "gh_" + strconv.FormatInt(ghID, 10)
	}

	now := time.Now()
	user := models.UserBasic{
		Name:          name,
		Role:          "user",
		Avatar:        toString(info["avatar_url"]),
		Email:         toString(info["email"]),
		GitHubID:      ghID,
		Phone:         "0" + strconv.FormatInt(ghID, 10), // phone is NOT NULL; synthetic value for OAuth users
		LoginTime:     &now,
		LoginOutTime:  &now,
		HeartBeatTime: &now,
	}
	created, err := s.dao.CreateUser(user)
	if err != nil {
		return nil, fmt.Errorf("create github user: %w", err)
	}
	return created, nil
}

func toInt64(v interface{}) (int64, error) {
	switch x := v.(type) {
	case float64:
		return int64(x), nil
	case json.Number:
		return x.Int64()
	case string:
		return strconv.ParseInt(x, 10, 64)
	}
	return 0, fmt.Errorf("unexpected type %T", v)
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	}
	return ""
}
