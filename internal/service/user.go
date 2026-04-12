package service

import (
	"emotionalBeach/internal/common"
	ebmetrics "emotionalBeach/internal/infra"
	"emotionalBeach/internal/models"
	"errors"
	"fmt"
	"math/rand"
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
func (s *Service) Login(username, password string) (*models.UserBasic, error) {
	user, err := s.dao.FindUserByName(username)
	if err != nil || user == nil {
		ebmetrics.UserLoginsTotal.WithLabelValues("not_found").Inc()
		return nil, errors.New("user not found")
	}
	if !common.CheckPassWord(password, user.Salt, user.Password) {
		ebmetrics.UserLoginsTotal.WithLabelValues("wrong_password").Inc()
		return nil, errors.New("invalid password")
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

// Register creates a new user after validating uniqueness and hashing the password.
func (s *Service) Register(name, password, phone, email string) (*models.UserBasic, error) {
	if s.dao.UserNameExists(name) {
		return nil, errors.New("username already taken")
	}
	salt := strconv.Itoa(int(rand.Int31()))
	now := time.Now()
	user := models.UserBasic{
		Name:          name,
		Password:      common.SaltPassWord(password, salt),
		Salt:          salt,
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
		salt := fmt.Sprintf("%d", rand.Int31())
		user.Salt = salt
		user.Password = common.SaltPassWord(req.Password, salt)
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

