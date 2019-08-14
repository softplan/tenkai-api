package dbms

import (
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/dbms/model"
)

//CreateOrUpdateConfig - Create or update a new config
func (database *Database) CreateOrUpdateConfig(item model.ConfigMap) (int, error) {
	var result model.ConfigMap

	edit := true
	if err := database.Db.Where(&model.ConfigMap{Name: item.Name}).Find(&result).Error; err != nil {
		edit = false
		if !gorm.IsRecordNotFoundError(err) {
			return -1, err
		}
	}

	if edit {
		result.Value = item.Value
		if err := database.Db.Save(&result).Error; err != nil {
			return -1, err
		}
		return int(result.ID), nil
	}

	if err := database.Db.Create(&item).Error; err != nil {
		return -1, err
	}
	return int(item.ID), nil

}

//GetConfigByName - Retrieves a config by name
func (database *Database) GetConfigByName(name string) (model.ConfigMap, error) {
	var result model.ConfigMap
	if err := database.Db.Where(&model.ConfigMap{Name: name}).Find(&result).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return result, err
		}
		return result, nil
	}
	return result, nil
}
