package middlewares

import (
	"net/http"
	"strings"
	"github.com/golang-jwt/jwt/v5"
)

var JwtKey = []byte("KUNCI_RAHASIA_SUPER_KUAT_123") // Harus sama dengan yang ada di handler login

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Akses ditolak: Token tidak ditemukan", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			return JwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Akses ditolak: Token tidak valid", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}
