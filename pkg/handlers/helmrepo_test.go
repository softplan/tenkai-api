package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/constraints"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
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

	req, err := http.NewRequest("POST", "/repositories", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, []string{constraints.TenkaiAdmin})

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
