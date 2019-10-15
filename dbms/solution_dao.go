package dbms

import (
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/dbms/model"
)

//SolutionDAOInterface SolutionDAOInterface
type SolutionDAOInterface interface {
	CreateSolution(solution model.Solution) (int, error)
	EditSolution(solution model.Solution) error
	DeleteSolution(id int) error
	ListSolutions() ([]model.Solution, error)
}

//SolutionDAOImpl SolutionDAOImpl
type SolutionDAOImpl struct {
	Db *gorm.DB
}

//CreateSolution - Create a new solution
func (dao SolutionDAOImpl) CreateSolution(solution model.Solution) (int, error) {
	if err := dao.Db.Create(&solution).Error; err != nil {
		return -1, err
	}
	return int(solution.ID), nil
}

//EditSolution - Updates an existing solution
func (dao SolutionDAOImpl) EditSolution(solution model.Solution) error {
	if err := dao.Db.Save(&solution).Error; err != nil {
		return err
	}
	return nil
}

//DeleteSolution - Deletes a solution
func (dao SolutionDAOImpl) DeleteSolution(id int) error {
	if err := dao.Db.Unscoped().Delete(model.Solution{}, id).Error; err != nil {
		return err
	}
	return nil
}

//ListSolutions - List solutions
func (dao SolutionDAOImpl) ListSolutions() ([]model.Solution, error) {
	list := make([]model.Solution, 0)
	if err := dao.Db.Find(&list).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model.Solution, 0), nil
		}
		return nil, err
	}
	return list, nil
}
