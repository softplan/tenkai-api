package handlers

import (
	"bytes"
	"encoding/json"
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

func getPayloadNewVariables() []byte {
	data := map[string]interface{}{
		"chartName":     "chartxpto",
		"chartVersion":  "0.1.0",
		"repo":          "repoxpto",
		"environmentId": 999,
	}
	payload, _ := json.Marshal(data)
	return payload
}

func getReturnAllVariablesByEnvironmentAndScope() []model.Variable {
	variables := make([]model.Variable, 0)
	variables = append(variables, model.Variable{
		Scope:         "repoxpto/chartxpto",
		Name:          "myvar",
		Value:         "myvalue",
		Secret:        false,
		Description:   "xpto",
		EnvironmentID: 999,
	})
	return variables
}

func getAppVars() map[string]interface{} {
	return map[string]interface{}{
		"app": map[string]interface{}{
			"myvar": "myvalue",
		},
	}
}

func TestListVariablesNewOk(t *testing.T) {
	appContext := AppContext{}

	mockVariableDAO := mocks.VariableDAOInterface{}
	mockVariableDAO.On("GetAllVariablesByEnvironmentAndScope", mock.Anything, mock.Anything).Return(getReturnAllVariablesByEnvironmentAndScope(), nil)

	appVars := getAppVars()
	data, _ := json.Marshal(appVars)
	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	mockHelmSvc.On("GetTemplate", mock.Anything, mock.Anything, mock.Anything, "values").Return(data, nil)

	mockGetAllEnvironmentsPrincipal(&appContext, "")

	appContext.Repositories.VariableDAO = &mockVariableDAO
	appContext.HelmServiceAPI = mockHelmSvc

	params := "?repo=repoxpto&chartName=chartxpto&chartVersion=0.1.0&environmentId=999" 

	data, _ = json.Marshal(payload)
	req, err := http.NewRequest("POST", "/listVariablesNew" + params, nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/listVariablesNew", appContext.listVariablesNew).Methods("POST")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Response should be 200")
}

func TestValidateNewVariablesBeforeInstallOK(t *testing.T) {
	appContext := AppContext{}

	mockVariableDAO := mocks.VariableDAOInterface{}
	mockVariableDAO.On("GetAllVariablesByEnvironmentsAndScopes", mock.Anything, mock.Anything).Return(getReturnAllVariablesByEnvironmentAndScope(), nil)

	appVars := getAppVars()
	data, _ := json.Marshal(appVars)
	mockHelmSvc := &mockSvc.HelmServiceInterface{}
	mockHelmSvc.On("GetTemplate", mock.Anything, mock.Anything, mock.Anything, "values").Return(data, nil)

	mockEnv := mocks.EnvironmentDAOInterface{}
	mockEnv.On("GetByID", mock.Anything).Return(nil, nil)

	appContext.Repositories.VariableDAO = &mockVariableDAO
	appContext.HelmServiceAPI = mockHelmSvc
	appContext.Repositories.EnvironmentDAO = &mockEnv
	payload := map[string]interface{}{
		"charts": []model.Chart{
			{Repo: "repoxpto", Name: "chartxpto", Version: "0.1.0"},
			{Repo: "repoqwert", Name: "chartqwert", Version: "0.1.0"},
		},
		"environments": []int{999, 888},
	}

	data, _ = json.Marshal(payload)
	req, err := http.NewRequest("POST", "/validateCharts", bytes.NewBuffer(data))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/validateCharts", appContext.validateNewVariablesBeforeInstall).Methods("POST")
	r.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be 200")
}
