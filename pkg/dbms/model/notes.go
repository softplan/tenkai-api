package model

import (
	"github.com/jinzhu/gorm"
)

//Notes - Notes Model
type Notes struct {
	gorm.Model
	ServiceName string `json:"serviceName"`
	Text        string `gorm:"type:text" json:"text"`
}
