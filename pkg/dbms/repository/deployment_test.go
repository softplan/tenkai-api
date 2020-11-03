package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	model2 "github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/stretchr/testify/assert"
)

func getDeployment() (model2.Deployment){
	now := time.Now()
	deployment := model2.Deployment{}
	deployment.CreatedAt = now
	deployment.UpdatedAt = now
	deployment.DeletedAt = nil
	deployment.Chart = "Chart Teste"
	deployment.Success = true
	deployment.Message = "Message teste"
	deployment.EnvironmentID = 1
	deployment.UserID = 1
	
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
		`INSERT INTO "deployments"`,
		).WithArgs(
			AnyTime{}, 
			AnyTime{}, 
			nil,
			deployment.EnvironmentID,
			deployment.Chart,
			deployment.UserID,
			deployment.Success,
			deployment.Message,
		).WillReturnRows(rows)
	
	_, err = deploymentDAO.CreateDeployment(deployment)
		
	assert.Nil(test, err)
	mock.ExpectationsWereMet()
}

func TestGetDeploymentByID(test *testing.T){
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
			"environment_id", "chart", "user_id", "success",
			"message",
		}).AddRow(999, time.Now(), time.Now(), nil, 17, 
			"Chart Teste", 1, true, "",
		)
	
	mock.ExpectQuery(`SELECT (.*) FROM "deployments"`).WillReturnRows(rows)

	result, err := deploymentDAO.GetDeploymentByID(999)
	assert.Nil(test, err)
	assert.NotNil(test, result)

	mock.ExpectationsWereMet()
}

func TestEditDeployment(test *testing.T){
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
		deployment.EnvironmentID,
		deployment.Chart,
		deployment.UserID,
		deployment.Success,
		deployment.Message,
		deployment.ID,
	).WillReturnResult(
		sqlmock.NewResult(1,1),
	)
	
	err = deploymentDAO.EditDeployment(deployment)
	
	assert.Nil(test, err)

	mock.ExpectationsWereMet()
}