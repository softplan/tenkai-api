package model

import "github.com/jinzhu/gorm"

// WebHook Structure
type WebHook struct {
	gorm.Model
	Name           string `json:"name"`
	Type           string `json:"type"`
	URL            string `json:"url"`
	EnvironmentID  int    `json:"environmentId"`
	AdditionalData string `json:"additionalData"`
}

//WebHookReponse struct
type WebHookReponse struct {
	List []WebHook `json:"list"`
}

//WebHookPostPayload struct
type WebHookPostPayload struct {
	Environment string `json:"environment"`
	ProductName string `json:"productName"`
	Release     string `json:"release"`
}

//WebHookNewReleasePostPayload struct
type WebHookNewReleasePostPayload struct {
	Environment    string `json:"environment"`
	ProductName    string `json:"productName"`
	Release        string `json:"release"`
	AdditionalData string `json:"additionalData"`
}
