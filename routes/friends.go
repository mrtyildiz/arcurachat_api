package routes

import (
	"net/http"

	"arcurachat_api/database"
	"arcurachat_api/models"

	"github.com/gin-gonic/gin"
)

// ✅ Arkadaşlık isteği gönder
func SendFriendRequest(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz işlem"})
		return
	}

	var input struct {
		ReceiverID uint `json:"receiver_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}

	// Kullanıcı zaten arkadaş mı?
	var existingFriendship models.Friendship
	if err := database.DB.Where("user_id = ? AND friend_id = ?", userID, input.ReceiverID).First(&existingFriendship).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Bu kullanıcıyla zaten arkadaşsınız"})
		return
	}

	// Önceden istek gönderilmiş mi?
	var existingRequest models.FriendRequest
	if err := database.DB.Where("sender_id = ? AND receiver_id = ? AND status = 'pending'", userID, input.ReceiverID).First(&existingRequest).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Zaten bekleyen bir arkadaşlık isteğiniz var"})
		return
	}

	friendRequest := models.FriendRequest{
		SenderID:   userID.(uint),
		ReceiverID: input.ReceiverID,
		Status:     "pending",
	}

	if err := database.DB.Create(&friendRequest).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Arkadaşlık isteği gönderilemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Arkadaşlık isteği gönderildi"})
}

// ✅ Gelen arkadaşlık isteklerini getir
func GetFriendRequests(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz işlem"})
		return
	}

	var requests []models.FriendRequest
	if err := database.DB.Where("receiver_id = ? AND status = 'pending'", userID).Find(&requests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Arkadaşlık istekleri alınamadı"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"friend_requests": requests})
}

// ✅ Arkadaşlık isteğini kabul et
func AcceptFriendRequest(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz işlem"})
		return
	}

	requestID := c.Param("request_id")
	var friendRequest models.FriendRequest
	if err := database.DB.First(&friendRequest, requestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Arkadaşlık isteği bulunamadı"})
		return
	}

	if friendRequest.ReceiverID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bu isteği kabul etmeye yetkiniz yok"})
		return
	}

	friendRequest.Status = "accepted"
	database.DB.Save(&friendRequest)

	// Arkadaşlık kaydını oluştur
	friendship := models.Friendship{
		UserID:   friendRequest.SenderID,
		FriendID: friendRequest.ReceiverID,
	}
	database.DB.Create(&friendship)

	// Çift yönlü arkadaşlık için ikinci kayıt
	friendshipReverse := models.Friendship{
		UserID:   friendRequest.ReceiverID,
		FriendID: friendRequest.SenderID,
	}
	database.DB.Create(&friendshipReverse)

	c.JSON(http.StatusOK, gin.H{"message": "Arkadaşlık isteği kabul edildi"})
}

// ✅ Arkadaşlık isteğini reddet
func RejectFriendRequest(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz işlem"})
		return
	}

	requestID := c.Param("request_id")
	var friendRequest models.FriendRequest
	if err := database.DB.First(&friendRequest, requestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Arkadaşlık isteği bulunamadı"})
		return
	}

	if friendRequest.ReceiverID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bu isteği reddetmeye yetkiniz yok"})
		return
	}

	friendRequest.Status = "rejected"
	database.DB.Save(&friendRequest)

	c.JSON(http.StatusOK, gin.H{"message": "Arkadaşlık isteği reddedildi"})
}

// ✅ Arkadaşlığı sonlandır
func RemoveFriend(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz işlem"})
		return
	}

	friendID := c.Param("friend_id")

	if err := database.DB.Where("user_id = ? AND friend_id = ?", userID, friendID).Delete(&models.Friendship{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Arkadaş silinemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Arkadaşlık sona erdirildi"})
}

// ✅ Arkadaşlık route'larını kaydet
func RegisterFriendRoutes(router *gin.Engine) {
	friendRoutes := router.Group("/friends")
	friendRoutes.Use(AuthMiddleware()) // 🔥 JWT Doğrulaması Ekledik
	{
		friendRoutes.POST("/request", SendFriendRequest)
		friendRoutes.GET("/requests", GetFriendRequests)
		friendRoutes.POST("/accept/:request_id", AcceptFriendRequest)
		friendRoutes.DELETE("/reject/:request_id", RejectFriendRequest)
		friendRoutes.DELETE("/:friend_id", RemoveFriend)
	}
}
