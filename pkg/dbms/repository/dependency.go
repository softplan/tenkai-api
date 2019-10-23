package repository

import (
	"github.com/jinzhu/gorm"
	model2 "github.com/softplan/tenkai-api/pkg/dbms/model"
)

//DependencyDAOInterface DependencyDAOInterface
type DependencyDAOInterface interface {
	GetDependencies(chartName string, tag string) ([]model2.Dependency, error)
	CreateDependency(dependency model2.Dependency) error
	DeleteDependency(id int) error
	ListDependencies(releaseID int) ([]model2.Dependency, error)
}

//DependencyDAOImpl DependencyDAOImpl
type DependencyDAOImpl struct {
	Db *gorm.DB
}

//GetDependencies - Retrieve dependencies
func (dao DependencyDAOImpl) GetDependencies(chartName string, tag string) ([]model2.Dependency, error) {
	var release model2.Release
	if err := dao.Db.Where(&model2.Release{ChartName: chartName, Release: tag}).First(&release).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model2.Dependency, 0), nil
		}
		return nil, err
	}
	dependencies, err := dao.ListDependencies(int(release.ID))
	return dependencies, err
}

//CreateDependency - Creates a new dependency
func (dao DependencyDAOImpl) CreateDependency(dependency model2.Dependency) error {
	return dao.Db.Create(&dependency).Error
}

//DeleteDependency - Deletes a dependency
func (dao DependencyDAOImpl) DeleteDependency(id int) error {
	return dao.Db.Unscoped().Delete(model2.Dependency{}, id).Error
}

//ListDependencies - List dependencies
func (dao DependencyDAOImpl) ListDependencies(releaseID int) ([]model2.Dependency, error) {
	dependencies := make([]model2.Dependency, 0)
	if err := dao.Db.Where(&model2.Dependency{ReleaseID: releaseID}).Find(&dependencies).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return make([]model2.Dependency, 0), nil
		}
		return nil, err
	}
	return dependencies, nil
}
