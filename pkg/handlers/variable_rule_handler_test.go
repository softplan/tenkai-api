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

func TestNewValueRule(t *testing.T) {
	appContext := AppContext{}

	p := mockValueRule()

	mockValueRule := &mockRepo.ValueRuleDAOInterface{}
	mockValueRule.On("CreateValueRule", p).Return(1, nil)

	appContext.Repositories.ValueRuleDAO = mockValueRule

	req, err := http.NewRequest("POST", "/valuerules", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newValueRule)
	handler.ServeHTTP(rr, req)

	mockValueRule.AssertNumberOfCalls(t, "CreateValueRule", 1)
	assert.Equal(t, http.StatusCreated, rr.Code, "Response should be Created")
}

func TestNewValueRule_UnmarshalPaylError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/valuerules", appContext.newValueRule)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestNewValueRule_Error(t *testing.T) {
	appContext := AppContext{}

	p := mockValueRule()

	mockValueRule := &mockRepo.ValueRuleDAOInterface{}
	mockValueRule.On("CreateValueRule", p).Return(1, errors.New("some error"))

	appContext.Repositories.ValueRuleDAO = mockValueRule

	req, err := http.NewRequest("POST", "/valuerules", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newValueRule)
	handler.ServeHTTP(rr, req)

	mockValueRule.AssertNumberOfCalls(t, "CreateValueRule", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500")
}

func TestEditValueRule(t *testing.T) {
	appContext := AppContext{}

	p := mockValueRule()

	mockValueRule := &mockRepo.ValueRuleDAOInterface{}
	mockValueRule.On("EditValueRule", p).Return(nil)

	appContext.Repositories.ValueRuleDAO = mockValueRule

	req, err := http.NewRequest("POST", "/valuerules/edit", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editValueRule)
	handler.ServeHTTP(rr, req)

	mockValueRule.AssertNumberOfCalls(t, "EditValueRule", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be ok")
}

func TestEditValueRule_UnmarshalPaylError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/valuerules", appContext.editValueRule)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestEditValueRule_Error(t *testing.T) {
	appContext := AppContext{}

	p := mockValueRule()

	mockValueRule := &mockRepo.ValueRuleDAOInterface{}
	mockValueRule.On("EditValueRule", p).Return(errors.New("some error"))

	appContext.Repositories.ValueRuleDAO = mockValueRule

	req, err := http.NewRequest("POST", "/valuerules/edit", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editValueRule)
	handler.ServeHTTP(rr, req)

	mockValueRule.AssertNumberOfCalls(t, "EditValueRule", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500")
}

func TestDeleteValueRule(t *testing.T) {
	appContext := AppContext{}

	mockValueRule := &mockRepo.ValueRuleDAOInterface{}
	mockValueRule.On("DeleteValueRule", 999).Return(nil)

	appContext.Repositories.ValueRuleDAO = mockValueRule

	req, err := http.NewRequest("DELETE", "/valuerules/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/valuerules/{id}", appContext.deleteValueRule).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockValueRule.AssertNumberOfCalls(t, "DeleteValueRule", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")
}

func TestDeleteValueRule_Error(t *testing.T) {
	appContext := AppContext{}

	mockValueRule := &mockRepo.ValueRuleDAOInterface{}
	mockValueRule.On("DeleteValueRule", 999).Return(errors.New("some error"))

	appContext.Repositories.ValueRuleDAO = mockValueRule

	req, err := http.NewRequest("DELETE", "/valuerules/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/valuerules/{id}", appContext.deleteValueRule).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockValueRule.AssertNumberOfCalls(t, "DeleteValueRule", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestListValueRule(t *testing.T) {
	appContext := AppContext{}

	mockValueRule := &mockRepo.ValueRuleDAOInterface{}
	result := &model.ValueRuleReponse{}
	result.List = append(result.List, mockValueRuleWithID())
	mockValueRule.On("ListValueRules", 999).Return(result.List, nil)

	appContext.Repositories.ValueRuleDAO = mockValueRule

	req, err := http.NewRequest("GET", "/valuerules?variableRuleId=999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listValueRules)
	handler.ServeHTTP(rr, req)

	mockValueRule.AssertNumberOfCalls(t, "ListValueRules", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be ok")

	response := string(rr.Body.Bytes())
	assert.Contains(t, response, `{"list":[{"ID":888,`)
	assert.Contains(t, response, `"type":"StartsWith",`)
	assert.Contains(t, response, `"value":"http","VariableRuleID":999}]}`)
}

func TestListValueRule_ParseError(t *testing.T) {
	appContext := AppContext{}

	req, err := http.NewRequest("GET", "/valuerules?xxxxxx=999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listValueRules)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500")
}

func TestListValueRule_Error(t *testing.T) {
	appContext := AppContext{}

	mockValueRule := &mockRepo.ValueRuleDAOInterface{}
	mockValueRule.On("ListValueRules", 999).Return(nil, errors.New("some error"))
	appContext.Repositories.ValueRuleDAO = mockValueRule

	req, err := http.NewRequest("GET", "/valuerules?variableRuleId=999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listValueRules)
	handler.ServeHTTP(rr, req)

	mockValueRule.AssertNumberOfCalls(t, "ListValueRules", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500")
}
