package repository

import (
	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func getProductVersion() *model.ProductVersion {
	item := model.ProductVersion{}
	now := time.Now()
	item.CreatedAt = now
	item.DeletedAt = nil
	item.UpdatedAt = now
	item.ProductID = 1
	item.Date = now
	item.Version = "1.0"
	item.CopyLatestRelease = true
	return &item
}

type AnyTime struct{}

func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func TestCreateProductVersionCopying(t *testing.T) {

	payload := getProductVersion()

	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	produtDAO := ProductDAOImpl{}
	produtDAO.Db = gormDB

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)

	mock.ExpectQuery(`INSERT INTO "product_versions"`).
		WithArgs(payload.CreatedAt, payload.UpdatedAt, payload.DeletedAt, payload.ProductID, payload.Date, payload.Version).
		WillReturnRows(rows)

	rows2 := sqlmock.NewRows([]string{"id", "product_version_id"}).AddRow(1, 1)

	mock.ExpectQuery(`SELECT (.+) FROM "product_versions"`).
		WithArgs(payload.ProductID, 1).
		WillReturnRows(rows2)

	rows3 := sqlmock.NewRows([]string{"id", "product_version_id", "service_name", "docker_image_tag", "latest_version", "chart_latest_version"}).
		AddRow(1, 1, "alfa", "latest", "alfa", "latest")

	mock.ExpectQuery(`SELECT (.+) FROM "product_version_services"`).
		WithArgs(payload.ProductID).
		WillReturnRows(rows3)

	rows4 := sqlmock.NewRows([]string{"id"}).AddRow(1)

	mock.ExpectQuery(`INSERT INTO "product_version_services"`).
		WithArgs(AnyTime{}, AnyTime{}, nil, 1, "alfa", "latest").
		WillReturnRows(rows4)

	produtDAO.CreateProductVersionCopying(*payload)

	mock.ExpectationsWereMet()

}
