package repository

import (
	"testing"

	"encoding/json"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/stretchr/testify/assert"
)

func TestCreateCompareEnvsQuery(t *testing.T) {

	db, mock, err := sqlmock.New()

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	assert.Nil(t, err)

	envDAO := CompareEnvsQueryDAOImpl{}
	envDAO.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)

	query := json.RawMessage(`{"foo":"bar"}`)

	var item model.CompareEnvsQuery
	item.Name = "My query"
	item.UserID = 9999
	item.Query = postgres.Jsonb{RawMessage: query}

	mock.ExpectQuery(`INSERT INTO "compare_envs_queries"`).
		WithArgs(item.CreatedAt, item.UpdatedAt, item.DeletedAt, item.Name, item.UserID, item.Query.RawMessage).
		WillReturnRows(rows)

	result, e := envDAO.CreateCompareEnvsQuery(item)
	assert.Nil(t, e)
	assert.Equal(t, 1, result)

	mock.ExpectationsWereMet()
}
