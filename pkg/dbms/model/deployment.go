package model

import "github.com/jinzhu/gorm"

//Deployment  struct
type Deployment struct {
	gorm.Model
	EnvironmentID uint   `json:"environment_id"`
	Chart         string `json:"chart"`
	UserID        uint   `json:"user_id"`
	Success       bool   `json:"success"`
	Message       string `json:"message"`
}
