package repository

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func getReleaseTestData() model.Release {
	now := time.Now()
	item := model.Release{}
	item.CreatedAt = now
	item.DeletedAt = nil
	item.UpdatedAt = now
	item.ChartName = "saj6/alfa"
	item.Release = "my_release"
	return item
}

func TestCreateRelease(t *testing.T) {

	db, mock, err := sqlmock.New()

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	assert.Nil(t, err)

	dockerDAO := ReleaseDAOImpl{}
	dockerDAO.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	rows := sqlmock.NewRows([]string{"ID"}).AddRow(1)

	item := getReleaseTestData()

	mock.ExpectQuery(`INSERT INTO "releases"`).
		WithArgs(item.CreatedAt, item.UpdatedAt, item.DeletedAt, item.ChartName, item.Release).
		WillReturnRows(rows)

	err = dockerDAO.CreateRelease(item)
	assert.Nil(t, err)

	mock.ExpectationsWereMet()

}
