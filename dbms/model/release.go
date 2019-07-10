package model

import (
	"github.com/jinzhu/gorm"
)

type Release struct {
	gorm.Model
	ChartName string    `json:"chartName"`
	Release   string    `json:"release"`
}

type ReleaseResult struct {
	Releases []Release `json:"releases"`
}
