package server

import (
	"emotionalBeach/internal/common"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// githubLogin godoc
// @Summary      GitHub 登录
// @Tags         注册登陆
// @Router       /login/github [get]
func githubLogin(c *gin.Context) {
	c.Redirect(http.StatusFound, svc.AuthURL())
}

// githubCallback godoc
// @Summary      GitHub 回调接口
// @Tags         注册登陆
// @Router       /callback [get]
func githubCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		common.Fail(c, http.StatusBadRequest, "code not found")
		return
	}
	accessToken, err := svc.ExchangeToken(code)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}
	if _, err = svc.GetUserInfo(accessToken); err != nil {
		common.ServerError(c, err.Error())
		return
	}
	redirectURI := svc.RedirectURI()
	c.Redirect(http.StatusFound, strings.Replace(redirectURI, "callback", "swagger/index.html", 1))
}
