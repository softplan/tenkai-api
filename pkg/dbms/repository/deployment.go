package repository

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
)

//DeploymentDAOInterface DeploymentDAOInterface
type DeploymentDAOInterface interface {
	CreateDeployment(deployment model.Deployment) (int, error)
	EditDeployment(deployment model.Deployment) error
	GetDeploymentByID(id int) (model.Deployment, error)
	ListDeployments(startDate, endDate, userID, environmentID string, pageNumber, pageSize int) ([]model.Deployments, error)
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
	sql := "deployments.created_at >= ? AND deployments.created_at <= ?"
	if userID != "" {
		sql += " AND deployments.user_id = " + userID
	}
	if environmentID != "" {
		sql += " AND deployments.environment_id = " + environmentID
	}
	return sql
}

//ListDeployments list all deployments filtered by date, environment and user
func (dao DeploymentDAOImpl) ListDeployments(startDate, endDate, userID, environmentID string, pageNumber, pageSize int) ([]model.Deployments, error) {
	var deployments []model.Deployments
	sql := prepareSQL(userID, environmentID)
	rows, err := dao.Db.Table("deployments").Select(
		"deployments.id AS id, deployments.created_at AS created_at, deployments.updated_at AS updated_at,chart, users.id AS user_id, users.email AS user_email, environments.id AS environments_id, environments.name AS environments_name, success, message ",
	).Joins(
		"JOIN users ON deployments.user_id = users.id",
	).Joins(
		"JOIN environments ON deployments.environment_id = environments.id",
	).Where(sql, startDate, endDate).Offset((pageNumber - 1) * pageSize).Limit(pageSize).Rows()

	for rows.Next() {
		userID, envID, id := 0, 0, 0
		createdAt := time.Time{}
		updatedAt := time.Time{}
		chart, userEmail, envName, message := "", "", "", ""
		success := false
		rows.Scan(&id, &createdAt, &updatedAt, &chart, &userID, &userEmail, &envID, &envName, &success, &message)

		deployment := model.Deployments{}
		deployment.ID = uint(id)
		deployment.CreatedAt = createdAt
		deployment.UpdatedAt = updatedAt
		deployment.ChartName = chart
		deployment.Environment.ID = uint(envID)
		deployment.Environment.Name = envName
		deployment.User.ID = uint(userID)
		deployment.User.Email = userEmail
		deployment.SuccessDeployment = success
		deployment.ErrorMessage = message

		deployments = append(deployments, deployment)
	}

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
