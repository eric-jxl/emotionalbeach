// Package relationsvc implements the relation.Service business-logic contract.
package relationsvc

import (
	reldomain "emotionalBeach/internal/domain/relation"
	userdomain "emotionalBeach/internal/domain/user"
	ebmetrics "emotionalBeach/internal/infra/metrics"
	"emotionalBeach/internal/models"
	"errors"

	"github.com/google/wire"
)

// Set is the Wire provider set for the relation service.
var Set = wire.NewSet(
	NewSvc,
	wire.Bind(new(reldomain.Service), new(*Svc)),
)

// Svc implements reldomain.Service.
type Svc struct {
	relRepo  reldomain.Repository
	userRepo userdomain.Repository
}

// NewSvc constructs a Svc with its repository dependencies.
func NewSvc(relRepo reldomain.Repository, userRepo userdomain.Repository) *Svc {
	return &Svc{relRepo: relRepo, userRepo: userRepo}
}

func (s *Svc) FriendList(userID uint) ([]models.UserBasic, error) {
	return s.relRepo.FriendList(userID)
}

// AddFriend adds a bidirectional friendship by user IDs.
func (s *Svc) AddFriend(userID, targetID uint) error {
	if userID == targetID {
		ebmetrics.FriendAddTotal.WithLabelValues("self").Inc()
		return errors.New("cannot add yourself as a friend")
	}
	if _, err := s.userRepo.FindByID(targetID); err != nil {
		ebmetrics.FriendAddTotal.WithLabelValues("not_found").Inc()
		return errors.New("target user not found")
	}
	if s.relRepo.Exists(userID, targetID) {
		ebmetrics.FriendAddTotal.WithLabelValues("already_exists").Inc()
		return errors.New("friendship already exists")
	}
	if err := s.relRepo.CreateBidirectional(userID, targetID); err != nil {
		ebmetrics.FriendAddTotal.WithLabelValues("error").Inc()
		return err
	}
	ebmetrics.FriendAddTotal.WithLabelValues("success").Inc()
	return nil
}

// AddFriendByName resolves targetName to a user ID and delegates to AddFriend.
func (s *Svc) AddFriendByName(userID uint, targetName string) error {
	target, err := s.userRepo.FindByName(targetName)
	if err != nil || target == nil {
		ebmetrics.FriendAddTotal.WithLabelValues("not_found").Inc()
		return errors.New("user not found")
	}
	return s.AddFriend(userID, target.ID)
}

