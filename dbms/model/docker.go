package model

import (
	"github.com/jinzhu/gorm"
	"time"
)

//ListDockerTagsRequest structure
type ListDockerTagsRequest struct {
	ImageName string `json:"imageName"`
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

//DockerRepo structure
type DockerRepo struct {
	gorm.Model
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
}

//TagsResult structure
type TagsResult struct {
	Name string
	Tags []string
}

//ManifestResult structure
type ManifestResult struct {
	History []ManifestHistory `json:"history"`
}

//ManifestHistory structure
type ManifestHistory struct {
	V1Compatibility string `json:"v1Compatibility"`
}

//V1Compatibility  structure
type V1Compatibility struct {
	Created time.Time `json:"created"`
}

//ListDockerRepositoryResponse structure
type ListDockerRepositoryResponse struct {
	Repositories []DockerRepo `json:"repositories"`
}
