package main

import (
	"arcurachat_api/database"
	"arcurachat_api/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	// Veritabanına bağlan
	database.ConnectDatabase()

	// Gin Router başlat
	r := gin.Default()

	// Rotaları kaydet
	routes.RegisterRoutes(r)
	routes.RegisterMessageRoutes(r) // Mesaj işlemleri

	// Sunucuyu başlat
	r.Run(":8080")
}
