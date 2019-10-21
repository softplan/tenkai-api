package repository

import (
	"github.com/jinzhu/gorm"
	model2 "github.com/softplan/tenkai-api/pkg/dbms/model"
)

//EnvironmentDAOInterface EnvironmentDAOInterface
type EnvironmentDAOInterface interface {
	CreateEnvironment(env model2.Environment) (int, error)
	EditEnvironment(env model2.Environment) error
	DeleteEnvironment(env model2.Environment) error
	GetAllEnvironments(principal string) ([]model2.Environment, error)
	GetByID(envID int) (*model2.Environment, error)
}

//EnvironmentDAOImpl EnvironmentDAOImpl
type EnvironmentDAOImpl struct {
	Db *gorm.DB
}

//CreateEnvironment - Create a new environment
func (dao EnvironmentDAOImpl) CreateEnvironment(env model2.Environment) (int, error) {
	if err := dao.Db.Create(&env).Error; err != nil {
		return -1, err
	}
	return int(env.ID), nil
}

// EditEnvironment - Updates an existing environment
func (dao EnvironmentDAOImpl) EditEnvironment(env model2.Environment) error {
	if err := dao.Db.Save(&env).Error; err != nil {
		return err
	}
	return nil
}

// DeleteEnvironment - Deletes an environment
func (dao EnvironmentDAOImpl) DeleteEnvironment(env model2.Environment) error {
	if err := dao.Db.Unscoped().Delete(&env).Error; err != nil {
		return err
	}
	return nil
}

//GetAllEnvironments - Retrieve all environments
func (dao EnvironmentDAOImpl) GetAllEnvironments(principal string) ([]model2.Environment, error) {
	envs := make([]model2.Environment, 0)
	if len(principal) > 0 {
		var user model2.User
		if err := dao.Db.Where(model2.User{Email: principal}).First(&user).Error; err != nil {
			return checkNotFound(err)
		}

		if err := dao.Db.Model(&user).Related(&envs, "Environments").Error; err != nil {
			return checkNotFound(err)
		}
	} else {
		if err := dao.Db.Find(&envs).Error; err != nil {
			return checkNotFound(err)
		}
	}
	return envs, nil
}

//GetByID - Get Environment By Id
func (dao EnvironmentDAOImpl) GetByID(envID int) (*model2.Environment, error) {
	var result model2.Environment
	if err := dao.Db.First(&result, envID).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func checkNotFound(err error) ([]model2.Environment, error) {
	if err == gorm.ErrRecordNotFound {
		return make([]model2.Environment, 0), nil
	} else {
		return nil, err
	}
}
