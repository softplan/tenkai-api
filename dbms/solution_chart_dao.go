package dbms

import (
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/dbms/model"
)

//CreateSolutionChart - Create a Solution Chart
func (database *Database) CreateSolutionChart(element model.SolutionChart) error {
	if err := database.Db.Create(&element).Error; err != nil {
		return err
	}
	return nil
}

//DeleteSolutionChart - Delete a solution chart
func (database *Database) DeleteSolutionChart(id int) error {
	if err := database.Db.Unscoped().Delete(model.SolutionChart{}, id).Error; err != nil {
		return err
	}
	return nil
}

//ListSolutionChart - List a Solution Chart
func (database *Database) ListSolutionChart(id int) ([]model.SolutionChart, error) {
	list := make([]model.SolutionChart, 0)
	if err := database.Db.Where(&model.SolutionChart{SolutionID: id}).Find(&list).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return make([]model.SolutionChart, 0), nil
		}

		return nil, err

	}
	return list, nil
}
