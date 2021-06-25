package repository

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
)

//RequestDeploymentDAOInterface RequestDeploymentDAOInterface
type RequestDeploymentDAOInterface interface {
	CreateRequestDeployment(deployment model.RequestDeployment) (int, error)
	EditRequestDeployment(rd model.RequestDeployment) error
	GetRequestDeploymentByID(id int) (model.RequestDeployment, error)
	ListRequestDeployments(startDate, endDate, environmentID, userID string, id, pageNumber, pageSize int) ([]model.RequestDeployments, error)
	CountRequestDeployments(startDate, endDate, environmentID, userID string) (int64, error)
	CheckIfRequestHasEnded(id int) (bool, error)
	HasErrorInRequest(id int) (bool, error)
}

//RequestDeploymentDAOImpl RequestDeploymentDAOImpl
type RequestDeploymentDAOImpl struct {
	Db *gorm.DB
}

//CreateRequestDeployment CreateRequestDeployment
func (dao RequestDeploymentDAOImpl) CreateRequestDeployment(rd model.RequestDeployment) (int, error) {
	err := dao.Db.Create(&rd).Error
	if err != nil {
		return -1, err
	}
	return int(rd.ID), nil
}

//GetRequestDeploymentByID GetRequestDeploymentByID
func (dao RequestDeploymentDAOImpl) GetRequestDeploymentByID(id int) (model.RequestDeployment, error) {
	var rd model.RequestDeployment
	if err := dao.Db.First(&rd, id).Error; err != nil {
		return model.RequestDeployment{}, err
	}
	return rd, nil
}

//EditRequestDeployment edit requestDeployment
func (dao RequestDeploymentDAOImpl) EditRequestDeployment(rd model.RequestDeployment) error {
	gorm := dao.Db.Save(&rd)
	return gorm.Error
}

//CheckIfRequestHasEnded verify if all deployments has ended
func (dao RequestDeploymentDAOImpl) CheckIfRequestHasEnded(id int) (bool, error) {
	var deployment model.Deployment
	var count = -1
	err := dao.Db.Where(
		"request_deployment_id = ? AND processed = ?",
		fmt.Sprint(id),
		false,
	).Model(
		&deployment,
	).Count(&count).Error
	if err != nil {
		return false, nil
	}
	if count == 0 {
		return true, nil
	}
	return false, err
}

//HasErrorInRequest verify if some deployment had some error
func (dao RequestDeploymentDAOImpl) HasErrorInRequest(id int) (bool, error) {
	var deployment model.Deployment
	var count = -1
	err := dao.Db.Where(
		"request_deployment_id = ? AND success = ?",
		id,
		false,
	).Model(
		&deployment,
	).Count(&count).Error
	if err != nil {
		return false, nil
	}
	if count > 0 {
		return true, nil
	}
	return false, nil
}

//ListRequestDeployments list
func (dao RequestDeploymentDAOImpl) ListRequestDeployments(startDate, endDate, environmentID, userID string, id, pageNumber, pageSize int) ([]model.RequestDeployments, error) {
	var rdList []model.RequestDeployments
	sql := prepareWhere(id, environmentID, userID)
	rows, err := dao.Db.Table("request_deployments").Select(
		"DISTINCT request_deployments.id, request_deployments.created_at, request_deployments.updated_at, request_deployments.processed, request_deployments.success, users.email as email",
	).Joins(
		"JOIN deployments ON deployments.request_deployment_id = request_deployments.id",
	).Joins(
		"JOIN users ON users.id = request_deployments.user_id",
	).Where(sql, startDate, endDate).Offset((pageNumber - 1) * pageSize).Limit(pageSize).Rows()

	for rows.Next() {
		id, userID := 0, 0
		createdAt := time.Time{}
		updatedAt := time.Time{}
		success, processed := false, false
		email := ""
		rows.Scan(&id, &createdAt, &updatedAt, &processed, &success, &userID, &email)

		request := model.RequestDeployments{}
		request.ID = uint(id)
		request.CreatedAt = createdAt
		request.UpdatedAt = updatedAt
		request.Processed = processed
		request.Success = success
		request.User = email

		rdList = append(rdList, request)
	}

	return rdList, err
}

//CountRequestDeployments count
func (dao RequestDeploymentDAOImpl) CountRequestDeployments(startDate, endDate, environmentID, userID string) (int64, error) {
	var deployment model.RequestDeployment
	var count int64
	sql := prepareWhere(-1, environmentID, userID)
	rows, err := dao.Db.Model(&deployment).Select("COUNT(DISTINCT request_deployments.id) AS total").Joins(
		"JOIN deployments ON deployments.request_deployment_id = request_deployments.id",
	).Joins(
		"JOIN users ON users.id = request_deployments.user_id",
	).Where(sql, startDate, endDate).Rows()

	for rows.Next() {
		rows.Scan(&count)
	}
	return count, err
}

func prepareWhere(id int, environmentID, userID string) string {
	where := "date(request_deployments.created_at) >= ? AND date(request_deployments.created_at) <= ?"
	if id != -1 {
		where = where + " AND request_deployments.id = " + fmt.Sprint(id)
	}
	if environmentID != "" {
		where += " AND deployments.environment_id = " + environmentID
	}
	if userID != "" {
		where += " AND request_deployments.user_id = " + userID
	}
	return where
}
