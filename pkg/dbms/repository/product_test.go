package repository

import (
	"errors"
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
	item.Locked = false
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
		WithArgs(payload.CreatedAt, payload.UpdatedAt, payload.DeletedAt, payload.ProductID, payload.Date, payload.Version, payload.Locked).
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
	product.ID = 99

	mock.ExpectExec(`UPDATE "products" SET (.*) WHERE (.*)`).
		WithArgs(AnyTime{}, nil, product.Name, product.ID).WillReturnResult(sqlmock.NewResult(1, 1))

	err = produtDAO.EditProduct(product)
	assert.Nil(t, err)

	productVersion := model.ProductVersion{}
	productVersion.Version = "19.3.0-0"
	productVersion.ProductID = 99
	productVersion.ID = 1
	productVersion.Locked = false

	mock.ExpectExec(`UPDATE "product_versions"`).
		WithArgs(AnyTime{}, nil, product.ID, AnyTime{},
			productVersion.Version, productVersion.Locked, productVersion.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

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
		WithArgs(AnyTime{}, nil, productVersionService.ProductVersionID, productVersionService.ServiceName,
			productVersionService.DockerImageTag, productVersionService.ID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = produtDAO.EditProductVersionService(productVersionService)
	assert.Nil(t, err)

	mock.ExpectationsWereMet()
}

func TestListProducts(t *testing.T) {
	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	produtDAO := ProductDAOImpl{}
	produtDAO.Db = gormDB

	rows1 := sqlmock.NewRows([]string{"id", "name"}).AddRow(999, "my product")

	mock.ExpectQuery(`SELECT (.*) FROM "products"`).
		WillReturnRows(rows1)

	result, err := produtDAO.ListProducts()
	assert.Nil(t, err)
	assert.NotNil(t, result)

	mock.ExpectationsWereMet()
}

func TestListProducts_ErrorNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	produtDAO := ProductDAOImpl{}
	produtDAO.Db = gormDB

	mock.ExpectQuery(`SELECT (.*) FROM "products"`).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := produtDAO.ListProducts()
	assert.Nil(t, err)
	assert.NotNil(t, result)

	mock.ExpectationsWereMet()
}

func TestListProducts_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	produtDAO := ProductDAOImpl{}
	produtDAO.Db = gormDB

	mock.ExpectQuery(`SELECT (.*) FROM "products"`).
		WillReturnError(errors.New("mock error"))

	_, err = produtDAO.ListProducts()
	assert.Error(t, err)

	mock.ExpectationsWereMet()
}

func TestListProductsVersions(t *testing.T) {
	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	produtDAO := ProductDAOImpl{}
	produtDAO.Db = gormDB

	rows1 := sqlmock.NewRows([]string{"id", "product_id", "version"}).AddRow(888, 999, "19.3.0-0")

	mock.ExpectQuery(`SELECT (.*) FROM "product_versions"`).
		WithArgs(999).
		WillReturnRows(rows1)

	result, err := produtDAO.ListProductsVersions(999)
	assert.Nil(t, err)
	assert.NotNil(t, result)

	mock.ExpectationsWereMet()
}

func TestListProductsVersions_ErrorNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	produtDAO := ProductDAOImpl{}
	produtDAO.Db = gormDB

	mock.ExpectQuery(`SELECT (.*) FROM "product_versions"`).
		WithArgs(999).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := produtDAO.ListProductsVersions(999)
	assert.Nil(t, err)
	assert.Empty(t, result)

	mock.ExpectationsWereMet()
}

func TestListProductsVersions_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	produtDAO := ProductDAOImpl{}
	produtDAO.Db = gormDB

	mock.ExpectQuery(`SELECT (.*) FROM "product_versions"`).
		WithArgs(999).
		WillReturnError(errors.New("mock error"))

	_, err = produtDAO.ListProductsVersions(999)
	assert.Error(t, err)

	mock.ExpectationsWereMet()
}

func TestListProductVersionsByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	produtDAO := ProductDAOImpl{}
	produtDAO.Db = gormDB

	rows1 := sqlmock.NewRows([]string{"id", "product_id", "version"}).AddRow(888, 999, "19.3.0-0")

	mock.ExpectQuery(`SELECT (.*) FROM "product_versions"`).
		WillReturnRows(rows1)

	result, err := produtDAO.ListProductVersionsByID(999)
	assert.Nil(t, err)
	assert.NotNil(t, result)

	mock.ExpectationsWereMet()
}

func TestListProductVersionsByID_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	produtDAO := ProductDAOImpl{}
	produtDAO.Db = gormDB

	mock.ExpectQuery(`SELECT (.*) FROM "product_versions"`).
		WillReturnError(errors.New("mock error"))

	_, err = produtDAO.ListProductVersionsByID(999)
	assert.Error(t, err)

	mock.ExpectationsWereMet()
}

func TestDeleteProduct(t *testing.T) {
	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	produtDAO := ProductDAOImpl{}
	produtDAO.Db = gormDB

	mock.ExpectExec(`DELETE FROM "products"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = produtDAO.DeleteProduct(999)
	assert.Nil(t, err)

	mock.ExpectationsWereMet()
}

func TestDeleteProductVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	produtDAO := ProductDAOImpl{}
	produtDAO.Db = gormDB

	mock.ExpectExec(`DELETE FROM "product_versions"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = produtDAO.DeleteProductVersion(999)
	assert.Nil(t, err)

	mock.ExpectationsWereMet()
}

func TestDeleteProductVersionService(t *testing.T) {
	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	produtDAO := ProductDAOImpl{}
	produtDAO.Db = gormDB

	mock.ExpectExec(`DELETE FROM "product_version_services"`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = produtDAO.DeleteProductVersionService(999)
	assert.Nil(t, err)

	mock.ExpectationsWereMet()
}

func TestListProductsVersionServices(t *testing.T) {
	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	produtDAO := ProductDAOImpl{}
	produtDAO.Db = gormDB

	rows1 := sqlmock.NewRows([]string{"id", "product_version_id", "service_name", "docker_image_tag", "latest_version", "chart_latest_version"}).
		AddRow(1, 1, "alfa", "latest", "alfa", "latest")

	mock.ExpectQuery(`SELECT (.+) FROM "product_version_services"`).
		WithArgs(999).
		WillReturnRows(rows1)

	result, err := produtDAO.ListProductsVersionServices(999)
	assert.Nil(t, err)
	assert.NotNil(t, result)

	mock.ExpectationsWereMet()
}

func TestListProductsVersionServices_ErrorNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	produtDAO := ProductDAOImpl{}
	produtDAO.Db = gormDB

	mock.ExpectQuery(`SELECT (.+) FROM "product_version_services"`).
		WithArgs(999).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := produtDAO.ListProductsVersionServices(999)
	assert.Nil(t, err)
	assert.Empty(t, result)

	mock.ExpectationsWereMet()
}

func TestListProductsVersionServices_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	produtDAO := ProductDAOImpl{}
	produtDAO.Db = gormDB

	mock.ExpectQuery(`SELECT (.+) FROM "product_version_services"`).
		WithArgs(999).
		WillReturnError(errors.New("mock error"))

	_, err = produtDAO.ListProductsVersionServices(999)
	assert.Error(t, err)

	mock.ExpectationsWereMet()
}

func TestListProductVersionServicesLatest(t *testing.T) {
	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	produtDAO := ProductDAOImpl{}
	produtDAO.Db = gormDB

	productID := 999
	productVersionID := 777

	rows1 := sqlmock.NewRows([]string{"id", "product_id", "version"}).AddRow(productVersionID, productID, "19.0.1-0")

	mock.ExpectQuery(`SELECT (.+) FROM "product_versions"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows1)

	rows2 := sqlmock.NewRows([]string{"id", "product_version_id", "service_name", "docker_image_tag", "latest_version", "chart_latest_version"}).
		AddRow(1, productVersionID, "alfa", "latest", "alfa", "latest")

	mock.ExpectQuery(`SELECT (.+) FROM "product_version_services"`).
		WithArgs(productVersionID).
		WillReturnRows(rows2)

	result, err := produtDAO.ListProductVersionServicesLatest(productID, productVersionID)
	assert.Nil(t, err)
	assert.NotNil(t, result)

	mock.ExpectationsWereMet()
}

func TestListProductVersionServicesLatest_ErrorNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	produtDAO := ProductDAOImpl{}
	produtDAO.Db = gormDB

	productID := 999
	productVersionID := 777

	mock.ExpectQuery(`SELECT (.+) FROM "product_versions"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(gorm.ErrRecordNotFound)

	result, err := produtDAO.ListProductVersionServicesLatest(productID, productVersionID)
	assert.Nil(t, err)
	assert.Empty(t, result)

	mock.ExpectationsWereMet()
}

func TestListProductVersionServicesLatest_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	mock.MatchExpectationsInOrder(false)
	assert.Nil(t, err)

	gormDB, err := gorm.Open("postgres", db)
	defer gormDB.Close()

	produtDAO := ProductDAOImpl{}
	produtDAO.Db = gormDB

	productID := 999
	productVersionID := 777

	rows1 := sqlmock.NewRows([]string{"id", "product_id", "version"}).AddRow(productVersionID, productID, "19.0.1-0")

	mock.ExpectQuery(`SELECT (.+) FROM "product_versions"`).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(rows1)

	mock.ExpectQuery(`SELECT (.+) FROM "product_version_services"`).
		WithArgs(productVersionID).
		WillReturnError(errors.New("mock error"))

	_, err = produtDAO.ListProductVersionServicesLatest(productID, productVersionID)
	assert.Error(t, err)

	mock.ExpectationsWereMet()
}
