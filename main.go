package main

import (
	"belajar-sso/database"
	"belajar-sso/handlers"
	"belajar-sso/middlewares"
	"fmt"
	"net/http"
)

func main() {
	database.ConnectDB()

	mux := http.NewServeMux()
	
	// Dua endpoint untuk publik (tanpa pelindung)
	// Endpoint Publik (Tanpa Token)
	mux.HandleFunc("POST /register", handlers.Register)
	mux.HandleFunc("POST /login", handlers.Login)

	// Endpoint Terlindungi (Hanya User yang sudah Login)
	mux.HandleFunc("GET /applications", middlewares.AuthMiddleware(handlers.GetApplications))

	// Bungkus seluruh router (Mux) dengan perizinan CORS
	handler := middlewares.CORSMiddleware(mux)

	fmt.Println("=======================================")
	fmt.Println("🔐 Server Microservice SSO berjalan di http://localhost:8081")
	fmt.Println("=======================================")
	
	// Berjalan di port 8081 agar tidak bertabrakan dengan aplikasi GoTracker (8080)
	err := http.ListenAndServe(":8081", handler)
	if err != nil {
		fmt.Println("Server gagal berjalan:", err)
	}
}
