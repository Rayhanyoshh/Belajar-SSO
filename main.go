package main

import (
	"belajar-sso/database"
	"belajar-sso/handlers"
	"belajar-sso/middlewares"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		err := database.DB.Ping()
		dbStatus := "connected"
		if err != nil {
			dbStatus = "disconnected"
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "gotracker-sso",
			"db":        dbStatus,
			"timestamp": time.Now().UTC(),
		})
	})

	// 4. Mendaftarkan Route
	// Menerapkan Rate Limiter: Maksimal 1 request per detik dengan burst 5 untuk endpoint auth (mencegah brute force)
	authLimiter := middlewares.RateLimitMiddleware(1, 5)

	r.POST("/register", authLimiter, handlers.Register)
	r.POST("/login", authLimiter, handlers.Login)
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

	slog.Info("Server Microservice SSO mulai berjalan", "port", "8081")

	srv := &http.Server{
		Addr:    ":8081",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Error saat menjalankan server", "error", err)
		}
	}()

	// Menunggu sinyal interrupt untuk graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Sinyal interrupt diterima, mematikan server SSO...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Server terpaksa mati", "error", err)
	}
	
	slog.Info("Server SSO telah berhenti dengan aman.")
}
