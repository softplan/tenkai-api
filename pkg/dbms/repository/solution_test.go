package repository

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func initTest(t *testing.T) (*sql.DB, sqlmock.Sqlmock, SolutionDAOInterface, error) {

	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)

	dao := SolutionDAOImpl{}
	dao.Db = gormDB

	return db, mock, dao, err

}

func buildSolution() model.Solution {
	s := model.Solution{}
	s.Name = "my solution"
	s.Team = "Brazil"
	return s
}

func TestCreateCreateSolution(t *testing.T) {

	db, mock, dao, err := initTest(t)
	defer db.Close()

	s := buildSolution()

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)

	mock.ExpectQuery(`INSERT INTO "solutions"`).
		WithArgs(AnyTime{}, AnyTime{}, nil, s.Name, s.Team).
		WillReturnRows(rows)

	res, err := dao.CreateSolution(s)
	assert.Nil(t, err)
	assert.NotEmpty(t, res)

	mock.ExpectationsWereMet()

}

func TestCreateCreateSolutionWithError(t *testing.T) {

	db, mock, dao, err := initTest(t)
	defer db.Close()
	s := buildSolution()

	mock.ExpectQuery(`INSERT INTO "solutions"`).
		WithArgs(AnyTime{}, AnyTime{}, nil, s.Name, s.Team).WillReturnError(errors.New("Error"))

	res, err := dao.CreateSolution(s)
	assert.NotNil(t, err)
	assert.Equal(t, -1, res)

	mock.ExpectationsWereMet()

}

func TestEditSolution(t *testing.T) {

	db, mock, dao, err := initTest(t)
	defer db.Close()
	s := buildSolution()
	s.ID = 10

	mock.ExpectExec(`UPDATE "solutions" SET (.*) WHERE (.*)`).
		WithArgs(AnyTime{}, nil, s.Name, s.Team, s.ID).WillReturnResult(sqlmock.NewResult(1, 1))

	err = dao.EditSolution(s)
	assert.Nil(t, err)

	mock.ExpectationsWereMet()

}

func TestDeleteSolutionError(t *testing.T) {

	db, mock, dao, err := initTest(t)
	defer db.Close()
	s := buildSolution()
	s.ID = 1

	mock.ExpectQuery(`DELETE FROM "solutions"`).WillReturnError(errors.New("error"))

	err = dao.DeleteSolution(int(s.ID))
	assert.NotNil(t, err)

	mock.ExpectationsWereMet()

}

func TestListSolutionsError(t *testing.T) {

	db, mock, dao, err := initTest(t)
	defer db.Close()

	mock.ExpectQuery(`SELECT (.*) FROM "users"`).WillReturnError(gorm.ErrRecordNotFound)

	_, err = dao.ListSolutions()
	assert.NotNil(t, err)

	mock.ExpectationsWereMet()

}
