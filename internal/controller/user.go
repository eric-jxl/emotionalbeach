package controller

import (
	"emotionalBeach/internal/common"
	"emotionalBeach/internal/dao"
	"emotionalBeach/internal/global"
	"emotionalBeach/internal/middleware"
	"emotionalBeach/internal/models"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetUsers
// @Summary 获取所有用户
// @Description 批量获取所有用户信息
// @Tags 用户
// @Produce json
// @Security ApiKeyAuth
// @Router /v1/user/list [get]
func GetUsers(ctx *gin.Context) {
	list, err := dao.GetUserList()
	if err != nil {
		global.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	infos := make([]userStruct, 0)

	for _, v := range list {
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
	global.Success(ctx, infos)
}

// findUser 根据给定的条件查找用户
func findUser(idStr, email, phone string) (*models.UserBasic, error) {
	if idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return nil, fmt.Errorf("非法用户ID: %v", err)
		}
		return dao.FindUserID(uint(id))
	} else if email != "" {
		return dao.FindUerByEmail(email)
	} else if phone != "" {
		return dao.FindUserByPhone(phone)
	}
	return nil, fmt.Errorf("至少需要一个参数（id、电子邮件或电话）")
}

// GetAppointUser
// @Summary 获取所有用户
// @Description 批量获取所有用户信息
// @Tags 用户
// @Param id query string false "ID"
// @Param phone query string false "手机号"
// @Param email query string false "Email"
// @Produce json
//
//	models.Resp "请求成功"
//
// @Security ApiKeyAuth
// @Router /v1/user/condition [get]
func GetAppointUser(ctx *gin.Context) {
	phone := ctx.Query("phone")
	email := ctx.Query("email")
	id := ctx.Query("id")
	if id == "" && email == "" && phone == "" {
		global.Error(ctx, http.StatusBadRequest, "至少需要一个参数（id、电子邮件或电话）")
		return
	}
	userBasic, err := findUser(id, email, phone)
	if err != nil {
		global.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	info := userStruct{
		Name:     userBasic.Name,
		Avatar:   userBasic.Avatar,
		Gender:   userBasic.Gender,
		Phone:    userBasic.Phone,
		Email:    userBasic.Email,
		Identity: userBasic.Identity,
	}
	global.Success(ctx, info)
}

// LoginByNameAndPassWord 登陆
// @Summary 登陆获取Token
// @Description 根据用户名、密码获取授权码
// @Tags 注册登陆
// @Accept application/json
// @Produce application/json
// @Param req body models.LoginRequest true "登录参数"
// @Router /login [post]
func LoginByNameAndPassWord(ctx *gin.Context) {
	req := new(models.LoginRequest)
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request: " + err.Error(),
		})
		return
	}
	data, err := dao.FindUserByName(req.Username)
	if data.Name == "" {
		global.Error(ctx, http.StatusNotFound, "用户名不存在")
		return
	}

	if err != nil {
		global.Error(ctx, http.StatusForbidden, "登录失败")
		return
	}

	//由于数据库密码保存是使用md5密文的
	ok := common.CheckPassWord(req.Password, data.Salt, data.Password)
	if !ok {
		global.Error(ctx, http.StatusUnauthorized, "密码错误")
		return
	}

	Rsp, err1 := dao.FindUserByNameAndPwd(req.Username, data.Password)
	if err1 != nil {
		zap.S().Info("登录失败", err)
		global.Error(ctx, http.StatusNotFound, "用户不存在")
		return
	}

	token, err2 := middleware.GenerateToken(Rsp.ID, Rsp.Name)
	if err2 != nil {
		zap.S().Info("生成token失败", err)
		return
	}
	global.Success(ctx, gin.H{
		"message": "登录成功",
		"user_id": Rsp.ID,
		"token":   token,
	})
}

// NewUser 登陆注册
// @Summary 创建用户
// @Description 根据名称、密码、二次密码、手机号、邮箱(可选)注册
// @Tags 注册登陆
// @Accept multipart/form-data
// @Produce application/json
// @Param name formData string true "Name"
// @Param password formData string true "Password"
// @Param repeat_password formData string true "repeat_password"
// @Param phone formData string true "Phone"
// @Param email formData string true  "EMAIL"
// @Router /register [post]
func NewUser(ctx *gin.Context) {
	user := models.UserBasic{}
	user.Name = ctx.Request.FormValue("name")
	password := ctx.Request.FormValue("password")
	repassword := ctx.Request.FormValue("repeat_password")
	phone := ctx.Request.FormValue("phone")
	email := ctx.Request.FormValue("email")

	if user.Name == "" || password == "" || repassword == "" {
		global.Error(ctx, http.StatusUnauthorized, "用户名或密码或确认密码不能为空！")
		return
	}

	//查询用户是否存在
	_, err := dao.FindUser(user.Name)
	if err != nil {
		global.Error(ctx, http.StatusUnauthorized, "该用户已注册")
		return
	}

	if password != repassword {
		global.Error(ctx, http.StatusUnauthorized, "两次密码不一致")
		return
	}

	if phone == "" {
		global.Error(ctx, http.StatusUnauthorized, "手机号不能为空!")
		return
	}
	if len(phone) != 11 {
		global.Error(ctx, http.StatusUnauthorized, "手机号必须为11位!")
		return
	}

	if !common.IsValidPhoneNumber(phone) {
		global.Error(ctx, http.StatusForbidden, "手机号非法")
		return
	}

	//生成盐值
	salt := fmt.Sprintf("%d", rand.Int31())

	//加密密码
	user.Password = common.SaltPassWord(password, salt)
	user.Salt = salt
	t := time.Now()
	user.LoginTime = &t
	user.LoginOutTime = &t
	user.HeartBeatTime = &t
	user.Phone = phone
	if email != "" {
		user.Email = email
	}
	userStruct, errs := dao.CreateUser(user)
	if errs != nil {
		global.Error(ctx, http.StatusInternalServerError, errs.Error())
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"data":    userStruct,
		"message": "新增用户成功！",
	})
}

// UpdateUser
// @Summary 更新用户信息
// @Description 更新用户信息
// @Tags 用户
// @Param id formData string true "ID"
// @Param name formData string false "用户名"
// @Param password formData string false "密码"
// @Param phone formData string false "手机号"
// @Param email formData string false "Email"
// @Param avatar formData string false "头像"
// @Param gender formData string false "性别"
// @Produce json
// @Security ApiKeyAuth
// @Router /v1/user/update [post]
func UpdateUser(ctx *gin.Context) {
	user := models.UserBasic{}
	idStr := ctx.PostForm("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		zap.S().Info("类型转换失败", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "注销账号失败",
		})
		return
	}
	user.ID = uint(id)
	Name := ctx.PostForm("name")
	Password := ctx.PostForm("password")
	Email := ctx.PostForm("email")
	Phone := ctx.PostForm("phone")
	avatar := ctx.PostForm("avatar")
	gender := ctx.PostForm("gender")
	if Name != "" {
		user.Name = Name
	}
	if Password != "" {
		salt := fmt.Sprintf("%d", rand.Int31())
		user.Salt = salt
		user.Password = common.SaltPassWord(Password, salt)
	}
	if Email != "" {
		user.Email = Email
	}
	if Phone != "" {
		user.Phone = Phone
	}
	if avatar != "" {
		user.Avatar = avatar
	}
	if gender != "" {
		user.Gender = gender
	}

	Rsp, err := dao.UpdateUser(user)
	if err != nil {
		zap.S().Info("更新用户失败", err)
		global.Error(ctx, http.StatusInternalServerError, "修改信息失败"+err.Error())
		return
	}

	global.Success(ctx, gin.H{
		"uid": Rsp.ID,
	})
}

// DeleteUser
// @Summary 删除用户
// @Description 删除用户
// @Tags 用户
// @Param id query uint true "ID"
// @Produce json
// @Security ApiKeyAuth
// @Router /v1/user/delete [delete]
func DeleteUser(ctx *gin.Context) {
	if ctx.Request.Method != http.MethodDelete {
		global.Error(ctx, http.StatusMethodNotAllowed, "Method not allowed, only DELETE is accepted")
		return
	}
	user := models.UserBasic{}
	idStr := ctx.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		zap.S().Info("类型转换失败", err)
		global.Error(ctx, http.StatusInternalServerError, "注销账号失败")
		return
	}

	user.ID = uint(id)
	err = dao.DeleteUser(uint(id))
	if err != nil {
		zap.S().Info("注销用户失败", err)
		global.Error(ctx, http.StatusInternalServerError, "注销账号失败")
		return
	}

	global.Success(ctx, gin.H{"msg": "注销帐户成功!"})
}
