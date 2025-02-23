package models

import (
	"gorm.io/gorm"
)

// ✅ Grup modeli
type Group struct {
	gorm.Model
	Name    string         `json:"name"`
	OwnerID uint           `json:"owner_id"`  // 🔥 Grup sahibi eklendi
	Members []GroupMember `json:"members"`
}

// ✅ Grup üyeleri için model
type GroupMember struct {
	gorm.Model
	GroupID uint `json:"group_id"`
	UserID  uint `json:"user_id"`
}
