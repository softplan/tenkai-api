package model

import (
	"time"

	"github.com/jinzhu/gorm"
)

//RequestDeployment deployment requested from user
type RequestDeployment struct {
	gorm.Model
	Success   bool `json:"success"`
	Processed bool `json:"processed"`
	UserID    uint `json:"user_id"`
}

//Deployment  struct
type Deployment struct {
	gorm.Model
	RequestDeploymentID uint   `json:"request_deployment_id"`
	EnvironmentID       uint   `json:"environment_id"`
	Chart               string `json:"chart"`
	ChartVersion        string `json:"chartVersion"`
	Processed           bool   `json:"processed"`
	Success             bool   `json:"success"`
	Message             string `json:"message"`
	DockerVersion       string `json:"dockerVersion"`
}

//DeploymentResponse struct response /deployments GET
type DeploymentResponse struct {
	Count      int64         `json:"count"`
	TotalPages int           `json:"total_pages"`
	Data       []Deployments `json:"data"`
}

//ResponseDeploymentResponse struct response /deployments GET
type ResponseDeploymentResponse struct {
	Count      int64                `json:"count"`
	TotalPages int                  `json:"total_pages"`
	Data       []RequestDeployments `json:"data"`
}

//EnvironmentName is a struct to be used with deployments payload response
type EnvironmentName struct {
	ID   uint   `json:"id"`
	Name string `json:"Name"`
}

//UserEmail is a struct to be used with deployments payload response
type UserEmail struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
}

//Deployments struct to fill with query result to response /requestDeployments GET
type Deployments struct {
	ID                  uint
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           *time.Time
	RequestDeploymentID uint            `json:"request_deployment_id"`
	Environment         string          `json:"environment"`
	Chart               string          `json:"chart"`
	Success             bool            `json:"success"`
	Message             string          `json:"message"`
	Processed           bool            `json:"processed"`
	DockerVersion       string          `json:"dockerVersion"`
	ChartVersion        string          `json:"chartVersion"`
}

//RequestDeployments struct to fill with query result to response /requestDeployments{id} GET
type RequestDeployments struct {
	ID        uint
	CreatedAt time.Time
	UpdatedAt time.Time
	User      string    `json:"user"`
	Success   bool      `json:"success"`
	Processed bool      `json:"processed"`
}
