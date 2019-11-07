package repository

import (
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/stretchr/testify/assert"

	"github.com/DATA-DOG/go-sqlmock"
)

func getTestData() model.DockerRepo {

	now := time.Now()
	item := model.DockerRepo{}
	item.Password = "my_password"
	item.Username = "my_username"
	item.Host = "my_host"
	item.CreatedAt = now
	item.DeletedAt = nil
	item.UpdatedAt = now
	return item

}

func TestCreateDockerRepo(t *testing.T) {

	db, mock, err := sqlmock.New()

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	assert.Nil(t, err)

	dockerDAO := DockerDAOImpl{}
	dockerDAO.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)

	item := getTestData()

	mock.ExpectQuery(`INSERT INTO "docker_repos"`).
		WithArgs(item.CreatedAt, item.UpdatedAt, item.DeletedAt, item.Host, item.Username, item.Password).
		WillReturnRows(rows)

	result, err := dockerDAO.CreateDockerRepo(item)
	assert.Nil(t, err)
	assert.Equal(t, 1, result)

	mock.ExpectationsWereMet()

}

func TestGetDockerRepositoryByHost(t *testing.T) {

	db, mock, err := sqlmock.New()

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	assert.Nil(t, err)

	dockerDAO := DockerDAOImpl{}
	dockerDAO.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	item := getTestData()

	rows := sqlmock.NewRows([]string{"id", "host"}).AddRow(1, item.Host)

	mock.ExpectQuery(`SELECT (.+) FROM "docker_repos"`).
		WithArgs(item.Host).
		WillReturnRows(rows)

	result, err := dockerDAO.GetDockerRepositoryByHost(item.Host)
	assert.Nil(t, err)
	assert.Equal(t, uint(1), result.ID)

	mock.ExpectationsWereMet()

}
