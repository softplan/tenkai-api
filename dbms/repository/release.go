package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/dbms/model"
)

//ReleaseDAOInterface ReleaseDAOInterface
type ReleaseDAOInterface interface {
	CreateRelease(release model.Release) error
	DeleteRelease(id int) error
	ListRelease(chartName string) ([]model.Release, error)
}

//ReleaseDAOImpl ReleaseDAOImpl
type ReleaseDAOImpl struct {
	Db *gorm.DB
}

//CreateRelease - Create a new Release
func (dao ReleaseDAOImpl) CreateRelease(release model.Release) error {
	if err := dao.Db.Create(&release).Error; err != nil {
		return err
	}
	return nil
}

//DeleteRelease - Delete a new Release
func (dao ReleaseDAOImpl) DeleteRelease(id int) error {
	if err := dao.Db.Unscoped().Delete(model.Release{}, id).Error; err != nil {
		return err
	}
	return nil
}

//ListRelease - List releases
func (dao ReleaseDAOImpl) ListRelease(chartName string) ([]model.Release, error) {
	releases := make([]model.Release, 0)
	if err := dao.Db.Where(&model.Release{ChartName: chartName}).Find(&releases).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, err
	}
	return releases, nil
}
