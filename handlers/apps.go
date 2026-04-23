package handlers

import (
	"belajar-sso/database"
	"belajar-sso/models"
	"encoding/json"
	"net/http"
)

func GetApplications(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query("SELECT id, app_name, description, icon_name, action_key FROM applications")
	if err != nil {
		http.Error(w, "Gagal mengambil data aplikasi", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var apps []models.Application
	for rows.Next() {
		var app models.Application
		if err := rows.Scan(&app.ID, &app.AppName, &app.Description, &app.IconName, &app.ActionKey); err == nil {
			apps = append(apps, app)
		}
	}

	if apps == nil {
		apps = []models.Application{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apps)
}
