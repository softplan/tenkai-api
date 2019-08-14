package dbms

import (
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/dbms/model"
)

//CreateSolution - Create a new solution
func (database *Database) CreateSolution(solution model.Solution) (int, error) {
	if err := database.Db.Create(&solution).Error; err != nil {
		return -1, err
	}
	return int(solution.ID), nil
}

//EditSolution - Updates an existing solution
func (database *Database) EditSolution(solution model.Solution) error {
	if err := database.Db.Save(&solution).Error; err != nil {
		return err
	}
	return nil
}

//DeleteSolution - Deletes a solution
func (database *Database) DeleteSolution(id int) error {
	if err := database.Db.Unscoped().Delete(model.Solution{}, id).Error; err != nil {
		return err
	}
	return nil
}

//ListSolutions - List solutions
func (database *Database) ListSolutions() ([]model.Solution, error) {
	list := make([]model.Solution, 0)
	if err := database.Db.Find(&list).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model.Solution, 0), nil
		}
		return nil, err
	}
	return list, nil
}
