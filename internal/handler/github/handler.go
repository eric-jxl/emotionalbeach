// Package githubhandler wires the HTTP presentation layer for GitHub OAuth.
package githubhandler

import (
	"emotionalBeach/internal/service/github"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Handler holds injected GitHub OAuth service.
type Handler struct {
	svc *githubsvc.Svc
}

// NewHandler constructs a Handler.
func NewHandler(svc *githubsvc.Svc) *Handler {
	return &Handler{svc: svc}
}

// GithubLogin godoc
// @Summary      GitHub 登录
// @Description  GitHub 一键授权登录，重定向到 GitHub OAuth 页面
// @Tags         注册登陆
// @Router       /login/github [get]
func (h *Handler) GithubLogin(c *gin.Context) {
	c.Redirect(http.StatusFound, h.svc.AuthURL())
}

// GithubCallback godoc
// @Summary      GitHub 回调接口
// @Description  GitHub 授权成功回调，换取 access_token 并获取用户信息
// @Tags         注册登陆
// @Router       /callback [get]
func (h *Handler) GithubCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code not found"})
		return
	}

	accessToken, err := h.svc.ExchangeToken(code)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	_, err = h.svc.GetUserInfo(accessToken)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	// TODO: generate JWT for the GitHub user, then redirect to the app.
	redirectURI := h.svc.RedirectURI()
	c.Redirect(http.StatusFound, strings.Replace(redirectURI, "callback", "swagger/index.html", 1))
}

