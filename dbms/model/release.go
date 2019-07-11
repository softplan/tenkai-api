package model

import (
	"github.com/jinzhu/gorm"
)

type Release struct {
	gorm.Model
	ChartName string `json:"chartName"`
	Release   string `json:"release"`
}

type ReleaseResult struct {
	Releases []Release `json:"releases"`
}

type Dependency struct {
	gorm.Model
	ReleaseID int    `json:"release_id"`
	ChartName string `json:"chartName"`
	Version   string `json:"version"`
}

type DependencyResult struct {
	Dependencies []Dependency `json:"dependencies"`
}
