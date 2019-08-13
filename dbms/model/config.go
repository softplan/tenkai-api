package model

import "github.com/jinzhu/gorm"

//ConfigMap  struct
type ConfigMap struct {
	gorm.Model
	Name  string `json:"name"`
	Value string `json:"value"`
}
