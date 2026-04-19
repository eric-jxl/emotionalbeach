package service

import (
	"emotionalBeach/internal/common"
	ebmetrics "emotionalBeach/internal/infra"
	"emotionalBeach/internal/models"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// UpdateRequest carries the mutable fields for a user update operation.
type UpdateRequest struct {
	ID       uint
	Name     string
	Password string
	Phone    string
	Email    string
	Avatar   string
	Gender   string
}

func (s *Service) GetList() ([]models.UserBasic, error) {
	return s.dao.ListUsers()
}

func (s *Service) GetByID(id uint) (*models.UserBasic, error) {
	return s.dao.FindUserByID(id)
}

func (s *Service) GetByPhone(phone string) (*models.UserBasic, error) {
	return s.dao.FindUserByPhone(phone)
}

func (s *Service) GetByEmail(email string) (*models.UserBasic, error) {
	return s.dao.FindUserByEmail(email)
}

// Login validates the username/password pair, refreshes the identity token,
// and returns the authenticated user.
// 支持 sha256 和旧版 MD5+salt 两种格式，旧格式登录成功后自动迁移到 sha256。
func (s *Service) Login(username, password string) (*models.UserBasic, error) {
	// 先尝试按用户名查找，再尝试邮箱/手机号
	user, err := s.dao.FindUserByName(username)
	if err != nil || user == nil {
		user, err = s.dao.FindUserByEmail(username)
	}
	if err != nil || user == nil {
		user, err = s.dao.FindUserByPhone(username)
	}
	if err != nil || user == nil {
		ebmetrics.UserLoginsTotal.WithLabelValues("not_found").Inc()
		return nil, errors.New("用户不存在")
	}

	// 验证密码：sha256 优先，旧版 MD5+salt 兜底
	if common.IsSha256Hash(user.Password) {
		if !common.CheckSha256Password(password, user.Password) {
			ebmetrics.UserLoginsTotal.WithLabelValues("wrong_password").Inc()
			return nil, errors.New("密码错误")
		}
	} else {
		// 旧版 MD5+salt 校验
		if !common.CheckPassWord(password, user.Salt, user.Password) {
			ebmetrics.UserLoginsTotal.WithLabelValues("wrong_password").Inc()
			return nil, errors.New("密码错误")
		}
		// 自动迁移：将密码升级为 sha256
		if hashed, herr := common.Sha256Password(password); herr == nil {
			_ = s.dao.UpdatePassword(user.ID, hashed, "")
			user.Password = hashed
			user.Salt = ""
		}
	}

	identity := common.Md5encoder(strconv.Itoa(int(time.Now().Unix())))
	if err := s.dao.UpdateIdentity(user.ID, identity); err != nil {
		ebmetrics.UserLoginsTotal.WithLabelValues("token_error").Inc()
		return nil, err
	}
	user.Identity = identity
	ebmetrics.UserLoginsTotal.WithLabelValues("success").Inc()
	return user, nil
}

// Register creates a new user after validating uniqueness and hashing the password with sha256.
func (s *Service) Register(name, password, phone, email string) (*models.UserBasic, error) {
	if s.dao.UserNameExists(name) {
		return nil, errors.New("username already taken")
	}
	hashed, err := common.Sha256Password(password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}
	now := time.Now()
	user := models.UserBasic{
		Name:          name,
		Password:      hashed,
		Salt:          "emo", // sha256 不需要独立 salt
		Phone:         phone,
		Email:         email,
		LoginTime:     &now,
		LoginOutTime:  &now,
		HeartBeatTime: &now,
	}
	created, err := s.dao.CreateUser(user)
	if err != nil {
		return nil, err
	}
	ebmetrics.UserRegistrationsTotal.Inc()
	return created, nil
}

// Update applies non-zero fields from req to the stored user.
func (s *Service) Update(req UpdateRequest) (*models.UserBasic, error) {
	if req.ID == 0 {
		return nil, errors.New("user ID is required")
	}
	user := models.UserBasic{}
	user.ID = req.ID
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Password != "" {
		hashed, err := common.Sha256Password(req.Password)
		if err != nil {
			return nil, fmt.Errorf("密码加密失败: %w", err)
		}
		user.Salt = ""
		user.Password = hashed
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Gender != "" {
		user.Gender = req.Gender
	}
	return s.dao.UpdateUser(user)
}

func (s *Service) DeleteUser(id uint) error {
	return s.dao.DeleteUser(id)
}
