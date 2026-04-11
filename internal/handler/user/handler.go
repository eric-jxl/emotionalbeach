// Package userhandler wires the HTTP presentation layer for user operations.
package userhandler

import (
	"emotionalBeach/internal/common"
	userdomain "emotionalBeach/internal/domain/user"
	"emotionalBeach/internal/global"
	"emotionalBeach/internal/middleware"
	"emotionalBeach/internal/models"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// userView is the outbound DTO — hides sensitive fields like password/salt.
type userView struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
	Gender   string `json:"gender"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Identity string `json:"identity"`
}

func toView(u models.UserBasic) userView {
	return userView{
		ID:       u.ID,
		Name:     u.Name,
		Avatar:   u.Avatar,
		Gender:   u.Gender,
		Phone:    u.Phone,
		Email:    u.Email,
		Identity: u.Identity,
	}
}

// Handler holds injected service dependencies for user routes.
type Handler struct {
	svc userdomain.Service
}

// NewHandler constructs a Handler.
func NewHandler(svc userdomain.Service) *Handler {
	return &Handler{svc: svc}
}

// GetUsers godoc
// @Summary      获取所有用户
// @Description  批量获取所有用户信息
// @Tags         用户
// @Produce      json
// @Security     ApiKeyAuth
// @Router       /v1/user/list [get]
func (h *Handler) GetUsers(c *gin.Context) {
	list, err := h.svc.GetList()
	if err != nil {
		global.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	views := make([]userView, 0, len(list))
	for _, u := range list {
		views = append(views, toView(u))
	}
	global.Success(c, views)
}

// GetAppointUser godoc
// @Summary      按条件查询用户
// @Description  通过 id / phone / email 查询单个用户
// @Tags         用户
// @Param        id     query string false "用户 ID"
// @Param        phone  query string false "手机号"
// @Param        email  query string false "Email"
// @Produce      json
// @Security     ApiKeyAuth
// @Router       /v1/user/condition [get]
func (h *Handler) GetAppointUser(c *gin.Context) {
	id, phone, email := c.Query("id"), c.Query("phone"), c.Query("email")
	if id == "" && phone == "" && email == "" {
		global.Error(c, http.StatusBadRequest, "至少需要一个参数（id、phone 或 email）")
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
			global.Error(c, http.StatusBadRequest, fmt.Sprintf("非法用户 ID: %v", convErr))
			return
		}
		u, err = h.svc.GetByID(uint(uid))
	case email != "":
		u, err = h.svc.GetByEmail(email)
	default:
		u, err = h.svc.GetByPhone(phone)
	}
	if err != nil {
		global.Error(c, http.StatusNotFound, err.Error())
		return
	}
	global.Success(c, toView(*u))
}

// Login godoc
// @Summary      登陆获取 Token
// @Description  根据用户名、密码获取授权 JWT
// @Tags         注册登陆
// @Accept       application/json
// @Produce      application/json
// @Param        req body models.LoginRequest true "登录参数"
// @Router       /login [post]
func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}
	authUser, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		global.Error(c, http.StatusUnauthorized, err.Error())
		return
	}
	token, err := middleware.GenerateToken(authUser.ID, authUser.Name)
	if err != nil {
		zap.S().Errorf("generate token: %v", err)
		global.Error(c, http.StatusInternalServerError, "生成 token 失败")
		return
	}
	global.Success(c, gin.H{"message": "登录成功", "user_id": authUser.ID, "token": token})
}

// Register godoc
// @Summary      创建用户
// @Description  根据名称、密码、手机号、邮箱注册
// @Tags         注册登陆
// @Accept       multipart/form-data
// @Produce      application/json
// @Param        name            formData string true  "Name"
// @Param        password        formData string true  "Password"
// @Param        repeat_password formData string true  "repeat_password"
// @Param        phone           formData string true  "Phone"
// @Param        email           formData string false "EMAIL"
// @Router       /register [post]
func (h *Handler) Register(c *gin.Context) {
	name := c.PostForm("name")
	password := c.PostForm("password")
	repeatPwd := c.PostForm("repeat_password")
	phone := c.PostForm("phone")
	email := c.PostForm("email")

	switch {
	case name == "" || password == "" || repeatPwd == "":
		global.Error(c, http.StatusBadRequest, "用户名、密码、确认密码不能为空")
		return
	case password != repeatPwd:
		global.Error(c, http.StatusBadRequest, "两次密码不一致")
		return
	case phone == "":
		global.Error(c, http.StatusBadRequest, "手机号不能为空")
		return
	case len(phone) != 11:
		global.Error(c, http.StatusBadRequest, "手机号必须为 11 位")
		return
	case !common.IsValidPhoneNumber(phone):
		global.Error(c, http.StatusForbidden, "手机号格式非法")
		return
	}

	created, err := h.svc.Register(name, password, phone, email)
	if err != nil {
		global.Error(c, http.StatusConflict, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "data": created, "message": "注册成功"})
}

// UpdateUser godoc
// @Summary      更新用户信息
// @Description  更新用户信息（仅更新非空字段）
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
func (h *Handler) UpdateUser(c *gin.Context) {
	idStr := c.PostForm("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		global.Error(c, http.StatusBadRequest, "非法用户 ID")
		return
	}
	req := userdomain.UpdateRequest{
		ID:       uint(id),
		Name:     c.PostForm("name"),
		Password: c.PostForm("password"),
		Email:    c.PostForm("email"),
		Phone:    c.PostForm("phone"),
		Avatar:   c.PostForm("avatar"),
		Gender:   c.PostForm("gender"),
	}
	updated, err := h.svc.Update(req)
	if err != nil {
		global.Error(c, http.StatusInternalServerError, "修改信息失败: "+err.Error())
		return
	}
	global.Success(c, gin.H{"uid": updated.ID})
}

// DeleteUser godoc
// @Summary      删除用户
// @Description  软删除用户（需 DELETE 方法）
// @Tags         用户
// @Param        id query uint true "ID"
// @Produce      json
// @Security     ApiKeyAuth
// @Router       /v1/user/delete [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		global.Error(c, http.StatusBadRequest, "非法用户 ID")
		return
	}
	if err = h.svc.Delete(uint(id)); err != nil {
		global.Error(c, http.StatusInternalServerError, "注销失败: "+err.Error())
		return
	}
	global.Success(c, gin.H{"msg": "注销成功"})
}
