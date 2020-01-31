package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/configs"
	"github.com/softplan/tenkai-api/pkg/constraints"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEditVariable(t *testing.T) {
	config := configs.Configuration{
		App: configs.App{
			Passkey: "qwert",
		},
	}

	appContext := &AppContext{Configuration: &config}
	appContext.K8sConfigPath = "/tmp/"

	mockVariableDAO := &mocks.VariableDAOInterface{}
	mockVariableDAO.On("EditVariable", mock.Anything).Return(nil)

	appContext.Repositories = Repositories{}
	appContext.Repositories.VariableDAO = mockVariableDAO

	req, err := http.NewRequest("POST", "/variables", payload(getDataVariableElement(true)))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-variables-save")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editVariable)
	handler.ServeHTTP(rr, req)

	mockVariableDAO.AssertNumberOfCalls(t, "EditVariable", 1)
	assert.Equal(t, http.StatusCreated, rr.Code, "Response is not Ok.")
}

func TestDeleteVariable(t *testing.T) {

	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	mockVariableDAO := &mocks.VariableDAOInterface{}
	mockVariableDAO.On("DeleteVariable", mock.Anything).Return(nil)

	appContext.Repositories = Repositories{}
	appContext.Repositories.VariableDAO = mockVariableDAO

	req, err := http.NewRequest("DELETE", "/variables/delete/1", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	roles := []string{constraints.TenkaiVariablesDelete}
	principal := model.Principal{Name: "alfa", Email: "beta@gmail.com", Roles: roles}
	pSe, _ := json.Marshal(principal)
	req.Header.Set("principal", string(pSe))

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/variables/delete/{id}", appContext.deleteVariable).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockVariableDAO.AssertNumberOfCalls(t, "DeleteVariable", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")

}

func TestDeleteVariable_Unauthorized(t *testing.T) {
	appContext := AppContext{}

	req, err := http.NewRequest("DELETE", "/variables/delete/1", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-user")

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/variables/delete/{id}", appContext.deleteVariable).Methods("DELETE")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Response should 401.")
}

func TestDeleteVariable_DeleteVariableError(t *testing.T) {
	appContext := AppContext{}

	mockVariableDAO := &mocks.VariableDAOInterface{}
	mockVariableDAO.On("DeleteVariable", mock.Anything).Return(errors.New("some error"))
	appContext.Repositories.VariableDAO = mockVariableDAO

	req, err := http.NewRequest("DELETE", "/variables/delete/1", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-variables-delete")

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/variables/delete/{id}", appContext.deleteVariable).Methods("DELETE")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestEditVariable_Unauthorized(t *testing.T) {
	appContext := AppContext{}

	req, err := http.NewRequest("POST", "/variables", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-user")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editVariable)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Response should be 401.")
}

func TestEditVariable_UnmarshalPayloadError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadErrorWithPrincipal(t, "/variables", appContext.editVariable, "tenkai-variables-save")
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestEditVariable_EditVariableError(t *testing.T) {
	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	mockVariableDAO := mockEditVariableError(&appContext)

	req, err := http.NewRequest("POST", "/variables", payload(getDataVariableElement(false)))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-variables-save")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editVariable)
	handler.ServeHTTP(rr, req)

	mockVariableDAO.AssertNumberOfCalls(t, "EditVariable", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestGetVariables(t *testing.T) {
	appContext := AppContext{}
	mockVariableDAO := mockGetAllVariablesByEnvironment(&appContext)

	req, err := http.NewRequest("GET", "/variables/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-user")
	mockEnvDao := mockGetAllEnvironments(&appContext)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/variables/{envId}", appContext.getVariables).Methods("GET")
	r.ServeHTTP(rr, req)

	mockVariableDAO.AssertNumberOfCalls(t, "GetAllVariablesByEnvironment", 1)
	mockEnvDao.AssertNumberOfCalls(t, "GetAllEnvironments", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")

	response := string(rr.Body.Bytes())
	assert.Contains(t, response, `{"Variables":[{"ID":0,`)
	assert.Contains(t, response, `"scope":"global","chartVersion":"","name":"username","value":"user",`)
	assert.Contains(t, response, `"secret":false,"description":"Login username.","environmentId":999},{"ID":0,`)
	assert.Contains(t, response, `"scope":"bar","chartVersion":"","name":"password","value":"password","secret":false,`)
	assert.Contains(t, response, `"description":"Login password.","environmentId":999}]}`)
}

func TestGetVariables_Unauthorized(t *testing.T) {
	appContext := AppContext{}

	req, err := http.NewRequest("GET", "/variables/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockEnvDao := mockGetAllEnvironmentsError(&appContext)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/variables/{envId}", appContext.getVariables).Methods("GET")
	r.ServeHTTP(rr, req)

	mockEnvDao.AssertNumberOfCalls(t, "GetAllEnvironments", 1)
	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Response should be 401.")
}

func TestGetVariables_GetAllVarByEnvError(t *testing.T) {
	appContext := AppContext{}
	mockVariableDAO := mockGetAllVariablesByEnvironmentError(&appContext)

	req, err := http.NewRequest("GET", "/variables/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-user")
	mockEnvDao := mockGetAllEnvironments(&appContext)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/variables/{envId}", appContext.getVariables).Methods("GET")
	r.ServeHTTP(rr, req)

	mockVariableDAO.AssertNumberOfCalls(t, "GetAllVariablesByEnvironment", 1)
	mockEnvDao.AssertNumberOfCalls(t, "GetAllEnvironments", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestCopyVariableValue(t *testing.T) {
	config := configs.Configuration{
		App: configs.App{
			Passkey: "qwert",
		},
	}

	appContext := &AppContext{Configuration: &config}
	appContext.K8sConfigPath = "/tmp/"

	mockVariableDAO := &mocks.VariableDAOInterface{}
	mockVariableDAO.On("EditVariable", mock.Anything).Return(nil)

	var srcVar model.Variable
	srcVar.ID = 999
	srcVar.Scope = "foo"
	srcVar.Name = "foo"
	srcVar.Value = "foo"
	srcVar.Secret = false
	srcVar.Description = "foo"
	srcVar.EnvironmentID = 999

	var tarVar model.Variable
	tarVar.ID = 888
	tarVar.Scope = "foo"
	tarVar.Name = "foo"
	tarVar.Value = "bar"
	tarVar.Secret = false
	tarVar.Description = "foo"
	tarVar.EnvironmentID = 888

	mockVariableDAO.On("GetByID", srcVar.ID).Return(&srcVar, nil)
	mockVariableDAO.On("GetByID", tarVar.ID).Return(&tarVar, nil)

	appContext.Repositories = Repositories{}
	appContext.Repositories.VariableDAO = mockVariableDAO

	var p model.CopyVariableValue
	p.SrcVarID = 999
	p.TarVarID = 888
	p.NewValue = "foo"

	req, err := http.NewRequest("POST", "/variables/copy-value", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	mockPrincipal(req, "tenkai-variables-save")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.copyVariableValue)
	handler.ServeHTTP(rr, req)

	mockVariableDAO.AssertNumberOfCalls(t, "EditVariable", 1)
	mockVariableDAO.AssertNumberOfCalls(t, "GetByID", 2)
	assert.Equal(t, http.StatusCreated, rr.Code, "Response is not Ok.")
}

func getDataVariableElement(secret bool) model.DataVariableElement {
	var payload model.DataVariableElement
	payload.Data.Secret = secret
	payload.Data.Name = "my_variable"
	payload.Data.Description = "my_description"
	payload.Data.Scope = "my_chart"
	payload.Data.Value = "my_value"
	payload.Data.EnvironmentID = 1
	return payload
}
