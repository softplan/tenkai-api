package repository

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/stretchr/testify/assert"
)

func getValueRule() model.ValueRule {
	var item model.ValueRule
	item.Type = "StartsWith"
	item.Value = "https"
	item.VariableRuleID = 888
	return item
}

func beforeTest(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, ValueRuleDAOImpl, model.ValueRule) {
	db, mock, err := sqlmock.New()
	gormDB, err := gorm.Open("postgres", db)

	assert.Nil(t, err)

	dao := ValueRuleDAOImpl{}
	dao.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	item := getValueRule()

	return gormDB, mock, dao, item
}

func TestCreateValueRule(t *testing.T) {
	gormDB, mock, dao, item := beforeTest(t)
	defer gormDB.Close()

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)

	mock.ExpectQuery(`INSERT INTO "value_rules"`).
		WithArgs(AnyTime{}, AnyTime{}, nil, item.Type, item.Value, item.VariableRuleID).
		WillReturnRows(rows)

	result, e := dao.CreateValueRule(item)
	assert.Nil(t, e)
	assert.Equal(t, 1, result)

	mock.ExpectationsWereMet()
}

func TestCreateValueRule_Error(t *testing.T) {
	gormDB, mock, dao, item := beforeTest(t)
	defer gormDB.Close()

	mock.ExpectQuery(`INSERT INTO "value_rules"`).
		WithArgs(AnyTime{}, AnyTime{}, nil, item.Type, item.Value, item.VariableRuleID).
		WillReturnError(errors.New("some error"))

	_, e := dao.CreateValueRule(item)
	assert.Error(t, e)

	mock.ExpectationsWereMet()
}

func TestEditValueRule(t *testing.T) {
	gormDB, mock, dao, item := beforeTest(t)
	defer gormDB.Close()

	item.ID = 999

	mock.ExpectExec(`UPDATE "value_rules" SET (.*) WHERE (.*)`).
		WithArgs(AnyTime{}, nil, item.Type, item.Value, item.VariableRuleID, item.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	e := dao.EditValueRule(item)
	assert.Nil(t, e)

	mock.ExpectationsWereMet()
}

func TestDeleteValueRule(t *testing.T) {
	gormDB, mock, dao, item := beforeTest(t)
	defer gormDB.Close()

	item.ID = 999
	mock.ExpectExec(`DELETE FROM "value_rules" WHERE (.*)`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := dao.DeleteValueRule(999)
	assert.Nil(t, err)

	mock.ExpectationsWereMet()
}

func TestListValueRules(t *testing.T) {
	gormDB, mock, dao, item := beforeTest(t)
	defer gormDB.Close()

	rows := sqlmock.NewRows([]string{"id", "type", "value", "variable_rule_id"}).
		AddRow(999, item.Type, item.Value, item.VariableRuleID)

	item.ID = 999
	mock.ExpectQuery(`SELECT (.+) FROM "value_rules" WHERE (.+)`).
		WithArgs(999).WillReturnRows(rows)

	result, err := dao.ListValueRules(999)
	assert.Nil(t, err)
	assert.NotNil(t, result)

	mock.ExpectationsWereMet()
}

func TestListValueRules_NotFound(t *testing.T) {
	gormDB, mock, dao, item := beforeTest(t)
	defer gormDB.Close()

	item.ID = 999
	mock.ExpectQuery(`SELECT (.+) FROM "value_rules" WHERE (.+)`).
		WithArgs(999).WillReturnError(gorm.ErrRecordNotFound)

	result, err := dao.ListValueRules(999)
	assert.Nil(t, err)
	assert.Empty(t, result)

	mock.ExpectationsWereMet()
}

func TestListValueRules_Error(t *testing.T) {
	gormDB, mock, dao, item := beforeTest(t)
	defer gormDB.Close()

	item.ID = 999
	mock.ExpectQuery(`SELECT (.+) FROM "value_rules" WHERE (.+)`).
		WithArgs(999).WillReturnError(errors.New("mock error"))

	_, err := dao.ListValueRules(999)
	assert.Error(t, err)

	mock.ExpectationsWereMet()
}
