package repository

import (
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
)

//DeploymentDAOInterface DeploymentDAOInterface
type DeploymentDAOInterface interface {
	CreateDeployment(deployment model.Deployment) (int, error)
	EditDeployment(deployment model.Deployment) error
	GetDeploymentByID(id int) (model.Deployment, error)
	ListDeployments(startDate, endDate, userID, environmentID string, pageNumber, pageSize int) ([]model.Deployment, error)
	CountDeployments(startDate, endDate, userID, environmentID string) (int64, error)
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
func (dao DeploymentDAOImpl) EditDeployment(deployment model.Deployment) error {
	gorm := dao.Db.Save(&deployment)
	return gorm.Error
}

func prepareSQL(userID, environmentID string) string {
	sql := "created_at >= ? AND created_at <= ?"
	if userID != "" {
		sql += " AND user_id = " + userID
	}
	if environmentID != "" {
		sql += " AND environment_id = " + environmentID
	}
	return sql
}

//ListDeployments list all deployments filtered by date, environment and user
func (dao DeploymentDAOImpl) ListDeployments(startDate, endDate, userID, environmentID string, pageNumber, pageSize int) ([]model.Deployment, error) {
	var deployments []model.Deployment
	sql := prepareSQL(userID, environmentID)
	err := dao.Db.Where(sql, startDate, endDate).Offset((pageNumber - 1) * pageSize).Limit(pageSize).Find(&deployments).Error
	return deployments, err
}

//CountDeployments count
func (dao DeploymentDAOImpl) CountDeployments(startDate, endDate, userID, environmentID string) (int64, error) {
	var deployment model.Deployment
	var count int64
	sql := prepareSQL(userID, environmentID)
	err := dao.Db.Where(sql, startDate, endDate).Model(&deployment).Count(&count).Error
	return count, err
}
