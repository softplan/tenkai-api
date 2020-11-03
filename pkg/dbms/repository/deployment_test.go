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
	deployment.Environment = 1
	deployment.User = 1
	
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
			deployment.Environment,
			deployment.Chart,
			deployment.User,
			deployment.Success,
			deployment.Message,
		).WillReturnRows(rows)
	
		_, err = deploymentDAO.CreateDeployment(deployment)
		

	assert.Nil(test, err)
	mock.ExpectationsWereMet()
}


