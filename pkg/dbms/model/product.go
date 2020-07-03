package model

import (
	"time"

	"github.com/jinzhu/gorm"
)

//Product struct
type Product struct {
	gorm.Model
	Name string `json:"name"`
}

//ProductVersion struct
type ProductVersion struct {
	gorm.Model
	ProductID   int       `json:"productId"`
	Date        time.Time `json:"date"`
	Version     string    `json:"version"`
	BaseRelease int       `gorm:"-" json:"baseRelease"`
	Locked      bool      `json:"locked"`
}

//ProductVersionService struct
type ProductVersionService struct {
	gorm.Model
	ProductVersionID   int    `json:"productVersionId"`
	ServiceName        string `json:"serviceName"`
	DockerImageTag     string `json:"dockerImageTag"`
	LatestVersion      string `gorm:"-" json:"latestVersion"`
	ChartLatestVersion string `gorm:"-" json:"chartLatestVersion"`
	Notes              string `json:"notes"`
}

//ProductRequestReponse struct
type ProductRequestReponse struct {
	List []Product `json:"list"`
}

//ProductVersionRequestReponse struct
type ProductVersionRequestReponse struct {
	List []ProductVersion `json:"list"`
}

//ProductVersionServiceRequestReponse struct
type ProductVersionServiceRequestReponse struct {
	List []ProductVersionService `json:"list"`
}
