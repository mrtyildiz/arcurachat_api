package routes

import (
	"net/http"
	"strconv"

	"arcurachat_api/database"
	"arcurachat_api/models"
	"arcurachat_api/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	
)

func RegisterRoutes(router *gin.Engine) {
	router.POST("/auth/register", RegisterUser)
	router.POST("/auth/login", LoginUser)
	router.GET("/profile", AuthMiddleware(), ProfileHandler)
	router.POST("/auth/logout", AuthMiddleware(), LogoutUser) // 🔥 Kullanıcı çıkışı
	router.POST("/auth/refresh", RefreshToken)               // 🔥 Token yenileme
	router.GET("/auth/me", AuthMiddleware(), GetCurrentUser) // 🔥 Oturum açmış kullanıcı bilgisi

	userRoutes := router.Group("/users")
	userRoutes.Use(AuthMiddleware())
	userRoutes.GET("/:id", GetUser)       // GET /users/:id
	userRoutes.PUT("/:id", UpdateUser)    // PUT /users/:id
	userRoutes.PUT("/:id/password", UpdatePassword) // 🔥 Şifre güncelleme
	userRoutes.DELETE("/:id", DeleteUser) // DELETE /users/:id
}

// Kullanıcı kayıt fonksiyonu
func RegisterUser(c *gin.Context) {
	var input models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Şifreyi hashle
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Şifre hashleme hatası"})
		return
	}
	input.Password = string(hashedPassword)

	// Kullanıcıyı kaydet
	if err := database.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Kullanıcı oluşturulamadı"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Kullanıcı başarıyla kaydedildi"})
}

// Kullanıcı giriş fonksiyonu
func LoginUser(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz giriş bilgileri"})
		return
	}

	var user models.User
	if err := database.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	// Şifreyi doğrula
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Geçersiz şifre"})
		return
	}

	// JWT oluştur
	token, expiresAt, err := utils.GenerateToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token oluşturulamadı"})
		return
	}

	// Kullanıcıya token ekleyelim
	database.DB.Model(&user).Updates(models.User{
		Token:          token,
		TokenExpiresAt: expiresAt,
	})

	c.JSON(http.StatusOK, gin.H{
		"message":   "Giriş başarılı",
		"token":     token,
		"expiresAt": expiresAt,
	})
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token gerekli"})
			c.Abort()
			return
		}

		// **🔥 `Bearer` kelimesini temizle**
		tokenString := ""
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:] // "Bearer " kelimesini çıkar
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token formatı hatalı"})
			c.Abort()
			return
		}

		// ✅ **Token doğrulama**
		userID, err := utils.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Geçersiz token"})
			c.Abort()
			return
		}

		// ✅ **User ID'yi context'e kaydet**
		c.Set("userID", userID)
		c.Next()
	}
}

func ProfileHandler(c *gin.Context) {
	userID := c.GetString("userID")

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"username":    user.Username,
		"first_name":  user.FirstName,
		"last_name":   user.LastName,
		"phone":       user.PhoneNumber,
		"email":       user.Email,
	})
}


func GetUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz işlem"})
		return
	}

	id := c.Param("id") // URL'den gelen ID

	// Kullanıcı sadece kendi bilgilerini görüntüleyebilir
	if id != strconv.Itoa(int(userID.(uint))) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bu işlemi yapmaya yetkiniz yok"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"email":      user.Email,
		"phone":      user.PhoneNumber,
	})
}


func UpdateUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz işlem"})
		return
	}

	id := c.Param("id")

	// Kullanıcı sadece kendi hesabını güncelleyebilir
	if id != strconv.Itoa(int(userID.(uint))) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bu işlemi yapmaya yetkiniz yok"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	var updateData struct {
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		Email       string `json:"email"`
		PhoneNumber string `json:"phone_number"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}

	database.DB.Model(&user).Updates(models.User{
		FirstName:   updateData.FirstName,
		LastName:    updateData.LastName,
		Email:       updateData.Email,
		PhoneNumber: updateData.PhoneNumber,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Kullanıcı bilgileri güncellendi"})
}


func DeleteUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz işlem"})
		return
	}

	id := c.Param("id")

	// Kullanıcı sadece kendi hesabını silebilir
	if id != strconv.Itoa(int(userID.(uint))) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bu işlemi yapmaya yetkiniz yok"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	database.DB.Delete(&user)
	c.JSON(http.StatusOK, gin.H{"message": "Kullanıcı başarıyla silindi"})
}


// 🔥 Kullanıcı Şifresini Güncelleme
func UpdatePassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz işlem"})
		return
	}

	id := c.Param("id")

	// Kullanıcı sadece kendi şifresini değiştirebilir
	if id != strconv.Itoa(int(userID.(uint))) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bu işlemi yapmaya yetkiniz yok"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	// Kullanıcının eski şifresini doğrulamak için giriş verisini al
	var passwordUpdate struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := c.ShouldBindJSON(&passwordUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}

	// 🔥 Eski şifreyi doğrula
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(passwordUpdate.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Eski şifre yanlış"})
		return
	}

	// 🔥 Yeni şifreyi hashle ve güncelle
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordUpdate.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Şifre hashleme başarısız"})
		return
	}

	// Yeni şifreyi veritabanına kaydet
	database.DB.Model(&user).Update("password", string(hashedPassword))

	c.JSON(http.StatusOK, gin.H{"message": "Şifre başarıyla güncellendi"})
}

// 🔥 Kullanıcı Çıkışı (POST /auth/logout)
func LogoutUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz işlem"})
		return
	}

	// Kullanıcının token'ını veritabanında sıfırla (logout)
	if err := database.DB.Model(&models.User{}).Where("id = ?", userID).Update("token", "").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Çıkış yapılamadı"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Başarıyla çıkış yapıldı"})
}

func RefreshToken(c *gin.Context) {
	var input struct {
		Token string `json:"token"`
	}

	// 🔥 Token başlıkta mı yoksa body içinde mi kontrol et
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		input.Token = authHeader[7:]
	} else if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token gerekli"})
		return
	}

	// Token doğrulama
	userID, err := utils.ValidateToken(input.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Geçersiz token"})
		return
	}

	// Yeni bir token oluştur
	newToken, expiresAt, err := utils.GenerateToken(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token yenileme başarısız"})
		return
	}

	// Yeni token'ı veritabanına kaydet
	if err := database.DB.Model(&models.User{}).Where("id = ?", userID).Update("token", newToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token veritabanına kaydedilemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Token yenilendi",
		"token":     newToken,
		"expiresAt": expiresAt,
	})
}

// 🔥 Oturum Açmış Kullanıcının Bilgilerini Getir (GET /auth/me)
func GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz işlem"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"email":      user.Email,
		"phone":      user.PhoneNumber,
	})
}
