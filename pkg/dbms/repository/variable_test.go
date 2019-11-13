package repository

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEditVariable(t *testing.T) {

	db, mock, err := sqlmock.New()

	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	dao := VariableDAOImpl{}
	dao.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	v := model.Variable{}
	v.EnvironmentID = 10
	v.Description = "Description value"
	v.Secret = false
	v.Value = "value value"
	v.Name = "name value"
	v.Scope = "serviceA"
	v.ID = 1

	mock.ExpectExec(`UPDATE "variables" SET (.*) WHERE (.*)`).
		WithArgs(AnyTime{}, nil, v.Scope, v.Name, v.Value, v.Secret, v.Description, v.EnvironmentID, v.ID).WillReturnResult(sqlmock.NewResult(1, 1))

	err = dao.EditVariable(v)

	assert.Nil(t, err)

	mock.ExpectationsWereMet()

}
