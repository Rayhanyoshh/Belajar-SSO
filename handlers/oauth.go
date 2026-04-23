package handlers

import (
	"belajar-sso/database"
	"belajar-sso/models"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	GoogleOauthConfig = &oauth2.Config{
		RedirectURL:  getEnvOrDefault("GOOGLE_REDIRECT_URL", "http://localhost:8081/auth/google/callback"),
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
	oauthStateString = "pseudo-random-state"
)

// Helper function untuk mengambil nilai dari environment variable, jika kosong gunakan default
func getEnvOrDefault(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// GoogleLogin mengarahkan user ke halaman login Google
func GoogleLogin(c *gin.Context) {
	url := GoogleOauthConfig.AuthCodeURL(oauthStateString)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback menangani respons dari Google setelah user login
func GoogleCallback(c *gin.Context) {
	state := c.Query("state")
	if state != oauthStateString {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
		return
	}

	code := c.Query("code")
	token, err := GoogleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menukar token dengan Google"})
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mendapatkan data user dari Google"})
		return
	}
	defer response.Body.Close()

	var googleUser struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(response.Body).Decode(&googleUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membaca data dari Google"})
		return
	}

	var user models.User
	err = database.DB.QueryRow("SELECT id, username FROM users WHERE email = $1 OR google_id = $2", googleUser.Email, googleUser.ID).
		Scan(&user.ID, &user.Username)

	if err != nil {
		if err == sql.ErrNoRows {
			// Buat user baru jika belum ada
			username := googleUser.Name
			
			// Jika nama sudah ada, tambahkan string unik (misal: ID Google di belakangnya)
			// Namun untuk kesederhanaan, asumsikan nama aman, atau biarkan database error jika duplikat
			// Di produksi, kita perlu mekanisme pembuatan username unik.
			
			err = database.DB.QueryRow(
				"INSERT INTO users (username, email, google_id) VALUES ($1, $2, $3) RETURNING id",
				username, googleUser.Email, googleUser.ID,
			).Scan(&user.ID)
			
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mendaftarkan user Google, nama mungkin sudah dipakai."})
				return
			}
			user.Username = username
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Terjadi kesalahan pada database"})
			return
		}
	}

	// Buat Access Token & Refresh Token lokal kita
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Minute * 15).Unix(),
	})
	tokenString, _ := jwtToken.SignedString(JWTSecretKey)

	refreshToken := generateRefreshToken()
	expiresAt := time.Now().Add(time.Hour * 24 * 7)

	database.DB.Exec(
		"INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)",
		user.ID, refreshToken, expiresAt,
	)

	// Redirect kembali ke halaman Frontend dengan membawa token
	// Gunakan FRONTEND_URL dari environment agar dinamis antara Local dan Production
	frontendURL := getEnvOrDefault("FRONTEND_URL", "http://localhost")
	redirectURL := frontendURL + "/?token=" + tokenString + "&refresh_token=" + refreshToken
	c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}
