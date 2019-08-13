package dbms

import (
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/dbms/model"
)


//GetDependencies - Retrieve dependencies
func (database *Database) GetDependencies(chartName string, tag string) ([]model.Dependency, error) {
	var release model.Release
	if err := database.Db.Where(&model.Release{ChartName: chartName, Release: tag}).First(&release).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model.Dependency, 0), nil
		}
		return nil, err
	}
	dependencies, err := database.ListDependencies(int(release.ID))
	return dependencies, err
}
