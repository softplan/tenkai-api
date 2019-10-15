package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/dbms/model"
)

//DependencyDAOInterface DependencyDAOInterface
type DependencyDAOInterface interface {
	GetDependencies(chartName string, tag string) ([]model.Dependency, error)
	CreateDependency(dependency model.Dependency) error
	DeleteDependency(id int) error
	ListDependencies(releaseID int) ([]model.Dependency, error)
}

//DependencyDAOImpl DependencyDAOImpl
type DependencyDAOImpl struct {
	Db *gorm.DB
}

//GetDependencies - Retrieve dependencies
func (dao DependencyDAOImpl) GetDependencies(chartName string, tag string) ([]model.Dependency, error) {
	var release model.Release
	if err := dao.Db.Where(&model.Release{ChartName: chartName, Release: tag}).First(&release).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model.Dependency, 0), nil
		}
		return nil, err
	}
	dependencies, err := dao.ListDependencies(int(release.ID))
	return dependencies, err
}

//CreateDependency - Creates a new dependency
func (dao DependencyDAOImpl) CreateDependency(dependency model.Dependency) error {
	if err := dao.Db.Create(&dependency).Error; err != nil {
		return err
	}
	return nil
}

//DeleteDependency - Deletes a dependency
func (dao DependencyDAOImpl) DeleteDependency(id int) error {
	if err := dao.Db.Unscoped().Delete(model.Dependency{}, id).Error; err != nil {
		return err
	}
	return nil
}

//ListDependencies - List dependencies
func (dao DependencyDAOImpl) ListDependencies(releaseID int) ([]model.Dependency, error) {
	dependencies := make([]model.Dependency, 0)
	if err := dao.Db.Where(&model.Dependency{ReleaseID: releaseID}).Find(&dependencies).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model.Dependency, 0), nil
		}
		return nil, err
	}
	return dependencies, nil
}
