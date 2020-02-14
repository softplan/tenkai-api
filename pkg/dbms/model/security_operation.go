package model

import "github.com/jinzhu/gorm"
import "github.com/lib/pq"

//SecurityOperation - SecurityOperation
type SecurityOperation struct {
	gorm.Model
	Name     string         `json:"name"`
	Policies pq.StringArray `gorm:"type:text[]" json:"policies" `
}

//SecurityOperationResponse - SecurityOperationResponse
type SecurityOperationResponse struct {
	List []SecurityOperation `json:"list"`
}

//GetUserPolicyByEnvironmentRequest - GetUserPolicyByEnvironmentRequest
type GetUserPolicyByEnvironmentRequest struct {
	Email         string `json:"email"`
	EnvironmentID int    `json:"environmentId"`
}
