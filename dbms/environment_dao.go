package dbms

import (
	"github.com/softplan/tenkai-api/dbms/model"
)

//CreateEnvironment - Create a new environment
func (database *Database) CreateEnvironment(env model.Environment) (int, error) {
	if err := database.Db.Create(&env).Error; err != nil {
		return -1, err
	}
	return int(env.ID), nil
}

// EditEnvironment - Updates an existing environment
func (database *Database) EditEnvironment(env model.Environment) error {
	if err := database.Db.Save(&env).Error; err != nil {
		return err
	}
	return nil
}

// DeleteEnvironment - Deletes an environment
func (database *Database) DeleteEnvironment(env model.Environment) error {
	if err := database.Db.Delete(&env).Error; err != nil {
		return err
	}
	return nil
}

//GetAllEnvironments - Retrieve all environments
func (database *Database) GetAllEnvironments(principal string) ([]model.Environment, error) {
	envs := make([]model.Environment, 0)

	if len(principal) > 0 {
		//Find User by email
		var user model.User

		if err := database.Db.Where(model.User{Email: principal}).First(&user).Error; err != nil {
			return nil, err
		}

		if err := database.Db.Model(&user).Related(&envs, "Environments").Error; err != nil {
			return nil, err
		}
	} else {
		if err := database.Db.Find(&envs).Error; err != nil {
			return nil, err
		}
	}
	return envs, nil
}

//GetByID - Get Environment By Id
func (database *Database) GetByID(envID int) (*model.Environment, error) {
	var result model.Environment
	if err := database.Db.First(&result, envID).Error; err != nil {
		return nil, err
	}
	return &result, nil
}
