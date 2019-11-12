package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"

	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAddEnvironments(t *testing.T) {

	var payload model.DataElement
	payload.Data.Group = "Test"
	payload.Data.Name = "Alfa"
	payload.Data.Namespace = "Beta"
	payload.Data.Gateway = "Tetra"
	payload.Data.CACertificate = "XPTOXPTOXPTO"
	payload.Data.Token = "kubeconfig-user-ph111:abbkdd57t68tq2lppg6lwb65sb69282jhsmh3ndwn4vhjtt8blmhh2"

	payS, _ := json.Marshal(payload)

	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	mockObject := &mocks.EnvironmentDAOInterface{}
	mockObject.On("CreateEnvironment", mock.Anything).Return(1, nil)

	appContext.Repositories = Repositories{}
	appContext.Repositories.EnvironmentDAO = mockObject

	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("POST", "/environments", bytes.NewBuffer(payS))
	if err != nil {
		t.Fatal(err)
	}

	roles := []string{"tenkai-admin"}
	principal := model.Principal{Name: "alfa", Email: "beta", Roles: roles}

	pSe, _ := json.Marshal(principal)
	req.Header.Set("principal", string(pSe))

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.addEnvironments)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

}

func TestDeleteEnvironment(t *testing.T) {
	appContext := AppContext{}

	env := mockGetEnv()
	mockEnvDAO := mockGetByID(&appContext)
	mockEnvDAO.On("DeleteEnvironment", env).Return(nil)

	req, err := http.NewRequest("DELETE", "/environments/delete/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, []string{"tenkai-admin"})

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/environments/delete/{id}", appContext.deleteEnvironment).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockEnvDAO.AssertNumberOfCalls(t, "GetByID", 1)
	mockEnvDAO.AssertNumberOfCalls(t, "DeleteEnvironment", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
}

func TestDeleteEnvironment_Unauthorized(t *testing.T) {
	appContext := AppContext{}

	req, err := http.NewRequest("DELETE", "/environments/delete/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, []string{"tenkai-user"})

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/environments/delete/{id}", appContext.deleteEnvironment).Methods("DELETE")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Response is not Ok.")
}

func TestDeleteEnvironment_StringConvError(t *testing.T) {
	appContext := AppContext{}

	req, err := http.NewRequest("DELETE", "/environments/delete/qwert", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, []string{"tenkai-admin"})

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/environments/delete/{id}", appContext.deleteEnvironment).Methods("DELETE")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDeleteEnvironment_GetByIDError(t *testing.T) {
	appContext := AppContext{}

	mockEnvDAO := mockGetByIDError(&appContext)

	req, err := http.NewRequest("DELETE", "/environments/delete/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, []string{"tenkai-admin"})

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/environments/delete/{id}", appContext.deleteEnvironment).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockEnvDAO.AssertNumberOfCalls(t, "GetByID", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestDeleteEnvironment_DeleteEnvironmentError(t *testing.T) {
	appContext := AppContext{}

	env := mockGetEnv()
	mockEnvDAO := mockGetByID(&appContext)
	mockEnvDAO.On("DeleteEnvironment", env).Return(errors.New("some error"))

	req, err := http.NewRequest("DELETE", "/environments/delete/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, []string{"tenkai-admin"})

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/environments/delete/{id}", appContext.deleteEnvironment).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockEnvDAO.AssertNumberOfCalls(t, "GetByID", 1)
	mockEnvDAO.AssertNumberOfCalls(t, "DeleteEnvironment", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestEditEnvironment(t *testing.T) {
	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	mockEnvDAO := mockGetByID(&appContext)
	mockEnvDAO.On("EditEnvironment", mock.Anything).Return(nil)

	var p model.DataElement
	env := mockGetEnv()
	env.Token = "kubeconfig-user-ph111:abbkdd57t68tq2lppg6lwb65sb69282jhsmh3ndwn4vhjtt8blmhh2"
	p.Data = env

	req, err := http.NewRequest("POST", "/environments/edit", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, []string{"tenkai-admin"})

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editEnvironment)
	handler.ServeHTTP(rr, req)

	mockEnvDAO.AssertNumberOfCalls(t, "GetByID", 1)
	mockEnvDAO.AssertNumberOfCalls(t, "EditEnvironment", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")
}
