package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/stretchr/testify/assert"
)

func getEnvironmentTestData() model.Environment {
	now := time.Now()
	item := model.Environment{}
	item.CreatedAt = now
	item.DeletedAt = nil
	item.UpdatedAt = now
	item.Group = "my-group"
	item.Name = "env-name"
	item.ClusterURI = "qwe"
	item.CACertificate = "asd"
	item.Token = "zxc"
	item.Namespace = "dev"
	item.Gateway = "dsa"
	return item
}

func TestCreateEnvironment(t *testing.T) {

	db, mock, err := sqlmock.New()

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	assert.Nil(t, err)

	envDAO := EnvironmentDAOImpl{}
	envDAO.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	rows := sqlmock.NewRows([]string{"ID"}).AddRow(1)

	item := getEnvironmentTestData()

	mock.ExpectQuery(`INSERT INTO "environments"`).
		WithArgs(item.CreatedAt, item.DeletedAt, item.UpdatedAt, item.Group,
			item.Name, item.Name, item.ClusterURI, item.CACertificate, item.Token,
			item.Namespace, item.Gateway).
		WillReturnRows(rows)

	_, e := envDAO.CreateEnvironment(item)
	assert.Nil(t, e)

	mock.ExpectationsWereMet()

}
