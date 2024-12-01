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

func GetUsers(ctx *gin.Context) {
	list, err := dao.GetUserList()
	if err != nil {
		models.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	models.Success(ctx, list)
}

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
		"userId":  Rsp.ID,
		"token":   token,
	})
}

func NewUser(ctx *gin.Context) {
	user := models.UserBasic{}
	user.Name = ctx.Request.FormValue("name")
	password := ctx.Request.FormValue("password")
	repassword := ctx.Request.FormValue("repeat_password")
	phone := ctx.Request.FormValue("phone")
	email := ctx.Request.FormValue("email")

	if user.Name == "" || password == "" || repassword == "" {
		models.Error(ctx, http.StatusUnauthorized, "用户名或密码不能为空")
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
	_, _ = dao.CreateUser(user)
	models.Success(ctx, gin.H{
		"message": "新增用户成功！",
		"data":    user,
	})
}

func UpdateUser(ctx *gin.Context) {
	user := models.UserBasic{}
	id, err := strconv.Atoi(ctx.Request.FormValue("id"))
	if err != nil {
		zap.S().Info("类型转换失败", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "注销账号失败",
		})
		return
	}
	user.ID = uint(id)
	Name := ctx.Request.FormValue("name")
	Password := ctx.Request.FormValue("password")
	Email := ctx.Request.FormValue("email")
	Phone := ctx.Request.FormValue("phone")
	avatar := ctx.Request.FormValue("icon")
	gender := ctx.Request.FormValue("gender")
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

	_, err = govalidator.ValidateStruct(user)
	if err != nil {
		zap.S().Info("参数不匹配", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    -1, //  0成功   -1失败
			"message": "参数不匹配",
		})
		return
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

func DeleteUser(ctx *gin.Context) {
	zap.S().Info(ctx.Request.URL, ctx.Request.Method)
	if ctx.Request.Method != http.MethodDelete {
		models.Error(ctx, http.StatusMethodNotAllowed, "Method not allowed, only DELETE is accepted")
		return
	}
	user := models.UserBasic{}
	id, err := strconv.Atoi(ctx.Request.FormValue("id"))
	if err != nil {
		zap.S().Info("类型转换失败", err)
		models.Error(ctx, http.StatusInternalServerError, "注销账号失败")
		return
	}

	user.ID = uint(id)
	err = dao.DeleteUser(user)
	if err != nil {
		zap.S().Info("注销用户失败", err)
		models.Error(ctx, http.StatusInternalServerError, "注销账号失败")
		return
	}

	models.Success(ctx, gin.H{"msg": "注销帐户成功!"})
}
