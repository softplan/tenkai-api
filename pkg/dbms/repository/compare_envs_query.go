package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
)

//CompareEnvsQueryDAOInterface CompareEnvsQueryDAOInterface
type CompareEnvsQueryDAOInterface interface {
	CreateCompareEnvsQuery(env model.CompareEnvsQuery) (int, error)
	// EditCompareEnvsQuery(env model.CompareEnvsQuery) error
	// DeleteCompareEnvsQuery(env model.CompareEnvsQuery) error
	GetByUser(userID int) ([]model.CompareEnvsQuery, error)
}

//CompareEnvsQueryDAOImpl CompareEnvsQueryDAOImpl
type CompareEnvsQueryDAOImpl struct {
	Db *gorm.DB
}

//CreateCompareEnvsQuery Create
func (dao CompareEnvsQueryDAOImpl) CreateCompareEnvsQuery(compareEnvQuery model.CompareEnvsQuery) (int, error) {
	if err := dao.Db.Create(&compareEnvQuery).Error; err != nil {
		return -1, err
	}
	return int(compareEnvQuery.ID), nil
}

//GetByUser GetByUser
func (dao CompareEnvsQueryDAOImpl) GetByUser(userID int) ([]model.CompareEnvsQuery, error) {
	list := make([]model.CompareEnvsQuery, 0)
	if err := dao.Db.Where(&model.CompareEnvsQuery{UserID: userID}).Find(&list).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return list, nil
		}
		return nil, err
	}
	return list, nil
}