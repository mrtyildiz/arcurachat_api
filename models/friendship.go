package models

import (
	"gorm.io/gorm"
)

// ✅ Arkadaşlık İstek Modeli
type FriendRequest struct {
	gorm.Model
	SenderID   uint `json:"sender_id"`
	ReceiverID uint `json:"receiver_id"`
	Status     string `json:"status"` // pending, accepted, rejected
}

// ✅ Arkadaşlık Modeli
type Friendship struct {
	gorm.Model
	UserID   uint `json:"user_id"`
	FriendID uint `json:"friend_id"`
}
