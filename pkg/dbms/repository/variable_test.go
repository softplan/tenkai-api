package repository

import (
	"errors"
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

func TestCreateVariable(t *testing.T) {

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
	v.ID = 999

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)

	mock.ExpectQuery(`SELECT (.*) FROM "variables" 
		WHERE "variables"."deleted_at" IS NULL AND \(\("variables"."scope" = (.*)\) 
		AND \("variables"."name" = (.*)\) AND \("variables"."environment_id" = (.*)\)\) 
		ORDER BY "variables"."id" ASC LIMIT 1`).
		WithArgs(v.Scope, v.Name, v.EnvironmentID).
		WillReturnError(errors.New("mock error"))

	mock.ExpectQuery(`INSERT INTO "variables"`).
		WithArgs(999, AnyTime{}, AnyTime{}, nil, v.Scope, v.Name, v.Value, v.Secret, v.Description, v.EnvironmentID).
		WillReturnRows(rows)

	audit, updated, err := dao.CreateVariable(v)
	assert.Nil(t, err)
	assert.NotNil(t, audit)
	assert.True(t, updated)

	mock.ExpectationsWereMet()
}

func TestCreateVariable_Audit(t *testing.T) {

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
	v.ID = 999

	rows1 := sqlmock.NewRows([]string{"id", "scope", "name", "value", "description", "environment_id", "secret"}).
		AddRow(v.ID, v.Scope, v.Name, "new value", v.Description, v.EnvironmentID, v.Secret)

	mock.ExpectQuery(`SELECT (.*) FROM "variables" 
	WHERE "variables"."deleted_at" IS NULL AND \(\("variables"."scope" = (.*)\) 
	AND \("variables"."name" = (.*)\) AND \("variables"."environment_id" = (.*)\)\) 
	ORDER BY "variables"."id" ASC LIMIT 1`).
		WithArgs(v.Scope, v.Name, v.EnvironmentID).
		WillReturnRows(rows1)

	mock.ExpectExec(`UPDATE "variables" SET (.*) WHERE (.*)`).
		WithArgs(AnyTime{}, nil, v.Scope, v.Name, v.Value, v.Secret, v.Description, v.EnvironmentID, v.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	audit, updated, err := dao.CreateVariable(v)
	assert.Nil(t, err)
	assert.NotNil(t, audit)
	assert.True(t, updated)

	mock.ExpectationsWereMet()
}
