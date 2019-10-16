package repository

import (
	"github.com/jinzhu/gorm"
	model2 "github.com/softplan/tenkai-api/pkg/dbms/model"
)

//ReleaseDAOInterface ReleaseDAOInterface
type ReleaseDAOInterface interface {
	CreateRelease(release model2.Release) error
	DeleteRelease(id int) error
	ListRelease(chartName string) ([]model2.Release, error)
}

//ReleaseDAOImpl ReleaseDAOImpl
type ReleaseDAOImpl struct {
	Db *gorm.DB
}

//CreateRelease - Create a new Release
func (dao ReleaseDAOImpl) CreateRelease(release model2.Release) error {
	if err := dao.Db.Create(&release).Error; err != nil {
		return err
	}
	return nil
}

//DeleteRelease - Delete a new Release
func (dao ReleaseDAOImpl) DeleteRelease(id int) error {
	if err := dao.Db.Unscoped().Delete(model2.Release{}, id).Error; err != nil {
		return err
	}
	return nil
}

//ListRelease - List releases
func (dao ReleaseDAOImpl) ListRelease(chartName string) ([]model2.Release, error) {
	releases := make([]model2.Release, 0)
	if err := dao.Db.Where(&model2.Release{ChartName: chartName}).Find(&releases).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		return nil, err
	}
	return releases, nil
}
