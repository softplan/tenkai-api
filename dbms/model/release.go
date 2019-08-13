package model

import (
	"github.com/jinzhu/gorm"
)

//Release struct
type Release struct {
	gorm.Model
	ChartName string `json:"chartName"`
	Release   string `json:"release"`
}

//ReleaseResult struct
type ReleaseResult struct {
	Releases []Release `json:"releases"`
}

//Dependency struct
type Dependency struct {
	gorm.Model
	ReleaseID int    `json:"release_id"`
	ChartName string `json:"chartName"`
	Version   string `json:"version"`
}

//DependencyResult struct
type DependencyResult struct {
	Dependencies []Dependency `json:"dependencies"`
}
