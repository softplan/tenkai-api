package dbms

import (
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/dbms/model"
)

//CreateRelease - Create a new Release
func (database *Database) CreateRelease(release model.Release) error {
	if err := database.Db.Create(&release).Error; err != nil {
		return err
	}
	return nil
}

//DeleteRelease - Delete a new Release
func (database *Database) DeleteRelease(id int) error {
	if err := database.Db.Unscoped().Delete(model.Release{}, id).Error; err != nil {
		return err
	}
	return nil
}

//ListRelease - List releases
func (database *Database) ListRelease(chartName string) ([]model.Release, error) {
	releases := make([]model.Release, 0)
	if err := database.Db.Where(&model.Release{ChartName: chartName}).Find(&releases).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, err
	}
	return releases, nil
}
