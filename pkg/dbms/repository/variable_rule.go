package repository

import (
	"github.com/jinzhu/gorm"
	model "github.com/softplan/tenkai-api/pkg/dbms/model"
)

//VariableRuleDAOInterface VariableRuleDAOInterface
type VariableRuleDAOInterface interface {
	CreateVariableRule(e model.VariableRule) (int, error)
	EditVariableRule(e model.VariableRule) error
	DeleteVariableRule(id int) error
	ListVariableRules() ([]model.VariableRule, error)
}

//VariableRuleDAOImpl VariableRuleDAOImpl
type VariableRuleDAOImpl struct {
	Db *gorm.DB
}

//CreateVariableRule - Create a new value rule
func (dao VariableRuleDAOImpl) CreateVariableRule(e model.VariableRule) (int, error) {
	if err := dao.Db.Create(&e).Error; err != nil {
		return -1, err
	}
	return int(e.ID), nil
}

//EditVariableRule - Updates an existing value rule
func (dao VariableRuleDAOImpl) EditVariableRule(e model.VariableRule) error {
	return dao.Db.Save(&e).Error
}

//DeleteVariableRule - Deletes a value rule
func (dao VariableRuleDAOImpl) DeleteVariableRule(id int) error {
	return dao.Db.Unscoped().Delete(model.VariableRule{}, id).Error
}

//ListVariableRules - List value rules
func (dao VariableRuleDAOImpl) ListVariableRules() ([]model.VariableRule, error) {
	list := make([]model.VariableRule, 0)
	if err := dao.Db.Find(&list).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model.VariableRule, 0), nil
		}
		return nil, err
	}
	return list, nil
}
