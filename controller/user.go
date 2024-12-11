package controller

import (
	"emotionalBeach/common"
	"emotionalBeach/dao"
	"emotionalBeach/middlewear"
	"emotionalBeach/models"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
}

// GetUsers
// @Summary 获取所有用户
// @Description 批量获取所有用户信息
// @Tags 用户
// @Produce json
// @Param Uid header string true "用户身份"
// @Security ApiKeyAuth
// @Router /v1/user/list [get]
func GetUsers(ctx *gin.Context) {
	list, err := dao.GetUserList()
	if err != nil {
		models.Error(ctx, http.StatusInternalServerError, err.Error())
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
	models.Success(ctx, infos)
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
// @Param Uid header string true "用户身份"
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
		models.Error(ctx, http.StatusBadRequest, "至少需要一个参数（id、电子邮件或电话）")
		return
	}
	userBasic, err := findUser(id, email, phone)
	if err != nil {
		models.Error(ctx, http.StatusInternalServerError, err.Error())
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
	models.Success(ctx, info)
}

// LoginByNameAndPassWord 登陆
// @Summary 登陆获取Token
// @Description 根据用户名、密码获取授权码
// @Tags 注册登陆
// @Accept multipart/form-data
// @Produce application/json
// @Param name formData string true "Name"
// @Param password formData string true "Password"
// @Router /login [post]
func LoginByNameAndPassWord(ctx *gin.Context) {
	name := ctx.PostForm("name")
	password := ctx.PostForm("password")
	data, err := dao.FindUserByName(name)
	if err != nil {
		models.Error(ctx, http.StatusForbidden, "登录失败")
		return
	}

	if data.Name == "" {
		models.Error(ctx, http.StatusNotFound, "用户名不存在")
		return
	}

	//由于数据库密码保存是使用md5密文的， 所以验证密码时，是将密码再次加密，然后进行对比，后期会讲解md:common.CheckPassWord
	ok := common.CheckPassWord(password, data.Salt, data.Password)
	if !ok {
		models.Error(ctx, http.StatusUnauthorized, "密码错误")
		return
	}

	Rsp, err := dao.FindUserByNameAndPwd(name, data.Password)
	if err != nil {
		zap.S().Info("登录失败", err)
	}

	//这里使用jwt做权限认证，后面将会介绍
	token, err := middlewear.GenerateToken(Rsp.ID, Rsp.Name)
	if err != nil {
		zap.S().Info("生成token失败", err)
		return
	}
	models.Success(ctx, gin.H{
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
		models.Error(ctx, http.StatusUnauthorized, "用户名或密码或确认密码不能为空！")
		return
	}

	//查询用户是否存在
	_, err := dao.FindUser(user.Name)
	if err != nil {
		models.Error(ctx, http.StatusUnauthorized, "该用户已注册")
		return
	}

	if password != repassword {
		models.Error(ctx, http.StatusUnauthorized, "两次密码不一致")
		return
	}

	if phone == "" {
		models.Error(ctx, http.StatusUnauthorized, "手机号不能为空!")
		return
	}
	if len(phone) != 11 {
		models.Error(ctx, http.StatusUnauthorized, "手机号必须为11位!")
		return
	}

	if !common.IsValidPhoneNumber(phone) {
		models.Error(ctx, http.StatusForbidden, "手机号非法")
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
		models.Error(ctx, http.StatusInternalServerError, errs.Error())
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
// @Param Uid header string true "用户身份"
// @Param id formData string true "ID"
// @Param name formData string false "用户名"
// @Param password formData string false "密码"
// @Param phone formData string false "手机号"
// @Param email formData string false "Email"
// @Param avatar formData string false "avatar"
// @Param gender formData string false "gender"
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
		models.Error(ctx, http.StatusInternalServerError, "修改信息失败"+err.Error())
		return
	}

	models.Success(ctx, gin.H{
		"uid": Rsp.ID,
	})
}

// DeleteUser
// @Summary 更新用户信息
// @Description 更新用户信息
// @Tags 用户
// @Param Uid header string true "用户身份"
// @Param id query uint true "ID"
// @Produce json
// @Security ApiKeyAuth
// @Router /v1/user/delete [delete]
func DeleteUser(ctx *gin.Context) {
	if ctx.Request.Method != http.MethodDelete {
		models.Error(ctx, http.StatusMethodNotAllowed, "Method not allowed, only DELETE is accepted")
		return
	}
	user := models.UserBasic{}
	idStr := ctx.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		zap.S().Info("类型转换失败", err)
		models.Error(ctx, http.StatusInternalServerError, "注销账号失败")
		return
	}

	user.ID = uint(id)
	err = dao.DeleteUser(uint(id))
	if err != nil {
		zap.S().Info("注销用户失败", err)
		models.Error(ctx, http.StatusInternalServerError, "注销账号失败")
		return
	}

	models.Success(ctx, gin.H{"msg": "注销帐户成功!"})
}
