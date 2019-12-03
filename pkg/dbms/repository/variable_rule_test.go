package repository

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/stretchr/testify/assert"
)

func getVariableRule() model.VariableRule {
	var item model.VariableRule
	item.Name = "uriApi*"
	return item
}

func beforeVarRuleTest(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, VariableRuleDAOImpl, model.VariableRule) {
	db, mock, err := sqlmock.New()
	gormDB, err := gorm.Open("postgres", db)

	assert.Nil(t, err)

	dao := VariableRuleDAOImpl{}
	dao.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	item := getVariableRule()

	return gormDB, mock, dao, item
}

func TestCreateVariableRule(t *testing.T) {
	gormDB, mock, dao, item := beforeVarRuleTest(t)
	defer gormDB.Close()

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)

	mock.ExpectQuery(`INSERT INTO "variable_rules"`).
		WithArgs(AnyTime{}, AnyTime{}, nil, item.Name).
		WillReturnRows(rows)

	result, e := dao.CreateVariableRule(item)
	assert.Nil(t, e)
	assert.Equal(t, 1, result)

	mock.ExpectationsWereMet()
}

func TestCreateVariableRule_Error(t *testing.T) {
	gormDB, mock, dao, item := beforeVarRuleTest(t)
	defer gormDB.Close()

	mock.ExpectQuery(`INSERT INTO "variable_rules"`).
		WithArgs(AnyTime{}, AnyTime{}, nil, item.Name).
		WillReturnError(errors.New("some error"))

	_, e := dao.CreateVariableRule(item)
	assert.Error(t, e)

	mock.ExpectationsWereMet()
}

func TestEditVariableRule(t *testing.T) {
	gormDB, mock, dao, item := beforeVarRuleTest(t)
	defer gormDB.Close()

	item.ID = 999

	mock.ExpectExec(`UPDATE "variable_rules" SET (.*) WHERE (.*)`).
		WithArgs(AnyTime{}, nil, item.Name, item.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	e := dao.EditVariableRule(item)
	assert.Nil(t, e)

	mock.ExpectationsWereMet()
}

func TestDeleteVariableRule(t *testing.T) {
	gormDB, mock, dao, item := beforeVarRuleTest(t)
	defer gormDB.Close()

	item.ID = 999
	mock.ExpectExec(`DELETE FROM "variable_rules" WHERE (.*)`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := dao.DeleteVariableRule(999)
	assert.Nil(t, err)

	mock.ExpectationsWereMet()
}

func TestListVariableRules(t *testing.T) {
	gormDB, mock, dao, item := beforeVarRuleTest(t)
	defer gormDB.Close()

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(999, item.Name)

	item.ID = 999
	mock.ExpectQuery(`SELECT (.+) FROM "variable_rules" WHERE (.+)`).
		WillReturnRows(rows)

	result, err := dao.ListVariableRules()
	assert.Nil(t, err)
	assert.NotNil(t, result)

	mock.ExpectationsWereMet()
}

func TestListVariableRules_NotFound(t *testing.T) {
	gormDB, mock, dao, item := beforeVarRuleTest(t)
	defer gormDB.Close()

	item.ID = 999
	mock.ExpectQuery(`SELECT (.+) FROM "variable_rules" WHERE (.+)`).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := dao.ListVariableRules()
	assert.Nil(t, err)
	assert.Empty(t, result)

	mock.ExpectationsWereMet()
}

func TestListVariableRules_Error(t *testing.T) {
	gormDB, mock, dao, item := beforeVarRuleTest(t)
	defer gormDB.Close()

	item.ID = 999
	mock.ExpectQuery(`SELECT (.+) FROM "variable_rules" WHERE (.+)`).
		WillReturnError(errors.New("mock error"))

	_, err := dao.ListVariableRules()
	assert.Error(t, err)

	mock.ExpectationsWereMet()
}
