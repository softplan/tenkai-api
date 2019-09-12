package dbms

import (
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/dbms/model"
)

//CreateDockerRepo - Create a new docker repo
func (database *Database) CreateDockerRepo(item model.DockerRepo) (int, error) {
	if err := database.Db.Create(&item).Error; err != nil {
		return -1, err
	}
	return int(item.ID), nil
}

//GetDockerRepositoryByHost - Get a repo by host
func (database *Database) GetDockerRepositoryByHost(host string) (model.DockerRepo, error) {
	item := model.DockerRepo{}
	if err := database.Db.Where(&model.DockerRepo{Host: host}).Find(&item).Error; err != nil {
		return item, err
	}
	return item, nil
}

//DeleteDockerRepo - Deletes a docker repo
func (database *Database) DeleteDockerRepo(id int) error {
	if err := database.Db.Unscoped().Delete(model.DockerRepo{}, id).Error; err != nil {
		return err
	}
	return nil
}

//ListDockerRepos - List docker repos
func (database *Database) ListDockerRepos() ([]model.DockerRepo, error) {
	list := make([]model.DockerRepo, 0)
	if err := database.Db.Find(&list).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model.DockerRepo, 0), nil
		}
		return nil, err
	}
	return list, nil
}
