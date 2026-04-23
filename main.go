package main

import (
	"belajar-sso/database"
	"belajar-sso/handlers"
	"belajar-sso/middlewares"
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	database.ConnectDB()

	r := gin.Default()

	// 3. Konfigurasi CORS (Gin)
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
		AllowHeaders:     []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 4. Mendaftarkan Route
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)
	r.POST("/refresh", handlers.RefreshToken)
	
	// Route OAuth2 Google
	r.GET("/auth/google/login", handlers.GoogleLogin)
	r.GET("/auth/google/callback", handlers.GoogleCallback)

	// Route yang dilindungi middleware
	protected := r.Group("/")
	protected.Use(middlewares.AuthMiddleware())
	{
		protected.GET("/applications", handlers.GetApplications)
	}

	fmt.Println("=======================================")
	fmt.Println("🔐 Server Microservice SSO berjalan di http://localhost:8081")
	fmt.Println("=======================================")

	r.Run(":8081")
}
