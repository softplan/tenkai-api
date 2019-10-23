package repository

import (
	"github.com/jinzhu/gorm"
	model2 "github.com/softplan/tenkai-api/pkg/dbms/model"
)

//SolutionDAOInterface SolutionDAOInterface
type SolutionDAOInterface interface {
	CreateSolution(solution model2.Solution) (int, error)
	EditSolution(solution model2.Solution) error
	DeleteSolution(id int) error
	ListSolutions() ([]model2.Solution, error)
}

//SolutionDAOImpl SolutionDAOImpl
type SolutionDAOImpl struct {
	Db *gorm.DB
}

//CreateSolution - Create a new solution
func (dao SolutionDAOImpl) CreateSolution(solution model2.Solution) (int, error) {
	if err := dao.Db.Create(&solution).Error; err != nil {
		return -1, err
	}
	return int(solution.ID), nil
}

//EditSolution - Updates an existing solution
func (dao SolutionDAOImpl) EditSolution(solution model2.Solution) error {
	return dao.Db.Save(&solution).Error
}

//DeleteSolution - Deletes a solution
func (dao SolutionDAOImpl) DeleteSolution(id int) error {
	return dao.Db.Unscoped().Delete(model2.Solution{}, id).Error
}

//ListSolutions - List solutions
func (dao SolutionDAOImpl) ListSolutions() ([]model2.Solution, error) {
	list := make([]model2.Solution, 0)
	if err := dao.Db.Find(&list).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model2.Solution, 0), nil
		}
		return nil, err
	}
	return list, nil
}
