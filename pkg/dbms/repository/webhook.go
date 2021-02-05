package repository

import (
	"github.com/jinzhu/gorm"
	model "github.com/softplan/tenkai-api/pkg/dbms/model"
)

//WebHookDAOInterface WebHookDAOInterface
type WebHookDAOInterface interface {
	CreateWebHook(e model.WebHook) (int, error)
	EditWebHook(e model.WebHook) error
	DeleteWebHook(id int) error
	ListWebHooks() ([]model.WebHook, error)
	ListWebHooksByEnvAndType(environmentID int, hookType string) ([]model.WebHook, error)
}

//WebHookDAOImpl WebHookDAOImpl
type WebHookDAOImpl struct {
	Db *gorm.DB
}

//CreateWebHook - Create a new webhook
func (dao WebHookDAOImpl) CreateWebHook(e model.WebHook) (int, error) {
	if err := dao.Db.Create(&e).Error; err != nil {
		return -1, err
	}
	return int(e.ID), nil
}

//EditWebHook - Updates an existing webhook
func (dao WebHookDAOImpl) EditWebHook(e model.WebHook) error {
	return dao.Db.Save(&e).Error
}

//DeleteWebHook - Deletes a webhook
func (dao WebHookDAOImpl) DeleteWebHook(id int) error {
	return dao.Db.Unscoped().Delete(model.WebHook{}, id).Error
}

//ListWebHooks - List webhooks
func (dao WebHookDAOImpl) ListWebHooks() ([]model.WebHook, error) {
	list := make([]model.WebHook, 0)
	if err := dao.Db.Find(&list).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model.WebHook, 0), nil
		}
		return nil, err
	}
	return list, nil
}

//ListWebHooksByEnvAndType - List webhooks by environment and webHook type
func (dao WebHookDAOImpl) ListWebHooksByEnvAndType(
	environmentID int, hookType string) ([]model.WebHook, error) {

	list := make([]model.WebHook, 0)
	var condition *model.WebHook
	if environmentID != -1 {
		condition = &model.WebHook{EnvironmentID: environmentID, Type: hookType}
	} else {
		condition = &model.WebHook{Type: hookType}
	}

	if err := dao.Db.Where(condition).Find(&list).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model.WebHook, 0), nil
		}
		return nil, err
	}
	return list, nil
}
