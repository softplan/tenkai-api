package repository

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/stretchr/testify/assert"
)

func getEnvironmentTestData() model.Environment {
	now := time.Now()
	item := model.Environment{}
	item.CreatedAt = now
	item.DeletedAt = nil
	item.UpdatedAt = now
	item.Group = "my-group"
	item.Name = "env-name"
	item.ClusterURI = "qwe"
	item.CACertificate = "asd"
	item.Token = "zxc"
	item.Namespace = "dev"
	item.Gateway = "dsa"
	item.ProductVersion = "19.0.1-0"
	item.CurrentRelease = "19.0.1"
	return item
}

func getUserTestData() model.User {
	var envs []model.Environment
	e := getEnvironmentTestData()
	e.ID = 999
	envs = append(envs, e)

	item := model.User{}
	item.ID = 998
	item.Email = "musk@mars.com"
	item.DefaultEnvironmentID = 999
	item.Environments = envs
	return item
}

func TestCreateEnvironment(t *testing.T) {

	db, mock, err := sqlmock.New()

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	assert.Nil(t, err)

	envDAO := EnvironmentDAOImpl{}
	envDAO.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)

	item := getEnvironmentTestData()

	mock.ExpectQuery(`INSERT INTO "environments"`).
		WithArgs(item.CreatedAt, item.UpdatedAt, item.DeletedAt, item.Group,
			item.Name, item.ClusterURI, item.CACertificate, item.Token,
			item.Namespace, item.Gateway, item.ProductVersion, item.CurrentRelease).
		WillReturnRows(rows)

	result, e := envDAO.CreateEnvironment(item)
	assert.Nil(t, e)
	assert.Equal(t, 1, result)

	mock.ExpectationsWereMet()

}

func TestCreateEnvironment_Error(t *testing.T) {

	db, mock, err := sqlmock.New()

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	assert.Nil(t, err)

	envDAO := EnvironmentDAOImpl{}
	envDAO.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	item := getEnvironmentTestData()

	mock.ExpectQuery(`INSERT INTO "environments"`).
		WithArgs(item.CreatedAt, item.UpdatedAt, item.DeletedAt, item.Group,
			item.Name, item.ClusterURI, item.CACertificate, item.Token,
			item.Namespace, item.Gateway, item.ProductVersion).
		WillReturnError(errors.New("mock error"))

	_, e := envDAO.CreateEnvironment(item)
	assert.Error(t, e)

	mock.ExpectationsWereMet()
}

func TestEditEnvironment(t *testing.T) {

	db, mock, err := sqlmock.New()

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	assert.Nil(t, err)

	envDAO := EnvironmentDAOImpl{}
	envDAO.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	item := getEnvironmentTestData()
	item.ID = 999

	mock.ExpectExec(`UPDATE "environments" SET (.*) WHERE (.*)`).
		WithArgs(item.CreatedAt, sqlmock.AnyArg(), item.DeletedAt, item.Group,
			item.Name, item.ClusterURI, item.CACertificate, item.Token,
			item.Namespace, item.Gateway, item.ProductVersion, item.CurrentRelease, item.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	result := envDAO.EditEnvironment(item)
	assert.Nil(t, result)

}

func TestDeleteEnvironment(t *testing.T) {

	db, mock, err := sqlmock.New()

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	assert.Nil(t, err)

	envDAO := EnvironmentDAOImpl{}
	envDAO.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	item := getEnvironmentTestData()
	item.ID = 999

	mock.ExpectExec(`DELETE FROM "environments" WHERE "environments"."id" = (.*)`).
		WithArgs(item.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	result := envDAO.DeleteEnvironment(item)
	assert.Nil(t, result)

}

func TestGetAllEnvironments(t *testing.T) {

	db, mock, err := sqlmock.New()

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	assert.Nil(t, err)

	envDAO := EnvironmentDAOImpl{}
	envDAO.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	user := getUserTestData()
	row1 := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "email", "default_environment_id"}).
		AddRow(user.ID, user.CreatedAt, user.UpdatedAt, user.DeletedAt, user.Email, user.DefaultEnvironmentID)

	mock.ExpectQuery(`SELECT (.*) FROM "users" 
		WHERE "users"."deleted_at" IS NULL AND \(\("users"."email" = (.*)\)\)
		ORDER BY "users"."id" ASC LIMIT 1`).
		WithArgs(user.Email).
		WillReturnRows(row1)

	e := getEnvironmentTestData()
	e.ID = 999
	row2 := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "group", "name", "cluster_uri", "ca_certificate", "token", "namespace", "gateway"}).
		AddRow(e.ID, e.CreatedAt, e.UpdatedAt, e.DeletedAt, e.Group, e.Name, e.ClusterURI, e.CACertificate, e.Token, e.Namespace, e.Gateway)

	mock.ExpectQuery(`SELECT (.*) FROM "environments"
		INNER JOIN "user_environment" ON "user_environment"."environment_id" = "environments"."id"
		WHERE "environments"."deleted_at" IS NULL AND \(\("user_environment"."user_id" IN (.*)\)\)`).
		WillReturnRows(row2)

	result, err := envDAO.GetAllEnvironments(user.Email)
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, e.ID, result[0].ID)

	mock.ExpectationsWereMet()
}

func TestGetAllEnvironmentsWithoutPrincipal(t *testing.T) {
	db, mock, err := sqlmock.New()

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	assert.Nil(t, err)

	envDAO := EnvironmentDAOImpl{}
	envDAO.Db = gormDB

	mock.MatchExpectationsInOrder(false)
	e := getEnvironmentTestData()
	e.ID = 999
	row := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "group", "name", "cluster_uri", "ca_certificate", "token", "namespace", "gateway"}).
		AddRow(e.ID, e.CreatedAt, e.UpdatedAt, e.DeletedAt, e.Group, e.Name, e.ClusterURI, e.CACertificate, e.Token, e.Namespace, e.Gateway)

	mock.ExpectQuery(`SELECT (.*) FROM "environments" WHERE "environments"."deleted_at" IS NULL`).
		WillReturnRows(row)

	result, err := envDAO.GetAllEnvironments("")
	assert.Nil(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, e.ID, result[0].ID)

	mock.ExpectationsWereMet()
}

func TestGetAllEnvironmentsWithoutPrincipal_ErrorNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	assert.Nil(t, err)

	envDAO := EnvironmentDAOImpl{}
	envDAO.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	mock.ExpectQuery(`SELECT (.*) FROM "environments" WHERE "environments"."deleted_at" IS NULL`).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := envDAO.GetAllEnvironments("")
	assert.Nil(t, err)
	assert.Empty(t, result)

	mock.ExpectationsWereMet()
}

func TestGetAllEnvironments_ErrorNotFound1(t *testing.T) {

	db, mock, err := sqlmock.New()

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	assert.Nil(t, err)

	envDAO := EnvironmentDAOImpl{}
	envDAO.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	user := getUserTestData()

	mock.ExpectQuery(`SELECT (.*) FROM "users"
		WHERE "users"."deleted_at" IS NULL AND \(\("users"."email" = (.*)\)\)
		ORDER BY "users"."id" ASC LIMIT 1`).
		WithArgs(user.Email).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := envDAO.GetAllEnvironments(user.Email)
	assert.Nil(t, err)
	assert.Empty(t, result)

	mock.ExpectationsWereMet()
}

func TestGetAllEnvironments_ErrorNotFound2(t *testing.T) {

	db, mock, err := sqlmock.New()

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	assert.Nil(t, err)

	envDAO := EnvironmentDAOImpl{}
	envDAO.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	user := getUserTestData()
	row1 := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "email", "default_environment_id"}).
		AddRow(user.ID, user.CreatedAt, user.UpdatedAt, user.DeletedAt, user.Email, user.DefaultEnvironmentID)

	mock.ExpectQuery(`SELECT (.*) FROM "users" 
		WHERE "users"."deleted_at" IS NULL AND \(\("users"."email" = (.*)\)\)
		ORDER BY "users"."id" ASC LIMIT 1`).
		WithArgs(user.Email).
		WillReturnRows(row1)

	mock.ExpectQuery(`SELECT (.*) FROM "environments"
		INNER JOIN "user_environment" ON "user_environment"."environment_id" = "environments"."id"
		WHERE "environments"."deleted_at" IS NULL AND \(\("user_environment"."user_id" IN (.*)\)\)`).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := envDAO.GetAllEnvironments(user.Email)
	assert.Nil(t, err)
	assert.Empty(t, result)

	mock.ExpectationsWereMet()
}

func TestGetByID(t *testing.T) {
	db, mock, err := sqlmock.New()

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	assert.Nil(t, err)

	envDAO := EnvironmentDAOImpl{}
	envDAO.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	e := getEnvironmentTestData()
	e.ID = 999
	row := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "group", "name", "cluster_uri", "ca_certificate", "token", "namespace", "gateway"}).
		AddRow(e.ID, e.CreatedAt, e.UpdatedAt, e.DeletedAt, e.Group, e.Name, e.ClusterURI, e.CACertificate, e.Token, e.Namespace, e.Gateway)

	mock.ExpectQuery(`SELECT (.*) FROM "environments" WHERE "environments"."deleted_at" IS NULL
		AND \(\("environments"."id" = 999\)\) ORDER BY "environments"."id" ASC LIMIT 1`).
		WillReturnRows(row)

	result, err := envDAO.GetByID(999)
	assert.Nil(t, err)
	assert.Equal(t, e.ID, result.ID)

	mock.ExpectationsWereMet()
}

func TestGetByID_Error(t *testing.T) {
	db, mock, err := sqlmock.New()

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	assert.Nil(t, err)

	envDAO := EnvironmentDAOImpl{}
	envDAO.Db = gormDB

	mock.MatchExpectationsInOrder(false)

	mock.ExpectQuery(`SELECT (.*) FROM "environments" WHERE "environments"."deleted_at" IS NULL
		AND \(\("environments"."id" = 999\)\) ORDER BY "environments"."id" ASC LIMIT 1`).
		WillReturnError(errors.New("mock error"))

	_, err = envDAO.GetByID(999)
	assert.Error(t, err)

	mock.ExpectationsWereMet()
}
