package dbms

import (
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/dbms/model"
)

//EnvironmentDAOInterface EnvironmentDAOInterface
type EnvironmentDAOInterface interface {
	CreateEnvironment(env model.Environment) (int, error)
	EditEnvironment(env model.Environment) error
	DeleteEnvironment(env model.Environment) error
	GetAllEnvironments(principal string) ([]model.Environment, error)
	GetByID(envID int) (*model.Environment, error)
}

//EnvironmentDAOImpl EnvironmentDAOImpl
type EnvironmentDAOImpl struct {
	Db *gorm.DB
}

//CreateEnvironment - Create a new environment
func (dao EnvironmentDAOImpl) CreateEnvironment(env model.Environment) (int, error) {
	if err := dao.Db.Create(&env).Error; err != nil {
		return -1, err
	}
	return int(env.ID), nil
}

// EditEnvironment - Updates an existing environment
func (dao EnvironmentDAOImpl) EditEnvironment(env model.Environment) error {
	if err := dao.Db.Save(&env).Error; err != nil {
		return err
	}
	return nil
}

// DeleteEnvironment - Deletes an environment
func (dao EnvironmentDAOImpl) DeleteEnvironment(env model.Environment) error {
	if err := dao.Db.Unscoped().Delete(&env).Error; err != nil {
		return err
	}
	return nil
}

//GetAllEnvironments - Retrieve all environments
func (dao EnvironmentDAOImpl) GetAllEnvironments(principal string) ([]model.Environment, error) {
	envs := make([]model.Environment, 0)

	if len(principal) > 0 {
		//Find User by email
		var user model.User

		if err := dao.Db.Where(model.User{Email: principal}).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return envs, nil
			}
			return nil, err
		}

		if err := dao.Db.Model(&user).Related(&envs, "Environments").Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return envs, nil
			}
			return nil, err
		}
	} else {
		if err := dao.Db.Find(&envs).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return envs, nil
			}
			return nil, err
		}
	}
	return envs, nil
}

//GetByID - Get Environment By Id
func (dao EnvironmentDAOImpl) GetByID(envID int) (*model.Environment, error) {
	var result model.Environment
	if err := dao.Db.First(&result, envID).Error; err != nil {
		return nil, err
	}
	return &result, nil
}
