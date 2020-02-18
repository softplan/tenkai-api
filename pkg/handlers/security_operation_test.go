package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	mockRepo "github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListSecurityOperation(t *testing.T) {
	appContext := AppContext{}

	so := mockSecurityOperations()

	var result []model.SecurityOperation
	result = append(result, so)

	mockSecOpDao := &mockRepo.SecurityOperationDAOInterface{}
	mockSecOpDao.On("List", mock.Anything).Return(result, nil)

	appContext.Repositories.SecurityOperationDAO = mockSecOpDao

	req, err := http.NewRequest("GET", "/security-operations", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listSecurityOperation)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok")

	response := string(rr.Body.Bytes())
	assert.Contains(t, response, `{"list":[{"ID":999,`)
	assert.Contains(t, response, `"name":"ONLY_DEPLOY",`)
	assert.Contains(t, response, `"policies":["ACTION_DEPLOY"]}]}`)

}

func TestListSecurityOperation_Error(t *testing.T) {
	appContext := AppContext{}

	mockSecOpDao := &mockRepo.SecurityOperationDAOInterface{}
	mockSecOpDao.On("List", mock.Anything).Return(nil, errors.New("some error"))

	appContext.Repositories.SecurityOperationDAO = mockSecOpDao

	req, err := http.NewRequest("GET", "/security-operations", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listSecurityOperation)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500")
}

func TestCreateOrUpdateSecurityOperation(t *testing.T) {
	appContext := AppContext{}

	p := mockSecurityOperations()

	mockSecOpDao := &mockRepo.SecurityOperationDAOInterface{}
	mockSecOpDao.On("CreateOrUpdate", p).Return(nil)

	appContext.Repositories.SecurityOperationDAO = mockSecOpDao

	req, err := http.NewRequest("POST", "/security-operations", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.createOrUpdateSecurityOperation)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code, "Response should be Created")
}

func TestCreateOrUpdateSecurityOperation_Unmarshal(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/security-operations", appContext.createOrUpdateSecurityOperation)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestCreateOrUpdateSecurityOperation_Error(t *testing.T) {
	appContext := AppContext{}

	p := mockSecurityOperations()

	mockSecOpDao := &mockRepo.SecurityOperationDAOInterface{}
	mockSecOpDao.On("CreateOrUpdate", p).Return(errors.New("some error"))

	appContext.Repositories.SecurityOperationDAO = mockSecOpDao

	req, err := http.NewRequest("POST", "/security-operations", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.createOrUpdateSecurityOperation)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500")
}

func TestDeleteSecurityOperation(t *testing.T) {
	appContext := AppContext{}

	mockSecOpDao := &mockRepo.SecurityOperationDAOInterface{}
	mockSecOpDao.On("Delete", mock.Anything).Return(nil)

	appContext.Repositories.SecurityOperationDAO = mockSecOpDao

	req, err := http.NewRequest("DELETE", "/security-operations/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/security-operations/{id}", appContext.deleteSecurityOperation).Methods("DELETE")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok")
}

func TestDeleteSecurityOperation_PrincipalError(t *testing.T) {
	appContext := AppContext{}
	mockSecOpDao := &mockRepo.SecurityOperationDAOInterface{}
	appContext.Repositories.SecurityOperationDAO = mockSecOpDao

	req, err := http.NewRequest("DELETE", "/security-operations/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/security-operations/{id}", appContext.deleteSecurityOperation).Methods("DELETE")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Response should be 401")
}

func TestDeleteSecurityOperation_Error(t *testing.T) {
	appContext := AppContext{}

	mockSecOpDao := &mockRepo.SecurityOperationDAOInterface{}
	mockSecOpDao.On("Delete", mock.Anything).Return(errors.New("some error"))

	appContext.Repositories.SecurityOperationDAO = mockSecOpDao

	req, err := http.NewRequest("DELETE", "/security-operations/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/security-operations/{id}", appContext.deleteSecurityOperation).Methods("DELETE")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500")
}
