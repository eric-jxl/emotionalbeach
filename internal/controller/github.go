package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	ClientID     string
	ClientSecret string
	redirectURI  = "https://api.ymmos.com/callback"
)

// GithubLogin 登录注册
// @Summary GitHub 登录
// @Description GitHub 一键授权登录
// @Tags 注册登陆
// @Accept application/x-www-form-urlencoded
// @Produce application/x-www-form-urlencoded
// @Router /login/github [get]
func GithubLogin(c *gin.Context) {
	authURL := fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=user", ClientID, url.QueryEscape(redirectURI))
	c.Redirect(http.StatusFound, authURL)
}

// GithubCallback 登录注册
// @Summary GitHub 回调接口
// @Description GitHub 授权成功回调接口
// @Tags 注册登陆
// @Accept application/x-www-form-urlencoded
// @Produce application/x-www-form-urlencoded
// @Router /callback [get]
func GithubCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code not found"})
		return
	}

	// Step 3: 用 code 换取 access_token
	tokenURL := "https://github.com/login/oauth/access_token"
	data := url.Values{}
	data.Set("client_id", ClientID)
	data.Set("client_secret", ClientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	req, _ := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var tokenResp map[string]interface{}
	_ = json.Unmarshal(body, &tokenResp)

	accessToken, ok := tokenResp["access_token"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get access token"})
		return
	}

	// Step 4: 用 access_token 获取 GitHub 用户信息
	userReq, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	userReq.Header.Set("Authorization", "token "+accessToken)

	userResp, err := http.DefaultClient.Do(userReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer userResp.Body.Close()

	userBody, _ := io.ReadAll(userResp.Body)
	var userInfo map[string]interface{}
	_ = json.Unmarshal(userBody, &userInfo)

	// Step 5: 生成 JWT

	c.Redirect(http.StatusFound, strings.Replace(redirectURI, "callback", "swagger/index.html", 1))
}
