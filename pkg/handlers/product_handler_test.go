package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/softplan/tenkai-api/pkg/dbms/model"
	mockRepo "github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	"github.com/stretchr/testify/assert"
)

func TestSplitSrvNameIfNeeded(t *testing.T) {
	assert.Equal(t, "repo/my-chart", splitSrvNameIfNeeded("repo/my-chart - 0.1.0"))
	assert.Equal(t, "repo/my-chart", splitSrvNameIfNeeded("repo/my-chart"))
}

func TestSplitChartVersion(t *testing.T) {
	assert.Equal(t, "0.1.0", splitChartVersion("repo/my-chart - 0.1.0"))
	assert.Equal(t, "", splitChartVersion("repo/my-chart"))
}

func TestSplitChartRepo(t *testing.T) {
	assert.Equal(t, "repo", splitChartRepo("repo/my-chart - 0.1.0"))
	assert.Equal(t, "", splitChartRepo("my-chart"))
}

func TestGetChartLatestVersion(t *testing.T) {
	appContext := AppContext{}

	var sr1 model.SearchResult
	sr1.Name = "repo/my-chart"
	sr1.ChartVersion = "0.1.0"
	sr1.AppVersion = "1.0.0"
	sr1.Description = "This is my chart"

	var results []model.SearchResult
	results = append(results, sr1)

	latestVersion := appContext.getChartLatestVersion("repo/my-chart - 0.1.0", results)
	assert.Equal(t, "", latestVersion, "Should not have a latest version")

	var sr2 model.SearchResult
	sr2.Name = "repo/my-chart"
	sr2.ChartVersion = "0.2.0"
	sr2.AppVersion = "1.0.0"
	sr2.Description = "This is my chart"
	results = append(results, sr2)

	latestVersion = appContext.getChartLatestVersion("repo/my-chart - 0.1.0", results)
	assert.Equal(t, "0.2.0", latestVersion, "Latest version should be 0.2.0")
}

func Test_getNumberOfTag(t *testing.T) {
	assert.Equal(t, uint64(19030015000000), getNumberOfTag("19.3.0-15"))
	assert.Equal(t, uint64(20401025000000), getNumberOfTag("20.40.10-25"))
	assert.Equal(t, uint64(10000000000), getNumberOfTag("0.1.0-0"))
}

func TestNewProduct(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("CreateProduct", getProductWithoutID()).Return(999, nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/products", payload(getProductWithoutID()))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newProduct)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "CreateProduct", 1)
	assert.Equal(t, http.StatusCreated, rr.Code, "Response should be Created")
}

func TestNewProduct_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}
	rr := commonTestUnmarshalPayloadError(t, "/products", appContext.newProduct)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestNewProduct_CreateProductError(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("CreateProduct", getProductWithoutID()).Return(0, errors.New("Error saving product"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/products", payload(getProductWithoutID()))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newProduct)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestEditProduct(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("EditProduct", getProduct()).Return(nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/products/edit", payload(getProduct()))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editProduct)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "EditProduct", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok")
}

func TestEditProduct_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}
	rr := commonTestUnmarshalPayloadError(t, "/products/edit", appContext.editProduct)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestEditProduct_CreateProductError(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("EditProduct", getProduct()).Return(errors.New("Error saving product"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/products/edit", payload(getProduct()))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editProduct)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func getProduct() model.Product {
	var payload model.Product
	payload.ID = 999
	payload.Name = "my-product"
	return payload
}

func getProductWithoutID() model.Product {
	var payload model.Product
	payload.Name = "my-product"
	return payload
}
