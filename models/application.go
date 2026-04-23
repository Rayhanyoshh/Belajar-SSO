package models

type Application struct {
	ID          int    `json:"id"`
	AppName     string `json:"app_name"`
	Description string `json:"description"`
	IconName    string `json:"icon_name"`
	ActionKey   string `json:"action_key"`
}
