package routes

import (
	"net/http"
	"time"

	"arcurachat_api/database"
	"arcurachat_api/models"
	"github.com/gin-gonic/gin"
)

func RegisterMessageRoutes(router *gin.Engine) {
	messageRoutes := router.Group("/messages")
	messageRoutes.Use(AuthMiddleware())

	messageRoutes.POST("/send", SendMessage)                   // 🔥 Mesaj gönderme
	messageRoutes.GET("/:conversation_id", GetMessagesByConversation) // 🔥 Belirli konuşmanın mesajlarını getir
	messageRoutes.DELETE("/:message_id", DeleteMessage)       // 🔥 Mesajı sil
	messageRoutes.PUT("/:message_id/edit", EditMessage)       // 🔥 Mesajı düzenle
	messageRoutes.POST("/:message_id/read", MarkMessageAsRead) // 🔥 Mesajı okundu olarak işaretle
}
// 🔥 1. Mesaj Gönderme (POST /messages/send)
func SendMessage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz işlem"})
		return
	}

	var input struct {
		ConversationID uint   `json:"conversation_id"`
		Content        string `json:"content"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}

	message := models.Message{
		ConversationID: input.ConversationID,
		SenderID:       userID.(uint),
		Content:        input.Content,
		IsRead:         false,
	}

	if err := database.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Mesaj gönderilemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mesaj başarıyla gönderildi", "data": message})
}

// 🔥 2. Belirli Bir Konuşmanın Mesajlarını Getir (GET /messages/:conversation_id)
func GetMessagesByConversation(c *gin.Context) {
	conversationID := c.Param("conversation_id")

	var messages []models.Message
	if err := database.DB.Where("conversation_id = ?", conversationID).Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Mesajlar alınamadı"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": messages})
}

// 🔥 3. Mesajı Silme (DELETE /messages/:message_id)
func DeleteMessage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz işlem"})
		return
	}

	messageID := c.Param("message_id")
	var message models.Message

	if err := database.DB.First(&message, messageID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mesaj bulunamadı"})
		return
	}

	// Sadece mesajın sahibi silebilir
	if message.SenderID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bu mesajı silmeye yetkiniz yok"})
		return
	}

	database.DB.Delete(&message)
	c.JSON(http.StatusOK, gin.H{"message": "Mesaj başarıyla silindi"})
}

// 🔥 4. Mesajı Düzenleme (PUT /messages/:message_id/edit)
func EditMessage(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz işlem"})
		return
	}

	messageID := c.Param("message_id")
	var message models.Message

	// Mesajı al
	if err := database.DB.First(&message, messageID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mesaj bulunamadı"})
		return
	}

	// 🔥 Sadece mesajın sahibi düzenleyebilir
	if message.SenderID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bu mesajı düzenlemeye yetkiniz yok"})
		return
	}

	var input struct {
		Content string `json:"content"`
	}

	// JSON formatındaki yeni içerik verisini al
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}

	// Mesaj içeriğini güncelle
	database.DB.Model(&message).Update("content", input.Content)

	c.JSON(http.StatusOK, gin.H{
		"message": "Mesaj başarıyla güncellendi",
		"data": gin.H{
			"id":             message.ID,
			"conversation_id": message.ConversationID,
			"sender_id":      message.SenderID,
			"content":        message.Content,
			"is_read":        message.IsRead,
			"read_at":        message.ReadAt,
		},
	})
}


// 🔥 5. Mesajı Okundu Olarak İşaretleme (POST /messages/:message_id/read)
func MarkMessageAsRead(c *gin.Context) {
	messageID := c.Param("message_id")
	var message models.Message

	if err := database.DB.First(&message, messageID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Mesaj bulunamadı"})
		return
	}

	// Zaten okunmuşsa işlem yapma
	if message.IsRead {
		c.JSON(http.StatusOK, gin.H{"message": "Mesaj zaten okunmuş"})
		return
	}

	database.DB.Model(&message).Updates(models.Message{
		IsRead: true,
		ReadAt: time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{"message": "Mesaj okundu olarak işaretlendi"})
}
