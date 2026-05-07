package middlewares

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Menyimpan rate limiter untuk setiap alamat IP
var clients = sync.Map{}

// GetVisitor mengembalikan rate limiter untuk IP tertentu
func getVisitor(ip string, r rate.Limit, b int) *rate.Limiter {
	limiter, exists := clients.Load(ip)
	if !exists {
		// Jika IP belum ada, buat limiter baru
		newLimiter := rate.NewLimiter(r, b)
		clients.Store(ip, newLimiter)
		return newLimiter
	}
	return limiter.(*rate.Limiter)
}

// RateLimitMiddleware membatasi jumlah request dari IP yang sama.
// requestsPerSecond: jumlah request yang diizinkan per detik (misal 1 request per 5 detik = 0.2)
// burst: jumlah request yang boleh menumpuk dalam satu waktu
func RateLimitMiddleware(requestsPerSecond float64, burst int) gin.HandlerFunc {
	limit := rate.Limit(requestsPerSecond)
	
	// Fitur pembersihan IP usang berjalan di background (opsional tapi disarankan)
	// Untuk kesederhanaan portfolio, kita biarkan sync.Map terus menyimpan data selama server hidup
	
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := getVisitor(ip, limit, burst)
		
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Terlalu banyak permintaan. Silakan coba lagi nanti.",
			})
			return
		}
		
		c.Next()
	}
}
