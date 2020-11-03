package model

import "github.com/jinzhu/gorm"

//Deployment  struct
type Deployment struct {
	gorm.Model
	Environment uint `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Chart  string `json:"chart"`
	User uint `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Status bool `json:"status"`
	Message string `json:"message"`
}
