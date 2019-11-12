package repository

import (
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

	_, err = produtDAO.CreateProductVersionCopying(*payload)
	assert.Nil(t, err)

	mock.ExpectationsWereMet()

}

func TestCreateProduct(t *testing.T) {

	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	produtDAO := ProductDAOImpl{}
	produtDAO.Db = gormDB

	product := model.Product{}
	product.Name = "xpto"

	rows := sqlmock.NewRows([]string{"id"}).AddRow(99)

	mock.ExpectQuery(`INSERT INTO "products"`).
		WithArgs(AnyTime{}, AnyTime{}, nil, product.Name).WillReturnRows(rows)

	_, err = produtDAO.CreateProduct(product)
	assert.Nil(t, err)

	mock.ExpectationsWereMet()

}

func TestEdit(t *testing.T) {

	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	produtDAO := ProductDAOImpl{}
	produtDAO.Db = gormDB

	product := model.Product{}
	product.Name = "xpto"
	product.Model.ID = 99
	product.ID = 1

	mock.ExpectExec(`UPDATE "products" SET (.*) WHERE (.*)`).
		WithArgs(AnyTime{}, nil, product.Name, product.ID).WillReturnResult(sqlmock.NewResult(1, 1))

	err = produtDAO.EditProduct(product)
	assert.Nil(t, err)

	productVersion := model.ProductVersion{}
	productVersion.Version = "200"
	productVersion.ProductID = 99
	productVersion.Model.ID = 1

	mock.ExpectExec(`UPDATE "product_versions" SET (.*) WHERE (.*)`).
		WithArgs(AnyTime{}, nil, productVersion.ProductID, productVersion.Date, productVersion.Version, productVersion.ID).WillReturnResult(sqlmock.NewResult(1, 1))

	err = produtDAO.EditProductVersion(productVersion)
	assert.Nil(t, err)

	productVersionService := model.ProductVersionService{}
	productVersionService.Model.ID = 1
	productVersionService.ServiceName = "alfa"
	productVersionService.ProductVersionID = 99
	productVersionService.DockerImageTag = "latest"
	productVersionService.LatestVersion = "latest"
	productVersionService.ChartLatestVersion = "latest"

	mock.ExpectExec(`UPDATE "product_version_services" SET (.*) WHERE (.*)`).
		WithArgs(AnyTime{}, nil, productVersionService.ProductVersionID, productVersionService.ServiceName, productVersionService.DockerImageTag, productVersionService.ID).WillReturnResult(sqlmock.NewResult(1, 1))

	err = produtDAO.EditProductVersionService(productVersionService)
	assert.Nil(t, err)

	mock.ExpectationsWereMet()
}
