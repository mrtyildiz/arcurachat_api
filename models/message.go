package models

import (
	"time"

	"gorm.io/gorm"
)

// 🔥 Mesaj Modeli
type Message struct {
	gorm.Model
	ConversationID uint      `json:"conversation_id"` // Hangi konuşmaya ait
	SenderID       uint      `json:"sender_id"`       // Mesajı gönderen
	Content        string    `json:"content"`         // Mesaj içeriği
	IsRead         bool      `json:"is_read"`         // Okundu bilgisi
	ReadAt         time.Time `json:"read_at"`         // Okunduğu zaman
}
