// Package userrepo implements the user.Repository interface using GORM.
package userrepo

import (
	"emotionalBeach/internal/common"
	userdomain "emotionalBeach/internal/domain/user"
	"emotionalBeach/internal/models"
	"errors"
	"strconv"
	"time"

	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Set is the Wire provider set for the user repository.
// It declares GormRepo as the concrete implementation of userdomain.Repository.
var Set = wire.NewSet(
	NewGormRepo,
	wire.Bind(new(userdomain.Repository), new(*GormRepo)),
)

// GormRepo implements userdomain.Repository via a GORM *DB.
type GormRepo struct {
	db *gorm.DB
}

// NewGormRepo constructs a GormRepo.
func NewGormRepo(db *gorm.DB) *GormRepo {
	return &GormRepo{db: db}
}

func (r *GormRepo) List() ([]models.UserBasic, error) {
	var list []models.UserBasic
	if tx := r.db.Order("id").Find(&list); tx.RowsAffected == 0 {
		return nil, errors.New("user list is empty")
	}
	return list, nil
}

func (r *GormRepo) FindByID(id uint) (*models.UserBasic, error) {
	var u models.UserBasic
	if tx := r.db.Where("id = ?", id).First(&u); tx.RowsAffected == 0 {
		return nil, errors.New("user not found")
	}
	return &u, nil
}

func (r *GormRepo) FindByName(name string) (*models.UserBasic, error) {
	var u models.UserBasic
	if tx := r.db.Where("name = ?", name).First(&u); tx.RowsAffected == 0 {
		return nil, errors.New("user not found")
	}
	return &u, nil
}

func (r *GormRepo) FindByPhone(phone string) (*models.UserBasic, error) {
	var u models.UserBasic
	if tx := r.db.Where("phone = ?", phone).First(&u); tx.RowsAffected == 0 {
		return nil, errors.New("user not found")
	}
	return &u, nil
}

func (r *GormRepo) FindByEmail(email string) (*models.UserBasic, error) {
	var u models.UserBasic
	if tx := r.db.Where("email = ?", email).First(&u); tx.RowsAffected == 0 {
		return nil, errors.New("user not found")
	}
	return &u, nil
}

// FindByNameAndPwd verifies credentials and refreshes the identity token on success.
func (r *GormRepo) FindByNameAndPwd(name, password string) (*models.UserBasic, error) {
	var u models.UserBasic
	if tx := r.db.Where("name = ? AND password = ?", name, password).First(&u); tx.RowsAffected == 0 {
		return nil, errors.New("invalid credentials")
	}
	identity := common.Md5encoder(strconv.Itoa(int(time.Now().Unix())))
	r.db.Model(&u).Where("id = ?", u.ID).Update("identity", identity)
	return &u, nil
}

func (r *GormRepo) NameExists(name string) bool {
	var u models.UserBasic
	return r.db.Where("name = ?", name).First(&u).RowsAffected == 1
}

func (r *GormRepo) Create(user models.UserBasic) (*models.UserBasic, error) {
	if tx := r.db.Create(&user); tx.RowsAffected == 0 {
		zap.S().Warn("create user: no rows affected")
		return nil, errors.New("create user failed")
	}
	return &user, nil
}

func (r *GormRepo) Update(user models.UserBasic) (*models.UserBasic, error) {
	tx := r.db.Model(&user).Updates(models.UserBasic{
		Name:     user.Name,
		Password: user.Password,
		Gender:   user.Gender,
		Phone:    user.Phone,
		Email:    user.Email,
		Avatar:   user.Avatar,
		Salt:     user.Salt,
	})
	if tx.RowsAffected == 0 {
		return nil, errors.New("update user failed")
	}
	return &user, nil
}

func (r *GormRepo) Delete(id uint) error {
	var u models.UserBasic
	tx := r.db.Model(&u).Where("id = ?", id).Delete(&u)
	if tx.RowsAffected == 0 {
		return errors.New("delete user failed")
	}
	return nil
}

