package dbms

import (
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/dbms/model"
)

//CreateDependency - Creates a new dependency
func (database *Database) CreateDependency(dependency model.Dependency) error {
	if err := database.Db.Create(&dependency).Error; err != nil {
		return err
	}
	return nil
}

//DeleteDependency - Deletes a dependency
func (database *Database) DeleteDependency(id int) error {
	if err := database.Db.Unscoped().Delete(model.Dependency{}, id).Error; err != nil {
		return err
	}
	return nil
}

//ListDependencies - List dependencies
func (database *Database) ListDependencies(releaseID int) ([]model.Dependency, error) {
	dependencies := make([]model.Dependency, 0)
	if err := database.Db.Where(&model.Dependency{ReleaseID: releaseID}).Find(&dependencies).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model.Dependency, 0), nil
		}
		return nil, err
	}
	return dependencies, nil
}
