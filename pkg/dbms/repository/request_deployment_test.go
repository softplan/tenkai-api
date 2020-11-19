package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	model2 "github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/stretchr/testify/assert"
)

func getRequestDeployment() model2.RequestDeployment {
	now := time.Now()
	deployment := model2.RequestDeployment{}
	deployment.CreatedAt = now
	deployment.UpdatedAt = now
	deployment.DeletedAt = nil
	deployment.Success = true
	deployment.Processed = true
	return deployment
}

func TestCreateRequestDeployment(test *testing.T) {
	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(test, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	requestDeploymentDAO := RequestDeploymentDAOImpl{}
	requestDeploymentDAO.Db = gormDB

	requestDeployment := getRequestDeployment()

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)

	mock.ExpectQuery(
		`INSERT INTO "request_deployments" .*`,
	).WithArgs(
		AnyTime{},
		AnyTime{},
		nil,
		requestDeployment.Success,
		requestDeployment.Processed,
	).WillReturnRows(rows)

	_, err = requestDeploymentDAO.CreateRequestDeployment(requestDeployment)

	assert.Nil(test, err)
	mock.ExpectationsWereMet()
}

func TestGetRequestDeploymentByID(test *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(test, err)
	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	mock.MatchExpectationsInOrder(false)

	requestDeploymentDAO := RequestDeploymentDAOImpl{
		Db: gormDB,
	}

	rows := sqlmock.NewRows(
		[]string{
			"id", "created_at", "updated_at", "deleted_at",
			"success", "processed",
		}).AddRow(999, time.Now(), time.Now(), nil, true, true,
	)

	mock.ExpectQuery(`SELECT (.*) FROM "request_deployments"`).WillReturnRows(rows)

	result, err := requestDeploymentDAO.GetRequestDeploymentByID(999)
	assert.Nil(test, err)
	assert.NotNil(test, result)

	mock.ExpectationsWereMet()
}

func TestEditRequestDeployment(test *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(test, err)
	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()
	deploymentDAO := RequestDeploymentDAOImpl{Db: gormDB}

	deployment := getRequestDeployment()
	deployment.ID = 999

	mock.ExpectExec(
		`UPDATE "request_deployments" SET (.*) WHERE (.*)`,
	).WithArgs(
		AnyTime{},
		AnyTime{},
		nil,
		deployment.Processed,
		deployment.Success,
		deployment.ID,
	).WillReturnResult(
		sqlmock.NewResult(1, 1),
	)

	err = deploymentDAO.EditRequestDeployment(deployment)

	assert.Nil(test, err)

	mock.ExpectationsWereMet()
}

func TestCheckIfRequestHasEnded(test *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(test, err)
	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()
	deploymentDAO := RequestDeploymentDAOImpl{Db: gormDB}
	rows := sqlmock.NewRows([]string{"1,1"}).AddRow(1)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "request_deployments" WHERE .*`).WithArgs("1", false).WillReturnRows(rows)
	result, err := deploymentDAO.CheckIfRequestHasEnded(1)
	assert.Nil(test, err, "Error on get count of deployments")
	assert.NotNil(test, result, "Result of count is nil")

}

func TestHasErrorInRequest(test *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(test, err)
	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()
	deploymentDAO := RequestDeploymentDAOImpl{Db: gormDB}
	rows := sqlmock.NewRows([]string{"1,1"}).AddRow(1)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "request_deployments" WHERE .*`).WithArgs("1", false).WillReturnRows(rows)
	deploymentDAO.HasErrorInRequest(1)
}

func TestListRequestDeployments(test *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(test, err)
	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()
	deploymentDAO := RequestDeploymentDAOImpl{Db: gormDB}

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "updated_at", "processed", "success",
	}).AddRow(1, time.Time{}, time.Time{}, true, true)
	
	mock.ExpectQuery(
		`SELECT .* FROM "request_deployments" .* WHERE .*`,
		).WithArgs(
			"2020-01-01","2020-01-01",
		).WillReturnRows(rows)
	_, err = deploymentDAO.ListRequestDeployments("2020-01-01","2020-01-01","",1,1,100)
	assert.Nil(test,err,"List has error")
}

func TestGetCountRequestDeployments(test *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(test, err)
	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()
	mock.MatchExpectationsInOrder(false)
	deploymentDAO := RequestDeploymentDAOImpl{
		Db: gormDB,
	}
	rows := sqlmock.NewRows([]string{"1,1"}).AddRow(1)
	mock.ExpectQuery(`SELECT .* FROM "request_deployments" .*`).WillReturnRows(rows)
	result, err := deploymentDAO.CountRequestDeployments("2020-01-01","2020-01-01","1")
	assert.Nil(test, err, "Error on get count of deployments")
	assert.NotNil(test, result, "Result of count is nil")
}