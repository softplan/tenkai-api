package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
)

//DeploymentDAOInterface DeploymentDAOInterface
type DeploymentDAOInterface interface {
	CreateDeployment(deployment model.Deployment) (int, error)
	EditDeployment(deployment model.Deployment) (error)
	ListDeploymentByUserID(userID int) ([]model.Deployment, error)
	GetDeploymentByID(id int) (model.Deployment, error)
}

//DeploymentDAOImpl DeploymentDAOImpl
type DeploymentDAOImpl struct {
	Db *gorm.DB
}

//GetDeploymentByID GetDeploymentByID
func (dao DeploymentDAOImpl) GetDeploymentByID(id int) (model.Deployment, error) {
	var deployment model.Deployment
	if err := dao.Db.First(&deployment, id).Error; err != nil {
		return model.Deployment{}, err
	}
	return deployment, nil
}

//CreateDeployment create deployment
func (dao DeploymentDAOImpl) CreateDeployment(deployment model.Deployment) (int, error) {
	if err := dao.Db.Create(&deployment).Error; err != nil {
		return -1, err
	}
	return int(deployment.ID), nil
}

//EditDeployment edit deployment
func (dao DeploymentDAOImpl) EditDeployment(deployment model.Deployment) (error) {
	return dao.Db.Save(&deployment).Error
}

//ListDeploymentByUserID list all deployments
func (dao DeploymentDAOImpl) ListDeploymentByUserID(userID int) ([]model.Deployment, error) {
	deployments := make([]model.Deployment, 0)
	err := dao.Db.Find(&deployments).Error
	if err != nil {
		return deployments, nil
	}
	return deployments, err
}