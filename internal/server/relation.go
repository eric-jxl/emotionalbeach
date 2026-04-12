package server

import (
	"emotionalBeach/internal/common"
	"emotionalBeach/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type friendView struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Gender string `json:"gender"`
}

func toFriendView(u models.UserBasic) friendView {
	return friendView{
		ID: u.ID, Name: u.Name, Avatar: u.Avatar, Gender: u.Gender,
	}
}

// friendList godoc
// @Summary      好友列表
// @Tags         好友关系
// @Param        userId formData uint false "用户 ID"
// @Produce      json
// @Security     ApiKeyAuth
// @Router       /v1/relation/list [post]
func friendList(c *gin.Context) {
	id, err := strconv.Atoi(c.PostForm("userId"))
	if err != nil {
		common.Fail(c, http.StatusBadRequest, "非法用户 ID")
		return
	}
	users, err := svc.FriendList(uint(id))
	if err != nil {
		zap.S().Infow("FriendList failed", "userID", id, "error", err)
		common.Fail(c, http.StatusNotFound, "好友为空")
		return
	}
	views := make([]friendView, 0, len(users))
	for _, u := range users {
		views = append(views, toFriendView(u))
	}
	common.Success(c, gin.H{"users": views, "count": len(views)})
}

// addFriendByName godoc
// @Summary      通过昵称加好友
// @Tags         好友关系
// @Param        userId     formData uint   false "发起方用户 ID"
// @Param        targetName formData string false "目标昵称或目标用户 ID"
// @Produce      json
// @Security     ApiKeyAuth
// @Router       /v1/relation/add [post]
func addFriendByName(c *gin.Context) {
	userID, err := strconv.Atoi(c.PostForm("userId"))
	if err != nil {
		zap.S().Infow("addFriendByName: bad userId", "raw", c.PostForm("userId"))
		common.Fail(c, http.StatusBadRequest, "非法用户 ID")
		return
	}
	targetRaw := c.PostForm("targetName")
	if targetID, convErr := strconv.Atoi(targetRaw); convErr == nil {
		err = svc.AddFriend(uint(userID), uint(targetID))
	} else {
		err = svc.AddFriendByName(uint(userID), targetRaw)
	}
	if err != nil {
		switch err.Error() {
		case "friendship already exists":
			common.Fail(c, http.StatusConflict, "该好友已经存在")
		case "cannot add yourself as a friend":
			common.Fail(c, http.StatusBadRequest, "不能添加自己")
		default:
			common.ServerError(c, err.Error())
		}
		return
	}
	common.Success(c, gin.H{"msg": "添加好友成功"})
}
