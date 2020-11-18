package repository

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
)

//RequestDeploymentDAOInterface RequestDeploymentDAOInterface
type RequestDeploymentDAOInterface interface {
	CreateRequestDeployment(deployment model.RequestDeployment) (int, error)
	EditRequestDeployment(rd model.RequestDeployment) error
	GetRequestDeploymentByID(id int) (model.RequestDeployment, error)
	ListRequestDeployments(startDate, endDate string, id, pageNumber, pageSize int) ([]model.RequestDeployment, error)
	CountRequestDeployments(startDate, endDate string, id int) (int64, error)
	CheckIfRequestHasEnded(id int) (bool,error)
	HasErrorInRequest(id int) (bool,error)
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
func (dao RequestDeploymentDAOImpl) CheckIfRequestHasEnded(id int) (bool,error) {
	var deployment model.Deployment
	var count = -1
	err := dao.Db.Where(
		"request_deployment_id = ? AND processed = ?",
		id,
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
func (dao RequestDeploymentDAOImpl) HasErrorInRequest(id int) (bool,error) {
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
func (dao RequestDeploymentDAOImpl) ListRequestDeployments(startDate, endDate string, id, pageNumber, pageSize int) ([]model.RequestDeployment, error) {
	var rdList []model.RequestDeployment
	sql := prepareWhere(id)
	rows, err := dao.Db.Table("request_deployments").Select(
		"id, created_at, updated_at, processed ,success",
	).Where(sql, startDate, endDate).Offset((pageNumber - 1) * pageSize).Limit(pageSize).Rows()

	for rows.Next() {
		id := 0
		createdAt := time.Time{}
		updatedAt := time.Time{}
		success, processed := false, false
		rows.Scan(&id, &createdAt, &updatedAt, &processed, &success)

		request := model.RequestDeployment{}
		request.ID = uint(id)
		request.CreatedAt = createdAt
		request.UpdatedAt = updatedAt
		request.Processed = processed
		request.Success = success
		
		rdList = append(rdList, request)
	}

	return rdList, err
}

//CountRequestDeployments count
func (dao RequestDeploymentDAOImpl) CountRequestDeployments(startDate, endDate string, id int) (int64, error) {
	var deployment model.RequestDeployment
	var count int64
	sql := prepareWhere(id)
	err := dao.Db.Where(sql, startDate, endDate).Model(&deployment).Count(&count).Error
	return count, err
}

func prepareWhere(id int) string {
	where := "date(created_at) >= ? AND date(created_at) <= ?"
	if id != -1 {
		where = where + "AND id = " + string(id)
	}
	return where
}