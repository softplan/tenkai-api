package dbms

import (
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/dbms/model"
)

func (database *Database) CreateDependency(dependency model.Dependency) error {
	if err := database.Db.Create(&dependency ).Error; err != nil {
		return err
	}
	return nil
}

func (database *Database) DeleteDependency(id int) error {
	if err := database.Db.Unscoped().Delete(model.Dependency{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (database *Database) ListDependencies(releaseId int) ([]model.Dependency, error) {
	dependencies := make([]model.Dependency, 0)
	if err := database.Db.Where(&model.Dependency{ReleaseID: releaseId}).Find(&dependencies).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, err
		} else {
			return nil, err
		}
	}
	return dependencies, nil
}