package dao

import (
	"emotionalBeach/internal/models"
	"errors"

	"go.uber.org/zap"
)

func (d *dao) ListUsers() ([]models.UserBasic, error) {
	var list []models.UserBasic
	if tx := d.db.Order("id").Find(&list); tx.RowsAffected == 0 {
		return nil, errors.New("user list is empty")
	}
	return list, nil
}

func (d *dao) FindUserByID(id uint) (*models.UserBasic, error) {
	var u models.UserBasic
	if tx := d.db.Where("id = ?", id).First(&u); tx.RowsAffected == 0 {
		return nil, errors.New("user not found")
	}
	return &u, nil
}

func (d *dao) FindUserByName(name string) (*models.UserBasic, error) {
	var u models.UserBasic
	if tx := d.db.Where("name = ?", name).First(&u); tx.RowsAffected == 0 {
		return nil, errors.New("user not found")
	}
	return &u, nil
}

func (d *dao) FindUserByPhone(phone string) (*models.UserBasic, error) {
	var u models.UserBasic
	if tx := d.db.Where("phone = ?", phone).First(&u); tx.RowsAffected == 0 {
		return nil, errors.New("user not found")
	}
	return &u, nil
}

func (d *dao) FindUserByEmail(email string) (*models.UserBasic, error) {
	var u models.UserBasic
	if tx := d.db.Where("email = ?", email).First(&u); tx.RowsAffected == 0 {
		return nil, errors.New("user not found")
	}
	return &u, nil
}

func (d *dao) UserNameExists(name string) bool {
	var u models.UserBasic
	return d.db.Where("name = ?", name).First(&u).RowsAffected == 1
}

func (d *dao) CreateUser(user models.UserBasic) (*models.UserBasic, error) {
	if tx := d.db.Create(&user); tx.RowsAffected == 0 {
		zap.S().Warn("CreateUser: no rows affected")
		return nil, errors.New("create user failed")
	}
	return &user, nil
}

func (d *dao) UpdateUser(user models.UserBasic) (*models.UserBasic, error) {
	tx := d.db.Model(&user).Updates(models.UserBasic{
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

func (d *dao) DeleteUser(id uint) error {
	var u models.UserBasic
	if tx := d.db.Model(&u).Where("id = ?", id).Delete(&u); tx.RowsAffected == 0 {
		return errors.New("delete user failed")
	}
	return nil
}

func (d *dao) UpdateIdentity(id uint, identity string) error {
	if tx := d.db.Model(&models.UserBasic{}).Where("id = ?", id).Update("identity", identity); tx.RowsAffected == 0 {
		return errors.New("update identity failed")
	}
	return nil
}

func (d *dao) UpdatePassword(id uint, hashed, salt string) error {
	updates := map[string]interface{}{"password": hashed, "salt": salt}
	if tx := d.db.Model(&models.UserBasic{}).Where("id = ?", id).Updates(updates); tx.Error != nil {
		return tx.Error
	}
	return nil
}

