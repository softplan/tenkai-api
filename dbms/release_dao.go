package dbms

import (
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/dbms/model"
)

func (database *Database) CreateRelease(release model.Release) error {
	if err := database.Db.Create(&release).Error; err != nil {
		return err
	}
	return nil
}

func (database *Database) DeleteRelease(id int) error {
	if err := database.Db.Unscoped().Delete(model.Release{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (database *Database) ListRelease(chartName string) ([]model.Release, error) {
	releases := make([]model.Release, 0)
	if err := database.Db.Where(&model.Release{ChartName: chartName}).Find(&releases).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, err
		} else {
			return nil, err
		}
	}
	return releases, nil
}