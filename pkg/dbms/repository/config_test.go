package repository

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetConfigByName(t *testing.T) {

	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	mock.MatchExpectationsInOrder(false)
	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	dao := ConfigDAOImpl{}
	dao.Db = gormDB

	mykey := "mykey"

	rows := sqlmock.NewRows([]string{"id", "name", "value"}).AddRow(1, mykey, "value")

	mock.ExpectQuery(`SELECT (.+) FROM "config_maps"`).
		WithArgs(mykey).
		WillReturnRows(rows)

	configMap, err := dao.GetConfigByName(mykey)
	assert.Nil(t, err)
	assert.NotNil(t, configMap)
	assert.Equal(t, mykey, configMap.Name)
	assert.Equal(t, "value", configMap.Value)

	mock.ExpectationsWereMet()

}

func TestCreateConfig(t *testing.T) {

	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	mock.MatchExpectationsInOrder(false)
	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	dao := ConfigDAOImpl{}
	dao.Db = gormDB

	item := model.ConfigMap{}
	item.Name = "mykey"
	item.Value = "myvalue"

	mock.ExpectQuery(`SELECT (.+) FROM "config_maps"`).
		WithArgs(item.Name).WillReturnError(gorm.ErrRecordNotFound)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)

	mock.ExpectQuery(`INSERT INTO "config_maps"`).
		WithArgs(AnyTime{}, AnyTime{}, nil, item.Name, item.Value).
		WillReturnRows(rows)

	i, err := dao.CreateOrUpdateConfig(item)
	assert.Nil(t, err)
	assert.NotEmpty(t, i)

	mock.ExpectationsWereMet()

}

func TestEditConfig(t *testing.T) {

	db, mock, err := sqlmock.New()
	assert.Nil(t, err)
	mock.MatchExpectationsInOrder(false)
	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	dao := ConfigDAOImpl{}
	dao.Db = gormDB

	item := model.ConfigMap{}
	item.Name = "mykey"
	item.Value = "myvalue"

	basicRows := sqlmock.NewRows([]string{"id", "name", "value"}).AddRow(1, item.Name, item.Value)

	mock.ExpectQuery(`SELECT (.+) FROM "config_maps"`).
		WithArgs(item.Name).WillReturnRows(basicRows)

	mock.ExpectExec(`UPDATE "config_maps" SET (.*) WHERE (.*)`).
		WithArgs(AnyTime{}, nil, item.Name, item.Value, 1).WillReturnResult(sqlmock.NewResult(1, 1))

	i, err := dao.CreateOrUpdateConfig(item)
	assert.Nil(t, err)
	assert.NotEmpty(t, i)

	mock.ExpectationsWereMet()

}
