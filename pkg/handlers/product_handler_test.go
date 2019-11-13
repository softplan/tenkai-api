package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	mockRepo "github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	mockSvc "github.com/softplan/tenkai-api/pkg/service/_helm/mocks"
	"github.com/softplan/tenkai-api/pkg/service/docker/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
	rr := testUnmarshalPayloadError(t, "/products", appContext.newProduct)
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
	rr := testUnmarshalPayloadError(t, "/products/edit", appContext.editProduct)
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

func TestDeleteProduct(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("DeleteProduct", 999).Return(nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("DELETE", "/products/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/products/{id}", appContext.deleteProduct).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "DeleteProduct", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
}

func TestDeleteProduct_Error(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("DeleteProduct", 999).Return(errors.New("Error deleting product"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("DELETE", "/products/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/products/{id}", appContext.deleteProduct).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "DeleteProduct", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestListProducts(t *testing.T) {
	result := &model.ProductRequestReponse{}
	result.List = append(result.List, getProduct())

	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProducts").Return(result.List, nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("GET", "/products", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listProducts)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProducts", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok")

	response := string(rr.Body.Bytes())
	assert.Contains(t, response, `{"list":[{"ID":999,`)
	assert.Contains(t, response, `"name":"my-product"}]}`)
}

func TestListProducts_Error(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProducts").Return(nil, errors.New("Error listing product"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("GET", "/products", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listProducts)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProducts", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestNewProductVersion(t *testing.T) {
	appContext := AppContext{}

	pv := getProductVersionWithoutID(false)
	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("CreateProductVersionCopying", mock.Anything).Return(999, nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/productVersions", payload(pv))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newProductVersion)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "CreateProductVersionCopying", 1)
	assert.Equal(t, http.StatusCreated, rr.Code, "Response should be Created")
}

func TestNewProductVersion_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/productVersions", appContext.newProductVersion)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestNewProductVersion_Error(t *testing.T) {
	appContext := AppContext{}

	pv := getProductVersionWithoutID(false)
	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("CreateProductVersionCopying", mock.Anything).Return(0, errors.New("Some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/productVersions", payload(pv))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newProductVersion)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "CreateProductVersionCopying", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500")
}

func TestDeleteProductVersion(t *testing.T) {
	appContext := AppContext{}

	childs := getProductVersionSvcReqResp()

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductsVersionServices", 999).Return(childs.List, nil)
	mockProductDAO.On("DeleteProductVersionService", 888).Return(nil)
	mockProductDAO.On("DeleteProductVersion", 999).Return(nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("DELETE", "/productVersions/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersions/{id}", appContext.deleteProductVersion).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductsVersionServices", 1)
	mockProductDAO.AssertNumberOfCalls(t, "DeleteProductVersionService", 1)
	mockProductDAO.AssertNumberOfCalls(t, "DeleteProductVersion", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")
}

func TestDeleteProductVersion_ListProductsVersionServicesError(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductsVersionServices", 999).Return(nil, errors.New("Some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("DELETE", "/productVersions/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersions/{id}", appContext.deleteProductVersion).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductsVersionServices", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDeleteProductVersion_DeleteProductVersionServiceError(t *testing.T) {
	appContext := AppContext{}

	childs := getProductVersionSvcReqResp()

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductsVersionServices", 999).Return(childs.List, nil)
	mockProductDAO.On("DeleteProductVersionService", 888).Return(errors.New("Some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("DELETE", "/productVersions/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersions/{id}", appContext.deleteProductVersion).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductsVersionServices", 1)
	mockProductDAO.AssertNumberOfCalls(t, "DeleteProductVersionService", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDeleteProductVersion_DeleteProductVersionError(t *testing.T) {
	appContext := AppContext{}

	childs := getProductVersionSvcReqResp()

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductsVersionServices", 999).Return(childs.List, nil)
	mockProductDAO.On("DeleteProductVersionService", 888).Return(nil)
	mockProductDAO.On("DeleteProductVersion", 999).Return(errors.New("Some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("DELETE", "/productVersions/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersions/{id}", appContext.deleteProductVersion).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductsVersionServices", 1)
	mockProductDAO.AssertNumberOfCalls(t, "DeleteProductVersionService", 1)
	mockProductDAO.AssertNumberOfCalls(t, "DeleteProductVersion", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestListProductVersions(t *testing.T) {
	appContext := AppContext{}

	result := getProductVersionReqResp()

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductsVersions", 777).Return(result.List, nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("GET", "/productVersions/?productId=777", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listProductVersions)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductsVersions", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")

	response := string(rr.Body.Bytes())
	assert.Contains(t, response, `{"list":[{"ID":777,`)
	assert.Contains(t, response, `"version":"19.0.1-0",`)
	assert.Contains(t, response, `"copyLatestRelease":false}]}`)
}

func TestListProductVersions_QueryError(t *testing.T) {
	appContext := AppContext{}

	req, err := http.NewRequest("GET", "/productVersions/?foo=bar", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listProductVersions)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestListProductVersions_ListProductsVersionsError(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductsVersions", 777).Return(nil, errors.New("some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("GET", "/productVersions/?productId=777", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listProductVersions)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductsVersions", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestListProductVersionServices(t *testing.T) {
	appContext := AppContext{}

	var pvs []model.ProductVersionService
	pvs = append(pvs, getProductVersionSvc())

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductsVersionServices", 999).Return(pvs, nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	data := getHelmSearchResult()
	mockHelmSvc.On("SearchCharts", mock.Anything, false).Return(&data)
	appContext.HelmServiceAPI = mockHelmSvc

	appContext.ChartImageCache.Store("repo/my-chart", "myrepo.com/my-chart")

	mockDockerSvc := mockGetDockerTagsWithDate(&appContext, getTagResponse("19.0.2-0"))

	req, err := http.NewRequest("GET", "/productVersionServices/?productVersionId=999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listProductVersionServices)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductsVersionServices", 1)
	mockHelmSvc.AssertNumberOfCalls(t, "SearchCharts", 1)
	mockDockerSvc.AssertNumberOfCalls(t, "GetDockerTagsWithDate", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")

	response := string(rr.Body.Bytes())
	assert.Contains(t, response, `{"list":[{"ID":888,`)
	assert.Contains(t, response, `"productVersionId":999,`)
	assert.Contains(t, response, `"serviceName":"repo/my-chart - 0.1.0",`)
	assert.Contains(t, response, `"dockerImageTag":"19.0.1-0"`)
	assert.Contains(t, response, `"latestVersion":"19.0.2-0",`)
	assert.Contains(t, response, `"chartLatestVersion":"1.0"}]}`)
}

func TestListProductVersionServices_QueryError(t *testing.T) {
	appContext := AppContext{}

	req, err := http.NewRequest("GET", "/productVersionServices/?foo=bar", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listProductVersionServices)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestListProductVersionServices_ListProdVerSvcError(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("ListProductsVersionServices", 999).Return(nil, errors.New("some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("GET", "/productVersionServices/?productVersionId=999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listProductVersionServices)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "ListProductsVersionServices", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be Ok.")
}

func TestNewProductVersionService(t *testing.T) {
	appContext := AppContext{}

	pvs := getProductVersionSvc()
	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("CreateProductVersionService", pvs).Return(888, nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/productVersionServices", payload(pvs))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newProductVersionService)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "CreateProductVersionService", 1)

	assert.Equal(t, http.StatusCreated, rr.Code, "Response should be Created.")
}

func TestNewProductVersionService_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/productVersionServices", appContext.newProductVersionService)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestNewProductVersionService_CreateProductVersionServiceError(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("CreateProductVersionService", mock.Anything).Return(0, errors.New("some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/productVersionServices", payload(getProductVersionSvc()))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newProductVersionService)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "CreateProductVersionService", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be Created.")
}

func TestEditProductVersionService(t *testing.T) {
	appContext := AppContext{}

	pvs := getProductVersionSvc()
	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("EditProductVersionService", pvs).Return(nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/productVersionServices/edit", payload(pvs))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editProductVersionService)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "EditProductVersionService", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")
}

func TestEditProductVersionService_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/productVersionServices/edit", appContext.editProductVersionService)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestEditProductVersionService_EditProductVersionServiceError(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("EditProductVersionService", mock.Anything).Return(errors.New("some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("POST", "/productVersionServices/edit", payload(getProductVersionSvc()))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editProductVersionService)
	handler.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "EditProductVersionService", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDeleteProductVersionService(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("DeleteProductVersionService", 888).Return(nil)
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("DELETE", "/productVersionServices/888", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersionServices/{id}", appContext.deleteProductVersionService).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "DeleteProductVersionService", 1)

	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")
}

func TestDeleteProductVersionService_DeleteProductVersionServiceError(t *testing.T) {
	appContext := AppContext{}

	mockProductDAO := &mockRepo.ProductDAOInterface{}
	mockProductDAO.On("DeleteProductVersionService", 888).Return(errors.New("some error"))
	appContext.Repositories.ProductDAO = mockProductDAO

	req, err := http.NewRequest("DELETE", "/productVersionServices/888", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/productVersionServices/{id}", appContext.deleteProductVersionService).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockProductDAO.AssertNumberOfCalls(t, "DeleteProductVersionService", 1)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestVerifyNewVersion(t *testing.T) {
	appContext := AppContext{}

	appContext.ChartImageCache.Store("repo/my-chart - 0.1.0", "myrepo.com/my-chart")

	mockDockerSvc := mockGetDockerTagsWithDate(&appContext, getTagResponse("19.0.2-0"))

	version, err := appContext.verifyNewVersion("repo/my-chart - 0.1.0", "19.0.1-0")
	assert.NoError(t, err)
	assert.NotNil(t, version)

	assert.Equal(t, "19.0.2-0", version)

	mockDockerSvc.AssertNumberOfCalls(t, "GetDockerTagsWithDate", 1)
}

func TestVerifyNewVersion_NotOk(t *testing.T) {
	appContext := AppContext{}

	appContext.ChartImageCache.Store("foo", "bar")

	mockDockerSvc := mockGetDockerTagsWithDate(&appContext, getTagResponse("19.0.2-0"))
	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	appContext.HelmServiceAPI = mockHelmSvc

	bytes := []byte("{\"image\":{\"repository\":\"myrepo.com/my-chart\"}}")
	mockHelmSvc.On("GetValues", "repo/my-chart - 0.1.0", "0").Return(bytes, nil)

	version, err := appContext.verifyNewVersion("repo/my-chart - 0.1.0", "19.0.1-0")
	assert.NoError(t, err)
	assert.NotNil(t, version)

	assert.Equal(t, "19.0.2-0", version)

	mockDockerSvc.AssertNumberOfCalls(t, "GetDockerTagsWithDate", 1)
}

func TestVerifyNewVersion_NotOk_Error(t *testing.T) {
	appContext := AppContext{}

	appContext.ChartImageCache.Store("repo/my-chart - 0.1.0", "")

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	appContext.HelmServiceAPI = mockHelmSvc

	mockHelmSvc.On("GetValues", mock.Anything, mock.Anything).Return(nil, errors.New("some error"))

	version, err := appContext.verifyNewVersion("repo/my-chart - 0.1.0", "19.0.1-0")
	assert.Error(t, err)
	assert.NotNil(t, version)

	assert.Equal(t, "", version)
}

func TestVerifyNewVersion_NotOk_UnmarshalError(t *testing.T) {
	appContext := AppContext{}

	appContext.ChartImageCache.Store("repo/my-chart - 0.1.0", "")

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	appContext.HelmServiceAPI = mockHelmSvc

	bytes := []byte(`["foo":"baz"]`)
	mockHelmSvc.On("GetValues", "repo/my-chart - 0.1.0", "0").Return(bytes, nil)

	version, err := appContext.verifyNewVersion("repo/my-chart - 0.1.0", "19.0.1-0")
	assert.Error(t, err)
	assert.NotNil(t, version)
	assert.Equal(t, "", version)
}

func TestVerifyNewVersion_GetDockerTagsWithDateError(t *testing.T) {
	appContext := AppContext{}

	appContext.ChartImageCache.Store("repo/my-chart - 0.1.0", "myrepo.com/my-chart")

	mockDockerSvc := &mocks.DockerServiceInterface{}
	mockDockerSvc.On("GetDockerTagsWithDate", mock.Anything, mock.Anything, mock.Anything).
		Return(nil, errors.New("some error"))
	appContext.DockerServiceAPI = mockDockerSvc

	version, err := appContext.verifyNewVersion("repo/my-chart - 0.1.0", "19.0.1-0")
	assert.Error(t, err)
	assert.NotNil(t, version)
	assert.Equal(t, "", version)

	mockDockerSvc.AssertNumberOfCalls(t, "GetDockerTagsWithDate", 1)
}

func TestVerifyNewVersion_NoNewVersion(t *testing.T) {
	appContext := AppContext{}

	appContext.ChartImageCache.Store("repo/my-chart - 0.1.0", "myrepo.com/my-chart")

	mockDockerSvc := mockGetDockerTagsWithDate(&appContext, getTagResponse("19.0.1-0"))

	version, err := appContext.verifyNewVersion("repo/my-chart - 0.1.0", "19.0.1-0")
	assert.NoError(t, err)
	assert.NotNil(t, version)

	assert.Equal(t, "", version)

	mockDockerSvc.AssertNumberOfCalls(t, "GetDockerTagsWithDate", 1)
}

func mockGetDockerTagsWithDate(appContext *AppContext, result *model.ListDockerTagsResult) *mocks.DockerServiceInterface {
	mockDockerSvc := &mocks.DockerServiceInterface{}
	mockDockerSvc.On("GetDockerTagsWithDate", mock.Anything, mock.Anything, mock.Anything).Return(result, nil)
	appContext.DockerServiceAPI = mockDockerSvc

	return mockDockerSvc
}

func getTagResponse(tag string) *model.ListDockerTagsResult {
	result := &model.ListDockerTagsResult{}

	var tr model.TagResponse
	tr.Tag = tag
	tr.Created = time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)
	result.TagResponse = append(result.TagResponse, tr)

	return result
}
