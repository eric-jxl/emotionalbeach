// Package relation defines the relation domain contracts (interfaces).
package relation

import "emotionalBeach/internal/models"

// Repository defines the data-access contract for friend-relation persistence.
type Repository interface {
	// FriendList returns all friends of the given user.
	FriendList(userID uint) ([]models.UserBasic, error)
	// Exists reports whether a bidirectional friendship already exists.
	Exists(ownerID, targetID uint) bool
	// CreateBidirectional inserts two relation records in a single transaction.
	CreateBidirectional(ownerID, targetID uint) error
}

// Service defines the business-logic contract for friend-relation operations.
type Service interface {
	FriendList(userID uint) ([]models.UserBasic, error)
	AddFriend(userID, targetID uint) error
	AddFriendByName(userID uint, targetName string) error
}

