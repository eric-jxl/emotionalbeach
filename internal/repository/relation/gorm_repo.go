// Package relationrepo implements the relation.Repository interface using GORM.
package relationrepo

import (
	reldomain "emotionalBeach/internal/domain/relation"
	"emotionalBeach/internal/models"
	"errors"

	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Set is the Wire provider set for the relation repository.
var Set = wire.NewSet(
	NewGormRepo,
	wire.Bind(new(reldomain.Repository), new(*GormRepo)),
)

// GormRepo implements reldomain.Repository via a GORM *DB.
type GormRepo struct {
	db *gorm.DB
}

// NewGormRepo constructs a GormRepo.
func NewGormRepo(db *gorm.DB) *GormRepo {
	return &GormRepo{db: db}
}

// FriendList returns all UserBasic records for the friends of userID.
func (r *GormRepo) FriendList(userID uint) ([]models.UserBasic, error) {
	var relations []models.Relation
	if tx := r.db.Where("owner_id = ? AND type = 1", userID).Find(&relations); tx.RowsAffected == 0 {
		return nil, errors.New("no friends found")
	}
	ids := make([]uint, 0, len(relations))
	for _, rel := range relations {
		ids = append(ids, rel.TargetID)
	}
	var users []models.UserBasic
	if tx := r.db.Where("id IN ?", ids).Find(&users); tx.RowsAffected == 0 {
		return nil, errors.New("friend users not found")
	}
	return users, nil
}

// Exists reports whether a friendship already exists in either direction.
func (r *GormRepo) Exists(ownerID, targetID uint) bool {
	var rel models.Relation
	if r.db.Where("owner_id = ? AND target_id = ? AND type = 1", ownerID, targetID).First(&rel).RowsAffected == 1 {
		return true
	}
	return r.db.Where("owner_id = ? AND target_id = ? AND type = 1", targetID, ownerID).First(&rel).RowsAffected == 1
}

// CreateBidirectional inserts two Relation rows in a single transaction.
func (r *GormRepo) CreateBidirectional(ownerID, targetID uint) error {
	tx := r.db.Begin()
	defer func() {
		if rec := recover(); rec != nil {
			tx.Rollback()
		}
	}()

	rows := []models.Relation{
		{OwnerId: ownerID, TargetID: targetID, Type: 1},
		{OwnerId: targetID, TargetID: ownerID, Type: 1},
	}
	for _, row := range rows {
		if t := tx.Create(&row); t.RowsAffected == 0 {
			zap.S().Warn("CreateBidirectional: insert failed, rolling back")
			tx.Rollback()
			return errors.New("create relation failed")
		}
	}
	return tx.Commit().Error
}


