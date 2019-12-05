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

func TestNewSolution(t *testing.T) {

	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	mockSolutionDAO := mocks.SolutionDAOInterface{}
	mockSolutionDAO.On("CreateSolution", mock.Anything).Return(1, nil)

	appContext.Repositories = Repositories{}
	appContext.Repositories.SolutionDAO = &mockSolutionDAO

	var payload model.Solution
	payload.Name = "alfa"
	payload.Team = "teamx"
	payloadStr, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "/solutions", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newSolution)
	handler.ServeHTTP(rr, req)

	mockSolutionDAO.AssertNumberOfCalls(t, "CreateSolution", 1)
	assert.Equal(t, http.StatusCreated, rr.Code, "Response is not Ok.")

}

func TestEditSolution(t *testing.T) {

	var payload model.Solution
	payload.Name = "alfa"
	payload.Team = "teamx"

	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	mockSolutionDAO := mocks.SolutionDAOInterface{}
	mockSolutionDAO.On("EditSolution", mock.Anything).Return(nil)

	appContext.Repositories = Repositories{}
	appContext.Repositories.SolutionDAO = &mockSolutionDAO

	payloadStr, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", "/solutions/edit", bytes.NewBuffer(payloadStr))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editSolution)
	handler.ServeHTTP(rr, req)

	mockSolutionDAO.AssertNumberOfCalls(t, "EditSolution", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")

}

func TestDeleteSolution(t *testing.T) {

	appContext := AppContext{}
	appContext.K8sConfigPath = "/tmp/"

	mockSolutionDAO := mocks.SolutionDAOInterface{}
	mockSolutionDAO.On("DeleteSolution", mock.Anything).Return(nil)

	appContext.Repositories = Repositories{}
	appContext.Repositories.SolutionDAO = &mockSolutionDAO

	req, err := http.NewRequest("DELETE", "/solutions/1", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/solutions/{id}", appContext.deleteSolution).Methods("DELETE")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
	mockSolutionDAO.AssertNumberOfCalls(t, "DeleteSolution", 1)

}

func TestListSolution(t *testing.T) {

	appContext := AppContext{}

	mockSolutionDAO := mocks.SolutionDAOInterface{}
	solutions := make([]model.Solution, 0)

	mockSolutionDAO.On("ListSolutions").Return(solutions, nil)

	appContext.Repositories = Repositories{}
	appContext.Repositories.SolutionDAO = &mockSolutionDAO

	req, err := http.NewRequest("GET", "/solutions", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/solutions", appContext.listSolution).Methods("GET")
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Response is not Ok.")
	mockSolutionDAO.AssertNumberOfCalls(t, "ListSolutions", 1)

}
