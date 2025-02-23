package database

import (
	"fmt"
	"log"
	"os"

	"arcurachat_api/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Veritabanına bağlanılamadı:", err)
	}

	fmt.Println("✅ Veritabanına bağlantı başarılı")
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Message{})
	db.AutoMigrate(&models.Group{})
	db.AutoMigrate(&models.GroupMember{})
	db.AutoMigrate(&models.Friendship{})
	db.AutoMigrate(&models.FriendRequest{})
	DB = db
}
