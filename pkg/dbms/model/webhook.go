package model

import "github.com/jinzhu/gorm"

// WebHook Structure
type WebHook struct {
	gorm.Model
	Name          string `json:"name"`
	Type          string `json:"type"`
	URL           string `json:"url"`
	EnvironmentID int    `json:"environmentId"`
}

//WebHookReponse struct
type WebHookReponse struct {
	List []WebHook `json:"list"`
}
