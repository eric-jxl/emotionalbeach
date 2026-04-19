// Package dao provides the unified data-access layer for the application.
// A single Dao interface aggregates all persistence operations, and the
// concrete dao struct implements it via GORM — mirroring the pattern in
// arip-samp/internal/dao/dao.go.
package dao

import (
	"emotionalBeach/internal/models"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// Provider is the Wire provider set for the DAO layer.
// It exposes New and binds *dao → Dao so callers only depend on the interface.
var Provider = wire.NewSet(
	New,
	wire.Bind(new(Dao), new(*dao)),
)

// Dao is the unified data-access interface for all persistence operations.
// Grouping every query behind a single interface keeps Wire injection flat,
// makes service constructors simple, and lets tests swap in one mock.
type Dao interface {
	ListUsers() ([]models.UserBasic, error)
	FindUserByID(id uint) (*models.UserBasic, error)
	FindUserByName(name string) (*models.UserBasic, error)
	FindUserByPhone(phone string) (*models.UserBasic, error)
	FindUserByEmail(email string) (*models.UserBasic, error)
	UserNameExists(name string) bool
	CreateUser(user models.UserBasic) (*models.UserBasic, error)
	UpdateUser(user models.UserBasic) (*models.UserBasic, error)
	DeleteUser(id uint) error
	UpdateIdentity(id uint, identity string) error
	UpdatePassword(id uint, hashed, salt string) error

	FriendList(userID uint) ([]models.UserBasic, error)
	FriendExists(ownerID, targetID uint) bool
	CreateFriendship(ownerID, targetID uint) error
}

// dao is the concrete GORM-backed implementation of Dao.
type dao struct {
	db *gorm.DB
}

// New constructs a dao and returns it as the Dao interface.
func New(db *gorm.DB) *dao {
	return &dao{db: db}
}
