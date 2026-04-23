package handlers

import (
	"belajar-sso/database"
	"belajar-sso/models"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Rahasia ini adalah "Tanda Tangan Digital" server Anda.
// Nanti kunci ini akan disalin ke aplikasi Tracker agar Tracker bisa memvalidasi token dari SSO.
var JWTSecretKey = []byte("KUNCI_RAHASIA_SUPER_KUAT_123")

// Handler: POST /register
func Register(w http.ResponseWriter, r *http.Request) {
	var input models.AuthInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		sendJSONError(w, "Format JSON salah", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		sendJSONError(w, "Gagal memproses enkripsi", http.StatusInternalServerError)
		return
	}

	var newID int
	err = database.DB.QueryRow(
		"INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id",
		input.Username, string(hash),
	).Scan(&newID)

	if err != nil {
		sendJSONError(w, "Username sudah terdaftar", http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Berhasil registrasi!"})
}

// Handler: POST /login
func Login(w http.ResponseWriter, r *http.Request) {
	var input models.AuthInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		sendJSONError(w, "Format JSON salah", http.StatusBadRequest)
		return
	}

	var user models.User
	err := database.DB.QueryRow("SELECT id, username, password_hash FROM users WHERE username = $1", input.Username).
		Scan(&user.ID, &user.Username, &user.PasswordHash)

	if err != nil {
		sendJSONError(w, "Username tidak ditemukan (Perhatikan huruf besar/kecil)", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		sendJSONError(w, "Password salah", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(JWTSecretKey)
	if err != nil {
		sendJSONError(w, "Gagal membuat token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
