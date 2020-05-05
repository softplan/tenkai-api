package repository

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/stretchr/testify/assert"
)

func TestEditVariable(t *testing.T) {

	db, mock, err := sqlmock.New()

	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	dao := VariableDAOImpl{}
	dao.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	v := getVariable()

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

	v := getVariable()

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)

	mock.ExpectQuery(`SELECT (.*) FROM "variables" WHERE (.*) ORDER BY (.*) ASC LIMIT 1`).
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

func TestCreateVariable_Error(t *testing.T) {

	db, mock, err := sqlmock.New()

	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	dao := VariableDAOImpl{}
	dao.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	v := getVariable()

	mock.ExpectQuery(`SELECT (.*) FROM "variables" WHERE (.*) ORDER BY (.*) ASC LIMIT 1`).
		WithArgs(v.Scope, v.Name, v.EnvironmentID).
		WillReturnError(errors.New("mock error"))

	mock.ExpectQuery(`INSERT INTO "variables"`).
		WithArgs(999, AnyTime{}, AnyTime{}, nil, v.Scope, v.Name, v.Value, v.Secret, v.Description, v.EnvironmentID).
		WillReturnError(errors.New("mock error"))

	audit, updated, err := dao.CreateVariable(v)
	assert.Error(t, err)
	assert.NotNil(t, audit)
	assert.False(t, updated)

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

	v := getVariable()

	rows1 := sqlmock.NewRows([]string{"id", "scope", "name", "value", "description", "environment_id", "secret"}).
		AddRow(v.ID, v.Scope, v.Name, "new value", v.Description, v.EnvironmentID, v.Secret)

	mock.ExpectQuery(`SELECT (.*) FROM "variables" WHERE (.*) ORDER BY (.*) ASC LIMIT 1`).
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

func TestCreateVariable_AuditSaveError(t *testing.T) {

	db, mock, err := sqlmock.New()

	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	dao := VariableDAOImpl{}
	dao.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	v := getVariable()

	rows1 := sqlmock.NewRows([]string{"id", "scope", "name", "value", "description", "environment_id", "secret"}).
		AddRow(v.ID, v.Scope, v.Name, "new value", v.Description, v.EnvironmentID, v.Secret)

	mock.ExpectQuery(`SELECT (.*) FROM "variables" WHERE (.*) ORDER BY (.*) ASC LIMIT 1`).
		WithArgs(v.Scope, v.Name, v.EnvironmentID).
		WillReturnRows(rows1)

	mock.ExpectExec(`UPDATE "variables" SET (.*) WHERE (.*)`).
		WithArgs(AnyTime{}, nil, v.Scope, v.Name, v.Value, v.Secret, v.Description, v.EnvironmentID, v.ID).
		WillReturnError(errors.New("mock error"))

	audit, updated, err := dao.CreateVariable(v)
	assert.Error(t, err)
	assert.NotNil(t, audit)
	assert.False(t, updated)

	mock.ExpectationsWereMet()
}

func TestCreateVariableWithDefaultValue(t *testing.T) {

	db, mock, err := sqlmock.New()

	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	dao := VariableDAOImpl{}
	dao.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	v := getVariable()

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)

	mock.ExpectQuery(`SELECT (.*) FROM "variables" WHERE (.*) ORDER BY (.*) ASC LIMIT 1`).
		WithArgs(v.Scope, v.Name, v.EnvironmentID).
		WillReturnError(gorm.ErrRecordNotFound)

	mock.ExpectQuery(`INSERT INTO "variables"`).
		WithArgs(999, AnyTime{}, AnyTime{}, nil, v.Scope, v.Name, v.Value, v.Secret, v.Description, v.EnvironmentID).
		WillReturnRows(rows)

	audit, updated, err := dao.CreateVariableWithDefaultValue(v)
	assert.Nil(t, err)
	assert.NotNil(t, audit)
	assert.True(t, updated)

	mock.ExpectationsWereMet()
}

func TestGetAllVariablesByEnvironment(t *testing.T) {

	db, mock, err := sqlmock.New()

	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	dao := VariableDAOImpl{}
	dao.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	rows1 := sqlmock.NewRows([]string{"id", "group", "name"}).
		AddRow(10, "my-group", "env-name")

	mock.ExpectQuery(`SELECT (.*) FROM "environments" WHERE (.*) ORDER BY (.*) ASC LIMIT 1`).
		WillReturnRows(rows1)

	v := getVariable()

	rows2 := sqlmock.NewRows([]string{"id", "scope", "name", "value", "description", "environment_id", "secret"}).
		AddRow(v.ID, v.Scope, v.Name, "new value", v.Description, v.EnvironmentID, v.Secret)

	mock.ExpectQuery(`SELECT (.*) FROM "variables" WHERE (.*) ORDER BY (.*)`).
		WithArgs(10).
		WillReturnRows(rows2)

	result, err := dao.GetAllVariablesByEnvironment(10)
	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestGetAllVariablesByEnvironment_Error1(t *testing.T) {

	db, mock, err := sqlmock.New()

	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	dao := VariableDAOImpl{}
	dao.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	mock.ExpectQuery(`SELECT (.*) FROM "environments" WHERE (.*) ORDER BY (.*) ASC LIMIT 1`).
		WillReturnError(errors.New("mock error"))

	_, err = dao.GetAllVariablesByEnvironment(10)
	assert.Error(t, err)
}

func TestGetAllVariablesByEnvironment_Error2(t *testing.T) {

	db, mock, err := sqlmock.New()

	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	dao := VariableDAOImpl{}
	dao.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	rows1 := sqlmock.NewRows([]string{"id", "group", "name"}).
		AddRow(10, "my-group", "env-name")

	mock.ExpectQuery(`SELECT (.*) FROM "environments" WHERE (.*) ORDER BY (.*) ASC LIMIT 1`).
		WillReturnRows(rows1)

	mock.ExpectQuery(`SELECT (.*) FROM "variables" WHERE (.*) ORDER BY (.*)`).
		WithArgs(10).
		WillReturnError(errors.New("mock error"))

	_, err = dao.GetAllVariablesByEnvironment(10)
	assert.Error(t, err)
}

func getVariable() model.Variable {
	v := model.Variable{}
	v.EnvironmentID = 10
	v.Description = "Description value"
	v.Secret = false
	v.Value = "value value"
	v.Name = "name value"
	v.Scope = "serviceA"
	v.ID = 999

	return v
}

func TestGetAllVariablesByEnvironmentAndScopeWithContext(t *testing.T) {

	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	dao := VariableDAOImpl{}
	dao.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	rows1 := sqlmock.NewRows([]string{"id", "scope", "name", "value", "description", "environment_id", "secret"}).
		AddRow(10, "xpto/alfa", "env-name", "env-value", "description_value", 999, false)

	mock.ExpectQuery(`SELECT (.*) FROM "variables" WHERE (.*)`).WithArgs("xpto/alfa", 999).WillReturnRows(rows1)

	variables, err := dao.GetAllVariablesByEnvironmentAndScope(999, "xpto/alfa")
	assert.Nil(t, err)
	assert.NotNil(t, variables)

	mock.ExpectationsWereMet()

}

func TestGetAllVariablesByEnvironmentAndScopeWithContext_Error(t *testing.T) {

	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	dao := VariableDAOImpl{}
	dao.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	mock.ExpectQuery(`SELECT (.*) FROM "variables" WHERE (.*)`).WithArgs("xpto/alfa", 999).
		WillReturnError(errors.New("mock error"))

	variables, err := dao.GetAllVariablesByEnvironmentAndScope(999, "xpto/alfa")
	assert.Error(t, err)
	assert.Nil(t, variables)

	mock.ExpectationsWereMet()

}

func TestGetAllVariablesByEnvironmentAndScopeWithoutContext(t *testing.T) {

	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	dao := VariableDAOImpl{}
	dao.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	rows1 := sqlmock.NewRows([]string{"id", "scope", "name", "value", "description", "environment_id", "secret"}).
		AddRow(10, "xpto/alfa", "env-name", "env-value", "description_value", 999, false)

	mock.ExpectQuery(`SELECT (.*) FROM "variables" WHERE "variables"."deleted_at" IS NULL AND \(\(environment_id = (.*) AND scope LIKE (.*)\)\)`).WithArgs(999, "%alfa").WillReturnRows(rows1)

	variables, err := dao.GetAllVariablesByEnvironmentAndScope(999, "alfa")
	assert.Nil(t, err)
	assert.NotNil(t, variables)

	mock.ExpectationsWereMet()

}

func TestGetAllVariablesByEnvironmentAndScopeWithoutContext_Error(t *testing.T) {

	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	dao := VariableDAOImpl{}
	dao.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	mock.ExpectQuery(`SELECT (.*) FROM "variables" WHERE "variables"."deleted_at" IS NULL AND \(\(environment_id = (.*) AND scope LIKE (.*)\)\)`).WithArgs(999, "%alfa").WillReturnError(errors.New("mock error"))

	variables, err := dao.GetAllVariablesByEnvironmentAndScope(999, "alfa")
	assert.Error(t, err)
	assert.Nil(t, variables)

	mock.ExpectationsWereMet()

}

func TestGetVarImageTagByEnvAndScope(t *testing.T) {

	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	dao := VariableDAOImpl{}
	dao.Db = gormDB

	mock.MatchExpectationsInOrder(false)
	rows1 := sqlmock.NewRows([]string{"id", "scope", "name", "value", "description", "environment_id", "secret"}).
		AddRow(888, "repo/my-chart", "env-name", "env-value", "description_value", 999, false)

	mock.ExpectQuery(`SELECT (.*) FROM "variables" WHERE (.*)`).WithArgs(999, "repo/my-chart").
		WillReturnRows(rows1)

	variable, err := dao.GetVarImageTagByEnvAndScope(999, "repo/my-chart")
	assert.Nil(t, err)
	assert.NotNil(t, variable)

	mock.ExpectationsWereMet()

}

func TestDeleteVariable(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	dao := VariableDAOImpl{}
	dao.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	mock.ExpectExec(`DELETE FROM "variables" WHERE (.*)`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = dao.DeleteVariable(999)
	assert.Nil(t, err)

	mock.ExpectationsWereMet()
}

func TestDeleteVariableByEnvironmentID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	dao := VariableDAOImpl{}
	dao.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	mock.ExpectExec(`DELETE FROM "variables" WHERE (.*)`).
		WithArgs(999).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = dao.DeleteVariableByEnvironmentID(999)
	assert.Nil(t, err)

	mock.ExpectationsWereMet()
}

func TestVariableGetByID(t *testing.T) {
	db, mock, err := sqlmock.New()

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	assert.Nil(t, err)

	varDAO := VariableDAOImpl{}
	varDAO.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	var v model.Variable
	v.ID = 999
	row := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at"}).
		AddRow(v.ID, v.CreatedAt, v.UpdatedAt, v.DeletedAt)

	mock.ExpectQuery(`SELECT (.*) FROM "variables"`).
		WillReturnRows(row)

	result, err := varDAO.GetByID(999)
	assert.Nil(t, err)
	assert.Equal(t, v.ID, result.ID)

	mock.ExpectationsWereMet()
}
