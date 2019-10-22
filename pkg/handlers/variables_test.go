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

func TestEditVariable(t *testing.T) {

	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	var payload model.DataVariableElement
	payload.Data.Secret = false
	payload.Data.Name = "my_variable"
	payload.Data.Description = "my_description"
	payload.Data.Scope = "my_chart"
	payload.Data.Value = "my_value"
	payload.Data.EnvironmentID = 1

	mockVariableDAO := &mocks.VariableDAOInterface{}
	mockVariableDAO.On("EditVariable", mock.Anything).Return(nil)

	appContext.Repositories = Repositories{}
	appContext.Repositories.VariableDAO = mockVariableDAO

	payloadStr, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "/variables", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	roles := []string{constraints.TenkaiVariablesSave}
	principal := model.Principal{Name: "alfa", Email: "beta@gmail.com", Roles: roles}
	pSe, _ := json.Marshal(principal)
	req.Header.Set("principal", string(pSe))

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
