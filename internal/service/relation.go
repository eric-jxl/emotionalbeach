package service

import (
	ebmetrics "emotionalBeach/internal/infra"
	"emotionalBeach/internal/models"
	"errors"
)

func (s *Service) FriendList(userID uint) ([]models.UserBasic, error) {
	return s.dao.FriendList(userID)
}

// AddFriend adds a bidirectional friendship by user IDs.
func (s *Service) AddFriend(userID, targetID uint) error {
	if userID == targetID {
		ebmetrics.FriendAddTotal.WithLabelValues("self").Inc()
		return errors.New("cannot add yourself as a friend")
	}
	if _, err := s.dao.FindUserByID(targetID); err != nil {
		ebmetrics.FriendAddTotal.WithLabelValues("not_found").Inc()
		return errors.New("target user not found")
	}
	if s.dao.FriendExists(userID, targetID) {
		ebmetrics.FriendAddTotal.WithLabelValues("already_exists").Inc()
		return errors.New("friendship already exists")
	}
	if err := s.dao.CreateFriendship(userID, targetID); err != nil {
		ebmetrics.FriendAddTotal.WithLabelValues("error").Inc()
		return err
	}
	ebmetrics.FriendAddTotal.WithLabelValues("success").Inc()
	return nil
}

// AddFriendByName resolves targetName to a user ID and delegates to AddFriend.
func (s *Service) AddFriendByName(userID uint, targetName string) error {
	target, err := s.dao.FindUserByName(targetName)
	if err != nil || target == nil {
		ebmetrics.FriendAddTotal.WithLabelValues("not_found").Inc()
		return errors.New("user not found")
	}
	return s.AddFriend(userID, target.ID)
}

