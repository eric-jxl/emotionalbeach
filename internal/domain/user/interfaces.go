// Package user defines the user domain contracts (interfaces).
package user

import "emotionalBeach/internal/models"

// Repository defines the data-access contract for user persistence.
type Repository interface {
	List() ([]models.UserBasic, error)
	FindByID(id uint) (*models.UserBasic, error)
	FindByName(name string) (*models.UserBasic, error)
	FindByPhone(phone string) (*models.UserBasic, error)
	FindByEmail(email string) (*models.UserBasic, error)
	// FindByNameAndPwd verifies credentials and refreshes the identity token.
	FindByNameAndPwd(name, password string) (*models.UserBasic, error)
	// NameExists reports whether a user with the given name is already registered.
	NameExists(name string) bool
	Create(user models.UserBasic) (*models.UserBasic, error)
	Update(user models.UserBasic) (*models.UserBasic, error)
	Delete(id uint) error
}

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

// Service defines the business-logic contract for user operations.
type Service interface {
	GetList() ([]models.UserBasic, error)
	GetByID(id uint) (*models.UserBasic, error)
	GetByPhone(phone string) (*models.UserBasic, error)
	GetByEmail(email string) (*models.UserBasic, error)
	// Login validates credentials and returns the authenticated user.
	Login(username, password string) (*models.UserBasic, error)
	Register(name, password, phone, email string) (*models.UserBasic, error)
	Update(req UpdateRequest) (*models.UserBasic, error)
	Delete(id uint) error
}

