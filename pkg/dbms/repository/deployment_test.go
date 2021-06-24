package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	model2 "github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/stretchr/testify/assert"
)

func getDeployment() model2.Deployment {
	now := time.Now()
	deployment := model2.Deployment{}
	deployment.CreatedAt = now
	deployment.UpdatedAt = now
	deployment.DeletedAt = nil
	deployment.Chart = "Chart Teste"
	deployment.Success = true
	deployment.Processed = true
	deployment.Message = "Message teste"
	deployment.EnvironmentID = 1
	deployment.RequestDeploymentID = 1
	deployment.ChartVersion = "0.1.0"
	deployment.DockerVersion = "master"
	return deployment
}

func TestCreateDeployment(test *testing.T) {
	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(test, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	deploymentDAO := DeploymentDAOImpl{}
	deploymentDAO.Db = gormDB

	deployment := getDeployment()

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)

	mock.ExpectQuery(
		`INSERT INTO "deployments" .*`,
	).WithArgs(
		AnyTime{},
		AnyTime{},
		nil,
		deployment.RequestDeploymentID,
		deployment.EnvironmentID,
		deployment.Chart,
		deployment.ChartVersion,
		deployment.Processed,
		deployment.Success,
		deployment.Message,
		deployment.DockerVersion,
	).WillReturnRows(rows)

	_, err = deploymentDAO.CreateDeployment(deployment)

	assert.Nil(test, err)
	mock.ExpectationsWereMet()
}

func TestGetDeploymentByID(test *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(test, err)
	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	mock.MatchExpectationsInOrder(false)

	deploymentDAO := DeploymentDAOImpl{
		Db: gormDB,
	}

	rows := sqlmock.NewRows(
		[]string{
			"id", "created_at", "updated_at", "deleted_at",
			"request_deployment_id", "environment_id", "chart",
			"processed", "success", "message",
		}).AddRow(999, time.Now(), time.Now(), nil, 17, 17,
		"Chart Teste", true, true, "",
	)

	mock.ExpectQuery(`SELECT (.*) FROM "deployments"`).WillReturnRows(rows)

	result, err := deploymentDAO.GetDeploymentByID(999)
	assert.Nil(test, err)
	assert.NotNil(test, result)

	mock.ExpectationsWereMet()
}

func TestEditDeployment(test *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(test, err)
	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()
	deploymentDAO := DeploymentDAOImpl{Db: gormDB}

	deployment := getDeployment()
	deployment.ID = 999

	mock.ExpectExec(
		`UPDATE "deployments" SET (.*) WHERE (.*)`,
	).WithArgs(
		AnyTime{},
		AnyTime{},
		nil,
		deployment.RequestDeploymentID,
		deployment.EnvironmentID,
		deployment.Chart,
		deployment.ChartVersion,
		deployment.Processed,
		deployment.Success,
		deployment.Message,
		deployment.DockerVersion,
		deployment.ID,
	).WillReturnResult(
		sqlmock.NewResult(1, 1),
	)

	err = deploymentDAO.EditDeployment(deployment)

	assert.Nil(test, err)

	mock.ExpectationsWereMet()
}

func TestGetCountDeployments(test *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(test, err)
	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()
	mock.MatchExpectationsInOrder(false)
	deploymentDAO := DeploymentDAOImpl{
		Db: gormDB,
	}
	rows := sqlmock.NewRows([]string{"1,1"}).AddRow(1)
	mock.ExpectQuery(`SELECT count\(\*\) FROM "deployments" WHERE .*`).WillReturnRows(rows)
	result, err := deploymentDAO.CountDeployments("1", "1")
	assert.Nil(test, err, "Error on get count of deployments")
	assert.NotNil(test, result, "Result of count is nil")
}

func TestListDeployments(test *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(test, err)
	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()
	mock.MatchExpectationsInOrder(false)
	deploymentDAO := DeploymentDAOImpl{
		Db: gormDB,
	}
	rows := sqlmock.NewRows([]string{
		"id",
		"created_at",
		"updated_at",
		"chart",
		"request_deployment_id",
		"environments_id",
		"environments_name",
		"processed",
		"success",
		"message",
	}).AddRow(1, time.Time{}, time.Time{}, "", 1, 1, "", true, true, "")

	mock.ExpectQuery(`SELECT .* FROM .*"`).WillReturnRows(rows)
	_, err = deploymentDAO.ListDeployments("1", "1", 1, 100)

	assert.Nil(test, err, "Error on get count of deployments")
}
