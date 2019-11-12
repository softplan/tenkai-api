package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewSolutionChart(t *testing.T) {

	var payload model.SolutionChart
	payload.ChartName = "saj6/beta3"
	payload.SolutionID = 3
	payloadStr, _ := json.Marshal(payload)

	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	mockSolutionChart := mocks.SolutionChartDAOInterface{}
	mockSolutionChart.On("CreateSolutionChart", mock.Anything).Return(nil)

	appContext.Repositories = Repositories{}
	appContext.Repositories.SolutionChartDAO = &mockSolutionChart

	req, err := http.NewRequest("POST", "/newSolutionChart", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newSolutionChart)
	handler.ServeHTTP(rr, req)

	mockSolutionChart.AssertNumberOfCalls(t, "CreateSolutionChart", 1)
	assert.Equal(t, http.StatusCreated, rr.Code, "Response is not Ok.")

}

func TestDeleteSolutionChart(t *testing.T) {

	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	mockSolutionChart := mocks.SolutionChartDAOInterface{}
	mockSolutionChart.On("DeleteSolutionChart", mock.Anything).Return(nil)

	appContext.Repositories = Repositories{}
	appContext.Repositories.SolutionChartDAO = &mockSolutionChart

	req, err := http.NewRequest("DELETE", "/solutionCharts/1", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/solutionCharts/{id}", appContext.deleteSolutionChart).Methods("DELETE")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
	mockSolutionChart.AssertNumberOfCalls(t, "DeleteSolutionChart", 1)

}

func TestListSolutionCharts(t *testing.T) {

	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	result := make([]model.SolutionChart, 0)

	mockSolutionChart := mocks.SolutionChartDAOInterface{}
	mockSolutionChart.On("ListSolutionChart", mock.Anything).Return(result, nil)

	appContext.Repositories = Repositories{}
	appContext.Repositories.SolutionChartDAO = &mockSolutionChart

	req, err := http.NewRequest("GET", "/solutionCharts?solutionId=1", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/solutionCharts", appContext.listSolutionCharts).Methods("GET")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
	mockSolutionChart.AssertNumberOfCalls(t, "ListSolutionChart", 1)

}
