package repository

import (
	"github.com/jinzhu/gorm"
	model2 "github.com/softplan/tenkai-api/pkg/dbms/model"
)

//SolutionChartDAOInterface SolutionChartDAOInterface
type SolutionChartDAOInterface interface {
	CreateSolutionChart(element model2.SolutionChart) error
	DeleteSolutionChart(id int) error
	ListSolutionChart(id int) ([]model2.SolutionChart, error)
}

//SolutionChartDAOImpl SolutionChartDAOImpl
type SolutionChartDAOImpl struct {
	Db *gorm.DB
}

//CreateSolutionChart - Create a Solution Chart
func (dao SolutionChartDAOImpl) CreateSolutionChart(element model2.SolutionChart) error {
	if err := dao.Db.Create(&element).Error; err != nil {
		return err
	}
	return nil
}

//DeleteSolutionChart - Delete a solution chart
func (dao SolutionChartDAOImpl) DeleteSolutionChart(id int) error {
	if err := dao.Db.Unscoped().Delete(model2.SolutionChart{}, id).Error; err != nil {
		return err
	}
	return nil
}

//ListSolutionChart - List a Solution Chart
func (dao SolutionChartDAOImpl) ListSolutionChart(id int) ([]model2.SolutionChart, error) {
	list := make([]model2.SolutionChart, 0)
	if err := dao.Db.Where(&model2.SolutionChart{SolutionID: id}).Find(&list).Error; err != nil {

		if gorm.IsRecordNotFoundError(err) {
			return make([]model2.SolutionChart, 0), nil
		}

		return nil, err

	}
	return list, nil
}
