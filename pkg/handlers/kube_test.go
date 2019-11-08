package handlers

import (
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServices(t *testing.T) {

	appContext := AppContext{}

	mockEnvDao := mockEnvDaoWithLotOfThings(&appContext)
	mockConvention := mockConventionInterface(&appContext)
	mockHelmSvc := mockHelmSvcWithLotOfThings(&appContext)

	appContext.Repositories = Repositories{}
	appContext.Repositories.EnvironmentDAO = mockEnvDao
	appContext.HelmServiceAPI = mockHelmSvc

	req, err := http.NewRequest("GET", "/listServices/123", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/listServices/{id}", appContext.services).Methods("GET")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
	mockConvention.AssertNumberOfCalls(t, "GetKubeConfigFileName", 1)

}
