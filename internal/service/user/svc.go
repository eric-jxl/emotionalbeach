// Package usersvc implements the user.Service business-logic contract.
package usersvc

import (
	"emotionalBeach/internal/common"
	userdomain "emotionalBeach/internal/domain/user"
	ebmetrics "emotionalBeach/internal/infra/metrics"
	"emotionalBeach/internal/models"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/google/wire"
)

// Set is the Wire provider set for the user service.
var Set = wire.NewSet(
	NewSvc,
	wire.Bind(new(userdomain.Service), new(*Svc)),
)

// Svc implements userdomain.Service.
type Svc struct {
	repo userdomain.Repository
}

// NewSvc constructs a Svc with its repository dependency.
func NewSvc(repo userdomain.Repository) *Svc {
	return &Svc{repo: repo}
}

func (s *Svc) GetList() ([]models.UserBasic, error) {
	return s.repo.List()
}

func (s *Svc) GetByID(id uint) (*models.UserBasic, error) {
	return s.repo.FindByID(id)
}

func (s *Svc) GetByPhone(phone string) (*models.UserBasic, error) {
	return s.repo.FindByPhone(phone)
}

func (s *Svc) GetByEmail(email string) (*models.UserBasic, error) {
	return s.repo.FindByEmail(email)
}

// Login validates the username/password pair and returns the authenticated user.
func (s *Svc) Login(username, password string) (*models.UserBasic, error) {
	// Step 1: fetch the user so we can read the salt.
	user, err := s.repo.FindByName(username)
	if err != nil || user == nil {
		ebmetrics.UserLoginsTotal.WithLabelValues("not_found").Inc()
		return nil, errors.New("user not found")
	}
	// Step 2: compare salted passwords.
	if !common.CheckPassWord(password, user.Salt, user.Password) {
		ebmetrics.UserLoginsTotal.WithLabelValues("wrong_password").Inc()
		return nil, errors.New("invalid password")
	}
	// Step 3: refresh identity token and return authenticated user.
	result, err := s.repo.FindByNameAndPwd(username, user.Password)
	if err != nil {
		ebmetrics.UserLoginsTotal.WithLabelValues("token_error").Inc()
		return nil, err
	}
	ebmetrics.UserLoginsTotal.WithLabelValues("success").Inc()
	return result, nil
}

// Register creates a new user after validating uniqueness and hashing the password.
func (s *Svc) Register(name, password, phone, email string) (*models.UserBasic, error) {
	if s.repo.NameExists(name) {
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
	created, err := s.repo.Create(user)
	if err != nil {
		return nil, err
	}
	ebmetrics.UserRegistrationsTotal.Inc()
	return created, nil
}

// Update applies non-zero fields from req to the stored user.
func (s *Svc) Update(req userdomain.UpdateRequest) (*models.UserBasic, error) {
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
	return s.repo.Update(user)
}

func (s *Svc) Delete(id uint) error {
	return s.repo.Delete(id)
}

