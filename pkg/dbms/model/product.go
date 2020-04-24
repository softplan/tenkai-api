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
	ProductID         int       `json:"productId"`
	Date              time.Time `json:"date"`
	Version           string    `json:"version"`
	CopyLatestRelease bool      `gorm:"-" json:"copyLatestRelease"`
	Locked            bool      `json:"locked"`
}

//ProductVersionService struct
type ProductVersionService struct {
	gorm.Model
	ProductVersionID   int    `json:"productVersionId"`
	ServiceName        string `json:"serviceName"`
	DockerImageTag     string `json:"dockerImageTag"`
	LatestVersion      string `gorm:"-" json:"latestVersion"`
	ChartLatestVersion string `gorm:"-" json:"chartLatestVersion"`
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

//ListDockerTagsRequest structure
type ListDockerTagsRequest struct {
	ImageName string `json:"imageName"`
	From      string `json:"from"`
}

//ListDockerTagsResult structure
type ListDockerTagsResult struct {
	TagResponse []TagResponse `json:"tags"`
}

//TagResponse Structure
type TagResponse struct {
	Created time.Time `json:"created"`
	Tag     string    `json:"tag"`
}
