package routes

import (
	"net/http"
	"strings"

	"arcurachat_api/database"
	"arcurachat_api/models"

	"github.com/gin-gonic/gin"
)

// ✅ SQL Injection'a karşı güvenli bir arama fonksiyonu
func sanitizeQuery(query string) string {
	query = strings.ReplaceAll(query, "%", `\%`)
	query = strings.ReplaceAll(query, "_", `\_`)
	query = strings.TrimSpace(query)
	return query
}

// ✅ Kullanıcıları Arama (SQL Injection'a karşı güvenli)
func SearchUsers(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arama terimi belirtilmelidir"})
		return
	}

	query = sanitizeQuery(query) // 🔥 Kullanıcı girdisini temizle

	var users []models.User
	if err := database.DB.Where("username LIKE ? ESCAPE '\\' OR email LIKE ? ESCAPE '\\'", "%"+query+"%", "%"+query+"%").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kullanıcılar aranırken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// ✅ Grupları Arama (SQL Injection'a karşı güvenli)
func SearchGroups(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arama terimi belirtilmelidir"})
		return
	}

	query = sanitizeQuery(query) // 🔥 Kullanıcı girdisini temizle

	var groups []models.Group
	if err := database.DB.Where("name LIKE ? ESCAPE '\\'", "%"+query+"%").Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gruplar aranırken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"groups": groups})
}

// ✅ Mesajları Arama (SQL Injection'a karşı güvenli)
func SearchMessages(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Arama terimi belirtilmelidir"})
		return
	}

	query = sanitizeQuery(query) // 🔥 Kullanıcı girdisini temizle

	var messages []models.Message
	if err := database.DB.Where("content LIKE ? ESCAPE '\\'", "%"+query+"%").Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Mesajlar aranırken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// ✅ Arama route'larını kaydet
func RegisterSearchRoutes(router *gin.Engine) {
	searchRoutes := router.Group("/search")
	searchRoutes.GET("/users", SearchUsers)
	searchRoutes.GET("/groups", SearchGroups)
	searchRoutes.GET("/messages", SearchMessages)
}
