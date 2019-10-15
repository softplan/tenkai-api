package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/dbms/model"
)

//DockerDAOInterface DockerDAOInterface
type DockerDAOInterface interface {
	CreateDockerRepo(item model.DockerRepo) (int, error)
	GetDockerRepositoryByHost(host string) (model.DockerRepo, error)
	DeleteDockerRepo(id int) error
	ListDockerRepos() ([]model.DockerRepo, error)
}

//DockerDAOImpl DockerDAOImpl
type DockerDAOImpl struct {
	Db *gorm.DB
}

//CreateDockerRepo - Create a new docker repo
func (dao DockerDAOImpl) CreateDockerRepo(item model.DockerRepo) (int, error) {
	if err := dao.Db.Create(&item).Error; err != nil {
		return -1, err
	}
	return int(item.ID), nil
}

//GetDockerRepositoryByHost - Get a repo by host
func (dao DockerDAOImpl) GetDockerRepositoryByHost(host string) (model.DockerRepo, error) {
	item := model.DockerRepo{}
	if err := dao.Db.Where(&model.DockerRepo{Host: host}).Find(&item).Error; err != nil {
		return item, err
	}
	return item, nil
}

//DeleteDockerRepo - Deletes a docker repo
func (dao DockerDAOImpl) DeleteDockerRepo(id int) error {
	if err := dao.Db.Unscoped().Delete(model.DockerRepo{}, id).Error; err != nil {
		return err
	}
	return nil
}

//ListDockerRepos - List docker repos
func (dao DockerDAOImpl) ListDockerRepos() ([]model.DockerRepo, error) {
	list := make([]model.DockerRepo, 0)
	if err := dao.Db.Find(&list).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model.DockerRepo, 0), nil
		}
		return nil, err
	}
	return list, nil
}
