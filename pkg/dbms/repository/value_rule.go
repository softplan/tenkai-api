package repository

import (
	"github.com/jinzhu/gorm"
	model "github.com/softplan/tenkai-api/pkg/dbms/model"
)

//ValueRuleDAOInterface ValueRuleDAOInterface
type ValueRuleDAOInterface interface {
	CreateValueRule(e model.ValueRule) (int, error)
	EditValueRule(e model.ValueRule) error
	DeleteValueRule(id int) error
	ListValueRules(variableRuleID int) ([]model.ValueRule, error)
}

//ValueRuleDAOImpl ValueRuleDAOImpl
type ValueRuleDAOImpl struct {
	Db *gorm.DB
}

//CreateValueRule - Create a new value rule
func (dao ValueRuleDAOImpl) CreateValueRule(e model.ValueRule) (int, error) {
	if err := dao.Db.Create(&e).Error; err != nil {
		return -1, err
	}
	return int(e.ID), nil
}

//EditValueRule - Updates an existing value rule
func (dao ValueRuleDAOImpl) EditValueRule(e model.ValueRule) error {
	return dao.Db.Save(&e).Error
}

//DeleteValueRule - Deletes a value rule
func (dao ValueRuleDAOImpl) DeleteValueRule(id int) error {
	return dao.Db.Unscoped().Delete(model.ValueRule{}, id).Error
}

//ListValueRules - List value rules
func (dao ValueRuleDAOImpl) ListValueRules(variableRuleID int) ([]model.ValueRule, error) {
	list := make([]model.ValueRule, 0)
	if err := dao.Db.Where(&model.ValueRule{VariableRuleID: uint(variableRuleID)}).Find(&list).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model.ValueRule, 0), nil
		}
		return nil, err
	}
	return list, nil
}
