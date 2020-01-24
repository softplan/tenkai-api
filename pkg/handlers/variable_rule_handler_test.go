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
)

func getVariableRule() model.VariableRule {
	var item model.VariableRule
	item.Name = "urlapi.*"
	return item
}

func TestNewVariableRule(t *testing.T) {
	appContext := AppContext{}

	p := mockVariableRule()

	mockVarRule := &mockRepo.VariableRuleDAOInterface{}
	mockVarRule.On("CreateVariableRule", p).Return(1, nil)

	appContext.Repositories.VariableRuleDAO = mockVarRule

	req, err := http.NewRequest("POST", "/variablerules", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newVariableRule)
	handler.ServeHTTP(rr, req)

	mockVarRule.AssertNumberOfCalls(t, "CreateVariableRule", 1)
	assert.Equal(t, http.StatusCreated, rr.Code, "Response should be Created")
}

func TestNewVariableRule_UnmarshalPaylError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/variablerules", appContext.newVariableRule)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestNewVariableRule_Error(t *testing.T) {
	appContext := AppContext{}

	p := mockVariableRule()

	mockVarRule := &mockRepo.VariableRuleDAOInterface{}
	mockVarRule.On("CreateVariableRule", p).Return(1, errors.New("some error"))

	appContext.Repositories.VariableRuleDAO = mockVarRule

	req, err := http.NewRequest("POST", "/variablerules", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newVariableRule)
	handler.ServeHTTP(rr, req)

	mockVarRule.AssertNumberOfCalls(t, "CreateVariableRule", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500")
}

func TestEditVariableRule(t *testing.T) {
	appContext := AppContext{}

	p := mockVariableRule()

	mockVarRule := &mockRepo.VariableRuleDAOInterface{}
	mockVarRule.On("EditVariableRule", p).Return(nil)

	appContext.Repositories.VariableRuleDAO = mockVarRule

	req, err := http.NewRequest("POST", "/variablerules/edit", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editVariableRule)
	handler.ServeHTTP(rr, req)

	mockVarRule.AssertNumberOfCalls(t, "EditVariableRule", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be ok")
}

func TestEditVariableRule_UnmarshalPaylError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/variablerules/edit", appContext.editVariableRule)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestEditVariableRule_Error(t *testing.T) {
	appContext := AppContext{}

	p := mockVariableRuleWithID()

	mockVarRule := &mockRepo.VariableRuleDAOInterface{}
	mockVarRule.On("EditVariableRule", p).Return(errors.New("some error"))

	appContext.Repositories.VariableRuleDAO = mockVarRule

	req, err := http.NewRequest("POST", "/variablerules/edit", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editVariableRule)
	handler.ServeHTTP(rr, req)

	mockVarRule.AssertNumberOfCalls(t, "EditVariableRule", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500")
}

func TestDeleteVariableRule(t *testing.T) {
	appContext := AppContext{}

	mockVarRule := &mockRepo.VariableRuleDAOInterface{}
	mockVarRule.On("DeleteVariableRule", 999).Return(nil)

	appContext.Repositories.VariableRuleDAO = mockVarRule

	req, err := http.NewRequest("DELETE", "/variablerules/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/variablerules/{id}", appContext.deleteVariableRule).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockVarRule.AssertNumberOfCalls(t, "DeleteVariableRule", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")
}

func TestDeleteVariableRule_Error(t *testing.T) {
	appContext := AppContext{}

	mockVarRule := &mockRepo.VariableRuleDAOInterface{}
	mockVarRule.On("DeleteVariableRule", 999).Return(errors.New("some error"))

	appContext.Repositories.VariableRuleDAO = mockVarRule

	req, err := http.NewRequest("DELETE", "/variablerules/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/variablerules/{id}", appContext.deleteVariableRule).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockVarRule.AssertNumberOfCalls(t, "DeleteVariableRule", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestListVariableRule(t *testing.T) {
	appContext := AppContext{}

	mockVarRule := &mockRepo.VariableRuleDAOInterface{}
	result := &model.VariableRuleReponse{}
	result.List = append(result.List, mockVariableRuleWithID())
	mockVarRule.On("ListVariableRules").Return(result.List, nil)

	appContext.Repositories.VariableRuleDAO = mockVarRule

	req, err := http.NewRequest("GET", "/variablerules", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listVariableRules)
	handler.ServeHTTP(rr, req)

	mockVarRule.AssertNumberOfCalls(t, "ListVariableRules", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be ok")

	response := string(rr.Body.Bytes())
	assert.Contains(t, response, `{"list":[{"ID":999,`)
	assert.Contains(t, response, `"name":"urlapi.*","ValueRules":[{"ID":888,`)
	assert.Contains(t, response, `"type":"StartsWith","value":"http","VariableRuleID":999}]}]}`)
}

func TestListVariableRule_Error(t *testing.T) {
	appContext := AppContext{}

	mockVarRule := &mockRepo.VariableRuleDAOInterface{}
	mockVarRule.On("ListVariableRules").Return(nil, errors.New("some error"))

	appContext.Repositories.VariableRuleDAO = mockVarRule

	req, err := http.NewRequest("GET", "/variablerules", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listVariableRules)
	handler.ServeHTTP(rr, req)

	mockVarRule.AssertNumberOfCalls(t, "ListVariableRules", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500")
}
