package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	mockSvc "github.com/softplan/tenkai-api/pkg/service/_helm/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListRepositories(t *testing.T) {

	appContext := AppContext{}

	mockHelmSvc := mockHelmSvcWithLotOfThings(&appContext)
	appContext.HelmServiceAPI = mockHelmSvc

	req, err := http.NewRequest("GET", "/repositories", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/repositories", appContext.listRepositories).Methods("GET")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
	mockHelmSvc.AssertNumberOfCalls(t, "GetRepositories", 1)

}

func TestNewRepository(t *testing.T) {

	appContext := AppContext{}

	var payload model.Repository
	payload.URL = "http://abacaxi"
	payload.Username = "alfa"
	payload.Password = "beta"
	payload.Name = "repo"

	payloadStr, _ := json.Marshal(payload)

	mockHelmSvc := mockHelmSvcWithLotOfThings(&appContext)
	appContext.HelmServiceAPI = mockHelmSvc
	appContext.RabbitImpl = getMockRabbitMQ()

	req, err := http.NewRequest("POST", "/repositories", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/repositories", appContext.newRepository).Methods("POST")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
	mockHelmSvc.AssertNumberOfCalls(t, "AddRepository", 1)

}

func TestSetDefaultRepo(t *testing.T) {

	appContext := AppContext{}

	var payload model.DefaultRepoRequest
	payload.Reponame = "abacaxi"

	payloadStr, _ := json.Marshal(payload)

	mockConfigDAO := mocks.ConfigDAOInterface{}
	mockConfigDAO.On("CreateOrUpdateConfig", mock.Anything).Return(-1, nil)
	appContext.Repositories.ConfigDAO = &mockConfigDAO

	req, err := http.NewRequest("POST", "/repo/default", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/repo/default", appContext.setDefaultRepo).Methods("POST")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
	mockConfigDAO.AssertNumberOfCalls(t, "CreateOrUpdateConfig", 1)

}

func TestSetDefaultRepo_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/repo/default", appContext.setDefaultRepo)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestSetDefaultRepo_Error(t *testing.T) {

	appContext := AppContext{}

	var payload model.DefaultRepoRequest
	payload.Reponame = "abacaxi"

	payloadStr, _ := json.Marshal(payload)

	mockConfigDAO := mocks.ConfigDAOInterface{}
	mockConfigDAO.On("CreateOrUpdateConfig", mock.Anything).Return(0, errors.New("some error"))
	appContext.Repositories.ConfigDAO = &mockConfigDAO

	req, err := http.NewRequest("POST", "/repo/default", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/repo/default", appContext.setDefaultRepo).Methods("POST")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
	mockConfigDAO.AssertNumberOfCalls(t, "CreateOrUpdateConfig", 1)
}

func TestGetDefaultRepo(t *testing.T) {
	appContext := AppContext{}

	result := model.ConfigMap{}
	result.Value = "abc"
	result.Name = "xpto"

	configDAO := mocks.ConfigDAOInterface{}
	configDAO.On("GetConfigByName", mock.Anything).Return(result, nil)
	appContext.Repositories.ConfigDAO = &configDAO

	req, err := http.NewRequest("GET", "/repo/default", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.getDefaultRepo)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")

	response := string(rr.Body.Bytes())
	assert.Contains(t, response, `{"ID":0,`)
	assert.Contains(t, response, `"name":"xpto","value":"abc"}`)
	configDAO.AssertNumberOfCalls(t, "GetConfigByName", 1)
}

func TestGetDefaultRepo_GetConfigByNameError(t *testing.T) {
	appContext := AppContext{}

	configDAO := mocks.ConfigDAOInterface{}
	configDAO.On("GetConfigByName", mock.Anything).Return(model.ConfigMap{}, errors.New("some error"))
	appContext.Repositories.ConfigDAO = &configDAO

	req, err := http.NewRequest("GET", "/repo/default", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.getDefaultRepo)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
	configDAO.AssertNumberOfCalls(t, "GetConfigByName", 1)
}

func TestDeleteRepository(t *testing.T) {
	appContext := AppContext{}

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	mockHelmSvc.On("RemoveRepository", "xyz").Return(nil)
	appContext.HelmServiceAPI = mockHelmSvc
	appContext.RabbitImpl = getMockRabbitMQ()
	req, err := http.NewRequest("DELETE", "/repositories/xyz", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/repositories/{name}", appContext.deleteRepository).Methods("DELETE")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")
	mockHelmSvc.AssertNumberOfCalls(t, "RemoveRepository", 1)
}

func TestDeleteRepository_Unauthorized(t *testing.T) {
	appContext := AppContext{}

	req, err := http.NewRequest("DELETE", "/repositories/xyz", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/repositories/{name}", appContext.deleteRepository).Methods("DELETE")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Response should be 401.")
}

func TestDeleteRepository_RemoveRepositoryError(t *testing.T) {
	appContext := AppContext{}

	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	mockHelmSvc.On("RemoveRepository", "xyz").Return(errors.New("some error"))
	appContext.HelmServiceAPI = mockHelmSvc

	req, err := http.NewRequest("DELETE", "/repositories/xyz", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/repositories/{name}", appContext.deleteRepository).Methods("DELETE")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
	mockHelmSvc.AssertNumberOfCalls(t, "RemoveRepository", 1)
}
