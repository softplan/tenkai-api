package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
)

//CompareEnvsQueryDAOInterface CompareEnvsQueryDAOInterface
type CompareEnvsQueryDAOInterface interface {
	SaveCompareEnvsQuery(env model.CompareEnvsQuery) (int, error)
	DeleteCompareEnvQuery(id int) error
	GetByUser(userID int) ([]model.CompareEnvsQuery, error)
}

//CompareEnvsQueryDAOImpl CompareEnvsQueryDAOImpl
type CompareEnvsQueryDAOImpl struct {
	Db *gorm.DB
}

//SaveCompareEnvsQuery Create or Update
func (dao CompareEnvsQueryDAOImpl) SaveCompareEnvsQuery(compareEnvQuery model.CompareEnvsQuery) (int, error) {
	if err := dao.Db.Save(&compareEnvQuery).Error; err != nil {
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

//DeleteCompareEnvQuery DeleteCompareEnvQuery
func (dao CompareEnvsQueryDAOImpl) DeleteCompareEnvQuery(id int) error {
	return dao.Db.Unscoped().Delete(model.CompareEnvsQuery{}, id).Error
}
