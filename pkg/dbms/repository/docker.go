package repository

import (
	"github.com/jinzhu/gorm"
	model2 "github.com/softplan/tenkai-api/pkg/dbms/model"
)

//DockerDAOInterface DockerDAOInterface
type DockerDAOInterface interface {
	CreateDockerRepo(item model2.DockerRepo) (int, error)
	GetDockerRepositoryByHost(host string) (model2.DockerRepo, error)
	DeleteDockerRepo(id int) error
	ListDockerRepos() ([]model2.DockerRepo, error)
}

//DockerDAOImpl DockerDAOImpl
type DockerDAOImpl struct {
	Db *gorm.DB
}

//CreateDockerRepo - Create a new docker repo
func (dao DockerDAOImpl) CreateDockerRepo(item model2.DockerRepo) (int, error) {
	if err := dao.Db.Create(&item).Error; err != nil {
		return -1, err
	}
	return int(item.ID), nil
}

//GetDockerRepositoryByHost - Get a repo by host
func (dao DockerDAOImpl) GetDockerRepositoryByHost(host string) (model2.DockerRepo, error) {
	item := model2.DockerRepo{}
	if err := dao.Db.Where(&model2.DockerRepo{Host: host}).Find(&item).Error; err != nil {
		return item, err
	}
	return item, nil
}

//DeleteDockerRepo - Deletes a docker repo
func (dao DockerDAOImpl) DeleteDockerRepo(id int) error {
	return dao.Db.Unscoped().Delete(model2.DockerRepo{}, id).Error
}

//ListDockerRepos - List docker repos
func (dao DockerDAOImpl) ListDockerRepos() ([]model2.DockerRepo, error) {
	list := make([]model2.DockerRepo, 0)
	if err := dao.Db.Find(&list).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model2.DockerRepo, 0), nil
		}
		return nil, err
	}
	return list, nil
}
