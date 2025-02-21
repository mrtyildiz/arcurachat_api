package utils

import (
	"errors"
	"log"
	"os"
	"time"
	"strconv"

	"github.com/golang-jwt/jwt/v4"
)

// 🔥 **Global `jwtSecret` değişkeni** - Tüm sistemde aynı olacak
var jwtSecret = []byte(getEnv("JWT_SECRET", "supersecretkey"))

// ✅ Çevresel değişken okuma fonksiyonu
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// ✅ **JWT Token oluşturma fonksiyonu**
func GenerateToken(userID uint) (string, time.Time, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // 24 saat geçerli

	claims := jwt.MapClaims{
		"sub": strconv.Itoa(int(userID)), // **Kullanıcı ID uint -> string**
		"exp": expirationTime.Unix(),     // **Token süresi**
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		log.Println("JWT oluşturulamadı:", err)
		return "", time.Time{}, err
	}

	return tokenString, expirationTime, nil
}

// ✅ **JWT Token doğrulama fonksiyonu**
func ValidateToken(tokenString string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		log.Println("Hata: Token çözülemedi -", err)
		return 0, errors.New("geçersiz token")
	}

	// ✅ **Claims bilgilerini al**
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// 🔥 `sub` değerini uint olarak döndür
		userID, err := strconv.Atoi(claims["sub"].(string))
		if err != nil {
			return 0, errors.New("token'daki user ID hatalı")
		}
		return uint(userID), nil
	}

	return 0, errors.New("token geçersiz")
}
