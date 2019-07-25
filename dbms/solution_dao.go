package dbms

import (
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/dbms/model"
)

//CreateEnvironment - Create a new environment
func (database *Database) CreateSolution(solution model.Solution) (int, error) {
	if err := database.Db.Create(&solution).Error; err != nil {
		return -1, err
	}
	return int(solution.ID), nil
}

// EditEnvironment - Updates an existing environment
func (database *Database) EditSolution(solution model.Solution) error {
	if err := database.Db.Save(&solution).Error; err != nil {
		return err
	}
	return nil
}

// DeleteEnvironment - Deletes an environment
func (database *Database) DeleteSolution(id int) error {
	if err := database.Db.Unscoped().Delete(model.Solution{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (database *Database) ListSolutions() ([]model.Solution, error) {
	list := make([]model.Solution, 0)
	if err := database.Db.Find(&list).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model.Solution, 0), nil
		} else {
			return nil, err
		}
	}
	return list, nil
}