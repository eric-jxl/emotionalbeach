package server

import (
	"emotionalBeach/internal/common"
	"emotionalBeach/internal/middleware"
	"emotionalBeach/internal/models"
	"emotionalBeach/internal/service"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type userView struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Gender string `json:"gender"`
	Phone  string `json:"phone"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

func toUserView(u models.UserBasic) userView {
	return userView{
		ID: u.ID, Name: u.Name, Avatar: u.Avatar,
		Gender: u.Gender, Phone: u.Phone, Email: u.Email, Role: u.Role,
	}
}

// getUsers godoc
// @Summary      获取所有用户
// @Tags         用户
// @Produce      json
// @Security     ApiKeyAuth
// @Router       /v1/user/list [get]
func getUsers(c *gin.Context) {
	list, err := svc.GetList()
	if err != nil {
		common.ServerError(c, err.Error())
		return
	}
	views := make([]userView, 0, len(list))
	for _, u := range list {
		views = append(views, toUserView(u))
	}
	common.Success(c, views)
}

// getAppointUser godoc
// @Summary      按条件查询用户
// @Tags         用户
// @Param        id     query string false "用户 ID"
// @Param        phone  query string false "手机号"
// @Param        email  query string false "Email"
// @Produce      json
// @Security     ApiKeyAuth
// @Router       /v1/user/condition [get]
func getAppointUser(c *gin.Context) {
	id, phone, email := c.Query("id"), c.Query("phone"), c.Query("email")
	if id == "" && phone == "" && email == "" {
		common.Fail(c, http.StatusBadRequest, "至少需要一个参数（id、phone 或 email）")
		return
	}
	var (
		u   *models.UserBasic
		err error
	)
	switch {
	case id != "":
		uid, convErr := strconv.Atoi(id)
		if convErr != nil {
			common.Fail(c, http.StatusBadRequest, fmt.Sprintf("非法用户 ID: %v", convErr))
			return
		}
		u, err = svc.GetByID(uint(uid))
	case email != "":
		u, err = svc.GetByEmail(email)
	default:
		u, err = svc.GetByPhone(phone)
	}
	if err != nil {
		common.Fail(c, http.StatusNotFound, err.Error())
		return
	}
	common.Success(c, toUserView(*u))
}

// login godoc
// @Summary      登陆获取 Token
// @Tags         注册登陆
// @Accept       application/json
// @Produce      application/json
// @Param        req body models.LoginRequest true "登录参数"
// @Router       /login [post]
func login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}
	authUser, err := svc.Login(req.Username, req.Password)
	if err != nil {
		common.Fail(c, http.StatusUnauthorized, err.Error())
		return
	}
	token, err := middleware.GenerateToken(authUser.ID, authUser.Name)
	if err != nil {
		zap.S().Errorf("generate token: %v", err)
		common.ServerError(c, "生成 token 失败")
		return
	}
	common.Success(c, gin.H{"message": "登录成功", "token": token})
}

// register godoc
// @Summary      创建用户
// @Tags         注册登陆
// @Accept       multipart/form-data
// @Produce      application/json
// @Param        name            formData string true  "Name"
// @Param        password        formData string true  "Password"
// @Param        repeat_password formData string true  "repeat_password"
// @Param        phone           formData string true  "Phone"
// @Param        email           formData string false "EMAIL"
// @Router       /register [post]
func register(c *gin.Context) {
	name := c.PostForm("name")
	password := c.PostForm("password")
	repeatPwd := c.PostForm("repeat_password")
	phone := c.PostForm("phone")
	email := c.PostForm("email")

	switch {
	case name == "" || password == "" || repeatPwd == "":
		common.Fail(c, http.StatusBadRequest, "用户名、密码、确认密码不能为空")
		return
	case password != repeatPwd:
		common.Fail(c, http.StatusBadRequest, "两次密码不一致")
		return
	case phone == "":
		common.Fail(c, http.StatusBadRequest, "手机号不能为空")
		return
	case len(phone) != 11:
		common.Fail(c, http.StatusBadRequest, "手机号必须为 11 位")
		return
	case !common.IsValidPhoneNumber(phone):
		common.Fail(c, http.StatusBadRequest, "手机号格式非法")
		return
	}
	created, err := svc.Register(name, password, phone, email)
	if err != nil {
		common.Fail(c, http.StatusConflict, err.Error())
		return
	}
	common.Success(c, toUserView(*created))
}

// updateUser godoc
// @Summary      更新用户信息
// @Tags         用户
// @Param        id       formData string true  "ID"
// @Param        name     formData string false "用户名"
// @Param        password formData string false "密码"
// @Param        phone    formData string false "手机号"
// @Param        email    formData string false "Email"
// @Param        avatar   formData string false "头像"
// @Param        gender   formData string false "性别"
// @Produce      json
// @Security     ApiKeyAuth
// @Router       /v1/user/update [post]
func updateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.PostForm("id"))
	if err != nil {
		common.Fail(c, http.StatusBadRequest, "非法用户 ID")
		return
	}
	req := service.UpdateRequest{
		ID:       uint(id),
		Name:     c.PostForm("name"),
		Password: c.PostForm("password"),
		Email:    c.PostForm("email"),
		Phone:    c.PostForm("phone"),
		Avatar:   c.PostForm("avatar"),
		Gender:   c.PostForm("gender"),
	}
	updated, err := svc.Update(req)
	if err != nil {
		common.ServerError(c, "修改信息失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"uid": updated.ID})
}

// deleteUser godoc
// @Summary      删除用户
// @Tags         用户
// @Param        id query uint true "ID"
// @Produce      json
// @Security     ApiKeyAuth
// @Router       /v1/user/delete [delete]
func deleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		common.Fail(c, http.StatusBadRequest, "非法用户 ID")
		return
	}
	if err = svc.DeleteUser(uint(id)); err != nil {
		common.ServerError(c, "注销失败: "+err.Error())
		return
	}
	common.Success(c, gin.H{"msg": "注销成功"})
}

// verifyToken godoc
// @Summary      校验 Token 有效性
// @Tags         注册登陆
// @Produce      application/json
// @Security     ApiKeyAuth
// @Router       /auth/verify [get]
func verifyToken(c *gin.Context) {
	raw := c.GetHeader("Authorization")
	if raw == "" {
		// 也尝试从 query 参数读取（兼容前端轮询场景）
		raw = c.Query("token")
	}
	token := middleware.ExtractToken(raw)
	if token == "" {
		common.Fail(c, http.StatusUnauthorized, "token missing")
		return
	}
	claims, err := middleware.ParseToken(token)
	if err != nil || claims == nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			common.Fail(c, http.StatusUnauthorized, "token_expired")
		default:
			common.Fail(c, http.StatusUnauthorized, "token_invalid")
		}
		return
	}
	common.Success(c, gin.H{"user_id": claims.UserID, "valid": true})
}

// refreshToken godoc
// @Summary      刷新 Token（旧 Token 未过期时可刷新）
// @Tags         注册登陆
// @Produce      application/json
// @Security     ApiKeyAuth
// @Router       /auth/refresh [post]
func refreshToken(c *gin.Context) {
	raw := c.GetHeader("Authorization")
	token := middleware.ExtractToken(raw)
	if token == "" {
		common.Fail(c, http.StatusUnauthorized, "token missing")
		return
	}
	claims, err := middleware.ParseToken(token)
	if err != nil || claims == nil || claims.UserID == 0 {
		common.Fail(c, http.StatusUnauthorized, "token invalid")
		return
	}
	newToken, err := middleware.GenerateToken(claims.UserID, claims.Issuer)
	if err != nil {
		zap.S().Errorf("refresh token: %v", err)
		common.ServerError(c, "刷新 token 失败")
		return
	}
	common.Success(c, gin.H{"token": newToken})
}
