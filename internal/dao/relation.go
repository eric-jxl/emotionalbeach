package dao

import (
	"emotionalBeach/internal/models"
	"errors"

	"go.uber.org/zap"
)

func (d *dao) FriendList(userID uint) ([]models.UserBasic, error) {
	var relations []models.Relation
	if tx := d.db.Where("owner_id = ? AND type = 1", userID).Find(&relations); tx.RowsAffected == 0 {
		return nil, errors.New("no friends found")
	}
	ids := make([]uint, 0, len(relations))
	for _, rel := range relations {
		ids = append(ids, rel.TargetID)
	}
	var users []models.UserBasic
	if tx := d.db.Where("id IN ?", ids).Find(&users); tx.RowsAffected == 0 {
		return nil, errors.New("friend users not found")
	}
	return users, nil
}

func (d *dao) FriendExists(ownerID, targetID uint) bool {
	var rel models.Relation
	if d.db.Where("owner_id = ? AND target_id = ? AND type = 1", ownerID, targetID).First(&rel).RowsAffected == 1 {
		return true
	}
	return d.db.Where("owner_id = ? AND target_id = ? AND type = 1", targetID, ownerID).First(&rel).RowsAffected == 1
}

func (d *dao) CreateFriendship(ownerID, targetID uint) error {
	tx := d.db.Begin()
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
			zap.S().Warn("CreateFriendship: insert failed, rolling back")
			tx.Rollback()
			return errors.New("create relation failed")
		}
	}
	return tx.Commit().Error
}

