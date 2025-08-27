package controller

import (
	"emotionalBeach/dao"
	"emotionalBeach/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// user对返回数据进行屏蔽
type userStruct struct {
	Name     string
	Role     string
	Avatar   string
	Gender   string
	Phone    string
	Email    string
	Identity string
}

// FriendList GetAppointUser
// @Summary 获取好友列表
// @Description 批量获取好友列表信息
// @Tags 好友关系
// @Param userId formData uint true "好友ID"
// @Produce json
// @Security ApiKeyAuth
// @Router /v1/relation/list [post]
func FriendList(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.PostForm("userId"))
	users, err := dao.FriendList(uint(id))
	if err != nil {
		zap.S().Info("获取好友列表失败", err)
		models.Error(ctx, http.StatusNotFound, "好友为空")
		return
	}

	infos := make([]userStruct, 0)

	for _, v := range *users {
		info := userStruct{
			Name:     v.Name,
			Avatar:   v.Avatar,
			Gender:   v.Gender,
			Phone:    v.Phone,
			Email:    v.Email,
			Identity: v.Identity,
		}
		infos = append(infos, info)
	}
	models.Success(ctx, gin.H{"users": infos, "count": len(infos)})
}

// AddFriendByName 通过昵称加好友
// @Summary 通过昵称加好友
// @Description 通过昵称加好友
// @Tags 好友关系
// @Param userId formData uint false "增加的用户id"
// @Produce json
// @Security ApiKeyAuth
// @Router /v1/relation/add [post]
func AddFriendByName(ctx *gin.Context) {
	user := ctx.PostForm("userId")
	userId, err := strconv.Atoi(user)
	if err != nil {
		zap.S().Info("类型转换失败", err)
		return
	}

	tar := ctx.PostForm("targetName")
	target, err := strconv.Atoi(tar)
	if err != nil {
		code, err := dao.AddFriendByName(uint(userId), tar)
		if err != nil {
			HandleErr(code, ctx, err)
			return
		}

	} else {
		code, err := dao.AddFriend(uint(userId), uint(target))
		if err != nil {
			HandleErr(code, ctx, err)
			return
		}
	}
	models.Success(ctx, gin.H{"msg": "添加好友成功"})

}

func HandleErr(code int, ctx *gin.Context, err error) {
	switch code {
	case -1:
		ctx.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": err.Error(),
		})
	case 0:
		ctx.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "该好友已经存在",
		})
	case -2:
		ctx.JSON(200, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "不能添加自己",
		})

	}
}
