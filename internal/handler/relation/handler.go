// Package relationhandler wires the HTTP presentation layer for relation operations.
package relationhandler

import (
	reldomain "emotionalBeach/internal/domain/relation"
	"emotionalBeach/internal/global"
	"emotionalBeach/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// friendView is the outbound DTO for a friend entry.
type friendView struct {
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
	Gender   string `json:"gender"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Identity string `json:"identity"`
}

func toFriendView(u models.UserBasic) friendView {
	return friendView{
		Name:     u.Name,
		Avatar:   u.Avatar,
		Gender:   u.Gender,
		Phone:    u.Phone,
		Email:    u.Email,
		Identity: u.Identity,
	}
}

// Handler holds injected service dependencies for relation routes.
type Handler struct {
	svc reldomain.Service
}

// NewHandler constructs a Handler.
func NewHandler(svc reldomain.Service) *Handler {
	return &Handler{svc: svc}
}

// FriendList godoc
// @Summary      获取好友列表
// @Description  批量获取好友列表信息
// @Tags         好友关系
// @Param        userId formData uint true "用户 ID"
// @Produce      json
// @Security     ApiKeyAuth
// @Router       /v1/relation/list [post]
func (h *Handler) FriendList(c *gin.Context) {
	id, _ := strconv.Atoi(c.PostForm("userId"))
	users, err := h.svc.FriendList(uint(id))
	if err != nil {
		zap.S().Infow("FriendList failed", "userID", id, "error", err)
		global.Error(c, http.StatusNotFound, "好友为空")
		return
	}
	views := make([]friendView, 0, len(users))
	for _, u := range users {
		views = append(views, toFriendView(u))
	}
	global.Success(c, gin.H{"users": views, "count": len(views)})
}

// AddFriendByName godoc
// @Summary      通过昵称加好友
// @Description  通过 userId + targetName（昵称）或 targetName（纯数字当 ID）添加好友
// @Tags         好友关系
// @Param        userId     formData uint   false "发起方用户 ID"
// @Param        targetName formData string false "目标昵称或目标用户 ID"
// @Produce      json
// @Security     ApiKeyAuth
// @Router       /v1/relation/add [post]
func (h *Handler) AddFriendByName(c *gin.Context) {
	userIDStr := c.PostForm("userId")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		zap.S().Infow("AddFriendByName: bad userId", "raw", userIDStr)
		global.Error(c, http.StatusBadRequest, "非法用户 ID")
		return
	}

	targetRaw := c.PostForm("targetName")
	// Try as numeric ID first; fall back to name lookup.
	if targetID, convErr := strconv.Atoi(targetRaw); convErr == nil {
		err = h.svc.AddFriend(uint(userID), uint(targetID))
	} else {
		err = h.svc.AddFriendByName(uint(userID), targetRaw)
	}

	if err != nil {
		switch err.Error() {
		case "friendship already exists":
			c.JSON(http.StatusOK, gin.H{"code": -1, "message": "该好友已经存在"})
		case "cannot add yourself as a friend":
			c.JSON(http.StatusOK, gin.H{"code": -1, "message": "不能添加自己"})
		default:
			c.JSON(http.StatusOK, gin.H{"code": -1, "message": err.Error()})
		}
		return
	}
	global.Success(c, gin.H{"msg": "添加好友成功"})
}

