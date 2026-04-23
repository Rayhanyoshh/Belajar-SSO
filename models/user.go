package models

// User merepresentasikan struktur tabel di database
type User struct {
	ID           int    `json:"id"`
	Username     string `json:"username"`
	// Tanda json:"-" memastikan password hash TIDAK PERNAH dikembalikan ke user via API (Keamanan)
	PasswordHash string `json:"-"` 
}

// AuthInput merepresentasikan data mentah yang dikirim user saat login/register
type AuthInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
