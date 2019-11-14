package repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/stretchr/testify/assert"
)

func getUser() model.User {
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

func TestAssociateEnvironmentUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	userDAO := UserDAOImpl{}
	userDAO.Db = gormDB

	payload := getUser()
	payload.ID = 888

	env := getEnvironmentTestData()
	env.ID = 999

	payload.Environments = append(payload.Environments, env)

	row1 := sqlmock.NewRows([]string{"id", "email", "default_environment_id"}).AddRow(payload.ID, payload.Email, payload.DefaultEnvironmentID)
	mock.ExpectQuery(`SELECT (.*) FROM "users"
		WHERE "users"."deleted_at" IS NULL AND \(\("users"."id" = 888\)\)
		ORDER BY "users"."id" ASC LIMIT 1
	`).WillReturnRows(row1)

	row2 := sqlmock.NewRows([]string{"id", "group", "name"}).AddRow(999, "foo", "bar")
	mock.ExpectQuery(`SELECT (.*) FROM "environments"
		WHERE "environments"."deleted_at" IS NULL AND \(\("environments"."id" = 999\)\)
		ORDER BY "environments"."id" ASC LIMIT 1
	`).WillReturnRows(row2)

	mock.ExpectExec(`INSERT INTO "user_environment" \("user_id","environment_id"\) SELECT (.*)  WHERE NOT EXISTS 
			\(SELECT (.*) FROM "user_environment" WHERE "user_id" = (.*) AND "environment_id" = (.*)\)`).
		WithArgs(888, 999, 888, 999).
		WillReturnResult(sqlmock.NewResult(1, 1))

	e := userDAO.AssociateEnvironmentUser(888, 999)
	assert.NoError(t, e)

	mock.ExpectationsWereMet()
}

func TestListAllUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	userDAO := UserDAOImpl{}
	userDAO.Db = gormDB

	payload := getUser()
	payload.ID = 888

	env := getEnvironmentTestData()
	env.ID = 999

	payload.Environments = append(payload.Environments, env)

	row1 := sqlmock.NewRows([]string{"id", "email", "default_environment_id"}).AddRow(payload.ID, payload.Email, payload.DefaultEnvironmentID)
	mock.ExpectQuery(`SELECT (.*) FROM "users"
		WHERE "users"."deleted_at" IS NULL
	`).WillReturnRows(row1)

	row2 := sqlmock.NewRows([]string{"id"}).AddRow(env.ID)
	mock.ExpectQuery(`SELECT (.*) FROM "environments" 
		INNER JOIN "user_environment" ON "user_environment"."environment_id" = "environments"."id" 
		WHERE "environments"."deleted_at" IS NULL AND \(\("user_environment"."user_id" IN (.*)\)\)
	`).WithArgs(888).WillReturnRows(row2)

	u, e := userDAO.ListAllUsers()
	assert.NoError(t, e)
	assert.NotNil(t, u)

	mock.ExpectationsWereMet()
}
