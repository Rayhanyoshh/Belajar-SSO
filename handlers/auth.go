package handlers

import (
	"belajar-sso/database"
	"belajar-sso/models"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Rahasia ini adalah "Tanda Tangan Digital" server Anda.
// Nanti kunci ini akan disalin ke aplikasi Tracker agar Tracker bisa memvalidasi token dari SSO.
var JWTSecretKey = []byte("KUNCI_RAHASIA_SUPER_KUAT_123")

func generateRefreshToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Handler: POST /register
func Register(c *gin.Context) {
	var input models.AuthInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON salah"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses enkripsi"})
		return
	}

	var newID int
	err = database.DB.QueryRow(
		"INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id",
		input.Username, string(hash),
	).Scan(&newID)

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username sudah terdaftar"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Berhasil registrasi!"})
}

// Handler: POST /login
func Login(c *gin.Context) {
	var input models.AuthInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON salah"})
		return
	}

	var user models.User
	err := database.DB.QueryRow("SELECT id, username, password_hash FROM users WHERE username = $1", input.Username).
		Scan(&user.ID, &user.Username, &user.PasswordHash)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Username tidak ditemukan (Perhatikan huruf besar/kecil)"})
		return
	}

	if user.PasswordHash == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Akun ini terdaftar dengan Google, silakan login dengan Google"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(input.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Password salah"})
		return
	}

	// Access Token: Singkat! Hanya 15 menit.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Minute * 15).Unix(),
	})

	tokenString, err := token.SignedString(JWTSecretKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token"})
		return
	}

	// Refresh Token: Panjang! Sampai 7 hari.
	refreshToken := generateRefreshToken()
	expiresAt := time.Now().Add(time.Hour * 24 * 7)

	_, err = database.DB.Exec(
		"INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)",
		user.ID, refreshToken, expiresAt,
	)

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"refresh_token": refreshToken,
	})
}

// Handler: POST /refresh
func RefreshToken(c *gin.Context) {
	var input models.RefreshInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format JSON salah"})
		return
	}

	// Cek token di database
	var userID int
	var expiresAt time.Time
	err := database.DB.QueryRow("SELECT user_id, expires_at FROM refresh_tokens WHERE token = $1", input.RefreshToken).
		Scan(&userID, &expiresAt)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token tidak valid atau sudah ditarik"})
		return
	}

	// Cek kedaluwarsa
	if time.Now().After(expiresAt) {
		database.DB.Exec("DELETE FROM refresh_tokens WHERE token = $1", input.RefreshToken)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token kedaluwarsa, silakan login ulang"})
		return
	}

	// Ambil data user
	var username string
	err = database.DB.QueryRow("SELECT username FROM users WHERE id = $1", userID).Scan(&username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Data user tidak ditemukan"})
		return
	}

	// Buat Access Token baru (15 menit)
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(time.Minute * 15).Unix(),
	})
	newTokenString, _ := newToken.SignedString(JWTSecretKey)

	// Buat Refresh Token baru (Rotasi demi keamanan)
	newRefreshToken := generateRefreshToken()
	newExpiresAt := time.Now().Add(time.Hour * 24 * 7)

	database.DB.Exec(
		"UPDATE refresh_tokens SET token = $1, expires_at = $2 WHERE token = $3",
		newRefreshToken, newExpiresAt, input.RefreshToken,
	)

	c.JSON(http.StatusOK, gin.H{
		"token": newTokenString,
		"refresh_token": newRefreshToken,
	})
}
