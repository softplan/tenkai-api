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

//UserEmail is a struct to be used with deployments payload response
type UserEmail struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
}

//EnvironmentName is a struct to be used with deployments payload response
type EnvironmentName struct {
	ID   uint   `json:"id"`
	Name string `json:"Name"`
}

//Deployments struct to fill with query result to response /deployments GET
type Deployments struct {
	Deployment
	Environment     EnvironmentName `json:"environment"`
	ChartResponse   string          `json:"chart"`
	User            UserEmail       `json:"user"`
	SuccessResponse bool            `json:"success"`
	MessageResponse string          `json:"message"`
}
