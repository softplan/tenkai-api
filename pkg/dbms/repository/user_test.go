package repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/stretchr/testify/assert"
)

func getUser() model.User {
	// var envs []model.Environment
	// e := getEnvironmentTestData()
	// e.ID = 999
	// envs = append(envs, e)

	item := model.User{}
	item.Email = "musk@mars.com"
	item.DefaultEnvironmentID = 999
	return item
}

func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	userDAO := UserDAOImpl{}
	userDAO.Db = gormDB

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)

	payload := getUser()

	mock.ExpectQuery(`INSERT INTO "users"`).
		WithArgs(AnyTime{}, AnyTime{}, nil, payload.Email, payload.DefaultEnvironmentID).
		WillReturnRows(rows)

	e := userDAO.CreateUser(payload)
	assert.NoError(t, e)

	mock.ExpectationsWereMet()
}

func TestDeleteUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	userDAO := UserDAOImpl{}
	userDAO.Db = gormDB

	payload := getUser()
	payload.ID = 999

	row1 := sqlmock.NewRows([]string{"id"}).AddRow(payload.ID)
	mock.ExpectQuery(`SELECT (.*) FROM "users"
		WHERE "users"."deleted_at" IS NULL AND \(\("users"."id" = 999\)\)
		ORDER BY "users"."id" ASC LIMIT 1
	`).WillReturnRows(row1)

	mock.ExpectExec(`DELETE FROM "user_environment" WHERE \("user_id" IN (.*)\)`).
		WithArgs(payload.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(`DELETE FROM "users" WHERE "users"."id" = (.*)`).
		WithArgs(payload.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	e := userDAO.DeleteUser(int(payload.ID))
	assert.NoError(t, e)

	mock.ExpectationsWereMet()
}
