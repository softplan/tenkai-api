package model

import "github.com/jinzhu/gorm"

type ConfigMap struct {
	gorm.Model
	Name  string `json:"name"`
	Value string `json:"value"`
}
