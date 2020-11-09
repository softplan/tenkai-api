package model

import (
	"github.com/jinzhu/gorm"
)

//Deployment  struct
type Deployment struct {
	gorm.Model
	EnvironmentID uint   `json:"environment_id"`
	Chart         string `json:"chart"`
	UserID        uint   `json:"user_id"`
	Success       bool   `json:"success"`
	Message       string `json:"message"`
}

//DeploymentResponse struct response /deployments GET
type DeploymentResponse struct {
	Count      int64         `json:"count"`
	TotalPages int           `json:"total_pages"`
	Data       []Deployments `json:"data"`
}

//Deployments struct to fill with query result to response /deployments GET
type Deployments struct {
	Deployment
	Environment Environment `json:"environment"`
	Chart       string      `json:"chart"`
	User        User        `json:"user"`
	Success     bool        `json:"success"`
	Message     string      `json:"message"`
}
