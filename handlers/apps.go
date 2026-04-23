package handlers

import (
	"belajar-sso/database"
	"belajar-sso/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetApplications(c *gin.Context) {
	rows, err := database.DB.Query("SELECT id, app_name, description, icon_name, action_key FROM applications")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data aplikasi"})
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

	c.JSON(http.StatusOK, apps)
}
