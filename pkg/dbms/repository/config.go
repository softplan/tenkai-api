package repository

import (
	"github.com/jinzhu/gorm"
	model2 "github.com/softplan/tenkai-api/pkg/dbms/model"
)

//ConfigDAOInterface ConfigDAOInterface
type ConfigDAOInterface interface {
	CreateOrUpdateConfig(item model2.ConfigMap) (int, error)
	GetConfigByName(name string) (model2.ConfigMap, error)
}

//ConfigDAOImpl ConfigDAOImpl
type ConfigDAOImpl struct {
	Db *gorm.DB
}

//CreateOrUpdateConfig - Create or update a new config
func (dao ConfigDAOImpl) CreateOrUpdateConfig(item model2.ConfigMap) (int, error) {
	var result model2.ConfigMap

	edit := true
	if err := dao.Db.Where(&model2.ConfigMap{Name: item.Name}).Find(&result).Error; err != nil {
		edit = false
		if !gorm.IsRecordNotFoundError(err) {
			return -1, err
		}
	}

	if edit {
		result.Value = item.Value
		if err := dao.Db.Save(&result).Error; err != nil {
			return -1, err
		}
		return int(result.ID), nil
	}

	if err := dao.Db.Create(&item).Error; err != nil {
		return -1, err
	}
	return int(item.ID), nil

}

//GetConfigByName - Retrieves a config by name
func (dao ConfigDAOImpl) GetConfigByName(name string) (model2.ConfigMap, error) {
	var result model2.ConfigMap
	if err := dao.Db.Where(&model2.ConfigMap{Name: name}).Find(&result).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return result, err
		}
		return result, nil
	}
	return result, nil
}
