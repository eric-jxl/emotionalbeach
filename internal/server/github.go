package server

import (
	"emotionalBeach/internal/common"
	"emotionalBeach/internal/middleware"
	"encoding/json"
	"net/http"

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
	info, err := svc.GetUserInfo(accessToken)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}
	user, err := svc.LoginWithGitHub(info)
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}
	token, err := middleware.GenerateToken(user.ID, user.Name)
	if err != nil {
		common.ServerError(c, "生成 token 失败")
		return
	}
	// GitHub OAuth2 不走验证码：直接签发 JWT 并通过页面写入前端 localStorage。
	renderOAuthCallback(c, token, "/swagger/index.html")
}

// renderOAuthCallback writes the token into the browser localStorage and then
// redirects to the target page. This avoids encoding the JWT in the URL while
// still satisfying the SPA frontend that reads eb_token from storage.
func renderOAuthCallback(c *gin.Context, token, redirect string) {
	tokenLit, _ := json.Marshal(token)
	redirectLit, _ := json.Marshal(redirect)
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>登录中</title>
<style>
body{margin:0;display:grid;place-items:center;min-height:100vh;background:#edf3ff;font-family:"PingFang SC","Microsoft YaHei",sans-serif}
.card{padding:40px 56px;border-radius:16px;background:rgba(255,255,255,.9);box-shadow:0 16px 48px rgba(15,23,42,.18);text-align:center}
.dot{display:inline-block;width:10px;height:10px;border-radius:50%;background:#2563eb;margin:0 4px;animation:pulse 1.2s infinite}
.dot:nth-child(2){animation-delay:.2s}
.dot:nth-child(3){animation-delay:.4s}
@keyframes pulse{0%,80%,100%{opacity:.3;transform:scale(.8)}40%{opacity:1;transform:scale(1.1)}}
</style>
</head>
<body>
<div class="card">
<h3>正在登录，请稍候…</h3>
<div style="margin-top:12px"><span class="dot"></span><span class="dot"></span><span class="dot"></span></div>
</div>
<script>
(function(){
  try {
    localStorage.setItem('eb_token', ` + string(tokenLit) + `);
    localStorage.setItem('eb_token_exp', String(Date.now() + 7*24*60*60*1000));
  } catch (e) {}
  window.location.replace(` + string(redirectLit) + `);
})();
</script>
</body>
</html>`
	c.Header("Cache-Control", "no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
