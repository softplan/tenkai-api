package repository

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/stretchr/testify/assert"
)

func compareBeforeTest(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, CompareEnvsQueryDAOImpl, model.CompareEnvsQuery) {
	db, mock, err := sqlmock.New()
	gormDB, err := gorm.Open("postgres", db)

	assert.Nil(t, err)

	dao := CompareEnvsQueryDAOImpl{}
	dao.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	query := json.RawMessage(`{"foo":"bar"}`)

	var item model.CompareEnvsQuery
	item.Name = "My query"
	item.UserID = 9999
	item.Query = postgres.Jsonb{RawMessage: query}

	return gormDB, mock, dao, item
}

func TestCreateCompareEnvsQuery(t *testing.T) {
	gormDB, mock, dao, item := compareBeforeTest(t)
	defer gormDB.Close()

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)

	mock.ExpectQuery(`INSERT INTO "compare_envs_queries"`).
		WithArgs(AnyTime{}, AnyTime{}, nil, item.Name, item.UserID, item.Query.RawMessage).
		WillReturnRows(rows)

	result, e := dao.SaveCompareEnvsQuery(item)
	assert.Nil(t, e)
	assert.Equal(t, 1, result)

	mock.ExpectationsWereMet()
}

func TestCreateCompareEnvsQuery_Error(t *testing.T) {
	gormDB, mock, dao, item := compareBeforeTest(t)
	defer gormDB.Close()

	mock.ExpectQuery(`INSERT INTO "compare_envs_queries"`).
		WithArgs(AnyTime{}, AnyTime{}, nil, item.Name, item.UserID, item.Query.RawMessage).
		WillReturnError(errors.New("some error"))

	_, e := dao.SaveCompareEnvsQuery(item)
	assert.Error(t, e)

	mock.ExpectationsWereMet()
}

func TestGetByUser(t *testing.T) {

	gormDB, mock, dao, item := compareBeforeTest(t)
	defer gormDB.Close()

	item.ID = 888
	rows := sqlmock.NewRows([]string{"id", "name", "user_id", "query"}).
		AddRow(item.ID, item.Name, item.UserID, item.Query)

	item.ID = 888
	mock.ExpectQuery(`SELECT (.+) FROM "compare_envs_queries" WHERE (.+)`).
		WithArgs(888).WillReturnRows(rows)

	list, err := dao.GetByUser(888)
	assert.Nil(t, err)
	assert.NotNil(t, list)

	mock.ExpectationsWereMet()
}

func TestGetByUser_NotFound(t *testing.T) {

	gormDB, mock, dao, item := compareBeforeTest(t)
	defer gormDB.Close()

	item.ID = 888

	item.ID = 888
	mock.ExpectQuery(`SELECT (.+) FROM "compare_envs_queries" WHERE (.+)`).
		WithArgs(888).WillReturnError(gorm.ErrRecordNotFound)

	result, err := dao.GetByUser(888)
	assert.Nil(t, err)
	assert.Empty(t, result)

	mock.ExpectationsWereMet()
}

func TestGetByUser_Error(t *testing.T) {

	gormDB, mock, dao, item := compareBeforeTest(t)
	defer gormDB.Close()

	item.ID = 888

	item.ID = 888
	mock.ExpectQuery(`SELECT (.+) FROM "compare_envs_queries" WHERE (.+)`).
		WithArgs(888).WillReturnError(errors.New("mock error"))

	_, err := dao.GetByUser(888)
	assert.Error(t, err)

	mock.ExpectationsWereMet()
}

func TestDeleteCompareEnvQuery(t *testing.T) {

	gormDB, mock, dao, item := compareBeforeTest(t)
	defer gormDB.Close()

	item.ID = 999
	mock.ExpectExec(`DELETE FROM "compare_envs_queries" WHERE (.*)`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := dao.DeleteCompareEnvQuery(999)
	assert.Nil(t, err)

	mock.ExpectationsWereMet()
}
