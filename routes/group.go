package routes

import (
	"net/http"

	"arcurachat_api/database"
	"arcurachat_api/models"

	"github.com/gin-gonic/gin"
)

// ✅ Grup oluşturma
func CreateGroup(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz işlem"})
		return
	}

	var input struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}

	group := models.Group{Name: input.Name, OwnerID: userID.(uint)}
	if err := database.DB.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Grup oluşturulamadı"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Grup başarıyla oluşturuldu", "data": group})
}

// ✅ Grup bilgilerini getir
func GetGroup(c *gin.Context) {
	groupID := c.Param("group_id")

	var group models.Group
	if err := database.DB.Preload("Members").First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Grup bulunamadı"})
		return
	}

	c.JSON(http.StatusOK, group)
}

// ✅ Grup bilgilerini güncelle
func UpdateGroup(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz işlem"})
		return
	}

	groupID := c.Param("group_id")

	var group models.Group
	if err := database.DB.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Grup bulunamadı"})
		return
	}

	if group.OwnerID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bu grubu güncellemeye yetkiniz yok"})
		return
	}

	var input struct {
		Name string `json:"name"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}

	database.DB.Model(&group).Updates(models.Group{Name: input.Name})
	c.JSON(http.StatusOK, gin.H{"message": "Grup bilgileri güncellendi", "data": group})
}

// ✅ Grubu sil
func DeleteGroup(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz işlem"})
		return
	}

	groupID := c.Param("group_id")

	var group models.Group
	if err := database.DB.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Grup bulunamadı"})
		return
	}

	if group.OwnerID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bu grubu silmeye yetkiniz yok"})
		return
	}

	if err := database.DB.Delete(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Grup silinemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Grup başarıyla silindi"})
}

// ✅ Gruba kullanıcı ekle
func AddMemberToGroup(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz işlem"})
		return
	}

	groupID := c.Param("group_id")

	var group models.Group
	if err := database.DB.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Grup bulunamadı"})
		return
	}

	// Kullanıcı grup sahibi mi?
	if group.OwnerID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bu gruba üye eklemeye yetkiniz yok"})
		return
	}

	// JSON Verisini Doğrula
	var input struct {
		UserID uint `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}

	// Kullanıcının var olup olmadığını kontrol et
	var user models.User
	if err := database.DB.First(&user, input.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	// Kullanıcı zaten grupta mı?
	var existingMember models.GroupMember
	if err := database.DB.Where("group_id = ? AND user_id = ?", group.ID, input.UserID).First(&existingMember).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Kullanıcı zaten grupta"})
		return
	}

	// Kullanıcıyı gruba ekle
	groupMember := models.GroupMember{
		GroupID: group.ID,
		UserID:  input.UserID,
	}

	if err := database.DB.Create(&groupMember).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kullanıcı gruba eklenemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Kullanıcı gruba eklendi"})
}


// ✅ Gruptan kullanıcı çıkar
func RemoveMemberFromGroup(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz işlem"})
		return
	}

	groupID := c.Param("group_id")
	removeUserID := c.Param("user_id")

	var group models.Group
	if err := database.DB.First(&group, groupID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Grup bulunamadı"})
		return
	}

	if group.OwnerID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bu gruptan kullanıcı çıkarmaya yetkiniz yok"})
		return
	}

	if err := database.DB.Where("group_id = ? AND user_id = ?", groupID, removeUserID).Delete(&models.GroupMember{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kullanıcı gruptan çıkarılamadı"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Kullanıcı gruptan çıkarıldı"})
}

// ✅ Grup route'larını kaydet
func RegisterGroupRoutes(router *gin.Engine) {
	groupRoutes := router.Group("/groups")
	groupRoutes.Use(AuthMiddleware()) // 🔥 JWT Doğrulaması Ekledik
	{
		groupRoutes.POST("/create", CreateGroup)
		groupRoutes.GET("/:group_id", GetGroup)
		groupRoutes.PUT("/:group_id", UpdateGroup)
		groupRoutes.DELETE("/:group_id", DeleteGroup)
		groupRoutes.POST("/:group_id/members", AddMemberToGroup)
		groupRoutes.DELETE("/:group_id/members/:user_id", RemoveMemberFromGroup)
	}
}
