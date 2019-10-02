package model

import (
	"github.com/jinzhu/gorm"
	"time"
)

//Product struct
type Product struct {
	gorm.Model
	Name string `json:"name"`
}

//ProductVersion struct
type ProductVersion struct {
	gorm.Model
	ProductID int       `json:"productId"`
	Date      time.Time `json:"date"`
	Version   string    `json:"version"`
}

//ProductVersionService struct
type ProductVersionService struct {
	gorm.Model
	ProductVersionID int    `json:"productVersionId"`
	ServiceName      string `json:"serviceName"`
	DockerImageTag   string `json:"dockerImageTag"`
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
