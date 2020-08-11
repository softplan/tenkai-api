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

func mockWebHook() model.WebHook {
	var item model.WebHook
	item.Name = "Product Deploy"
	item.Type = "HOOK_DEPLOY_PRODUCT"
	item.URL = "http://example.com"
	item.EnvironmentID = 999
	return item
}

func mockWebHookWithID() model.WebHook {
	var item model.WebHook
	item.ID = 999
	item.Name = "Product Deploy"
	item.Type = "HOOK_DEPLOY_PRODUCT"
	item.URL = "http://example.com"
	item.EnvironmentID = 999
	return item
}

func TestNewWebHook(t *testing.T) {
	appContext := AppContext{}

	p := mockWebHook()

	mockWebHook := &mockRepo.WebHookDAOInterface{}
	mockWebHook.On("CreateWebHook", p).Return(1, nil)

	appContext.Repositories.WebHookDAO = mockWebHook

	req, err := http.NewRequest("POST", "/webhooks", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newWebHook)
	handler.ServeHTTP(rr, req)

	mockWebHook.AssertNumberOfCalls(t, "CreateWebHook", 1)
	assert.Equal(t, http.StatusCreated, rr.Code, "Response should be Created")
}

func TestNewWebHook_UnmarshalPaylError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/webhooks", appContext.newWebHook)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestNewWebHook_Error(t *testing.T) {
	appContext := AppContext{}

	p := mockWebHook()

	mockWebHook := &mockRepo.WebHookDAOInterface{}
	mockWebHook.On("CreateWebHook", p).Return(1, errors.New("some error"))

	appContext.Repositories.WebHookDAO = mockWebHook

	req, err := http.NewRequest("POST", "/webhooks", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.newWebHook)
	handler.ServeHTTP(rr, req)

	mockWebHook.AssertNumberOfCalls(t, "CreateWebHook", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500")
}

func TestEditWebHook(t *testing.T) {
	appContext := AppContext{}

	p := mockWebHook()

	mockWebHook := &mockRepo.WebHookDAOInterface{}
	mockWebHook.On("EditWebHook", p).Return(nil)

	appContext.Repositories.WebHookDAO = mockWebHook

	req, err := http.NewRequest("POST", "/webhooks/edit", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editWebHook)
	handler.ServeHTTP(rr, req)

	mockWebHook.AssertNumberOfCalls(t, "EditWebHook", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be ok")
}

func TestEditWebHook_UnmarshalPaylError(t *testing.T) {
	appContext := AppContext{}
	rr := testUnmarshalPayloadError(t, "/webhooks/edit", appContext.editWebHook)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestEditWebHook_Error(t *testing.T) {
	appContext := AppContext{}

	p := mockWebHookWithID()

	mockWebHook := &mockRepo.WebHookDAOInterface{}
	mockWebHook.On("EditWebHook", p).Return(errors.New("some error"))

	appContext.Repositories.WebHookDAO = mockWebHook

	req, err := http.NewRequest("POST", "/webhooks/edit", payload(p))
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.editWebHook)
	handler.ServeHTTP(rr, req)

	mockWebHook.AssertNumberOfCalls(t, "EditWebHook", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500")
}

func TestDeleteWebHook(t *testing.T) {
	appContext := AppContext{}

	mockWebHook := &mockRepo.WebHookDAOInterface{}
	mockWebHook.On("DeleteWebHook", 999).Return(nil)

	appContext.Repositories.WebHookDAO = mockWebHook

	req, err := http.NewRequest("DELETE", "/webhooks/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/webhooks/{id}", appContext.deleteWebHook).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockWebHook.AssertNumberOfCalls(t, "DeleteWebHook", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be Ok.")
}

func TestDeleteWebHook_Error(t *testing.T) {
	appContext := AppContext{}

	mockWebHook := &mockRepo.WebHookDAOInterface{}
	mockWebHook.On("DeleteWebHook", 999).Return(errors.New("some error"))

	appContext.Repositories.WebHookDAO = mockWebHook

	req, err := http.NewRequest("DELETE", "/webhooks/999", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/webhooks/{id}", appContext.deleteWebHook).Methods("DELETE")
	r.ServeHTTP(rr, req)

	mockWebHook.AssertNumberOfCalls(t, "DeleteWebHook", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500.")
}

func TestListWebHook(t *testing.T) {
	appContext := AppContext{}

	mockWebHook := &mockRepo.WebHookDAOInterface{}
	result := &model.WebHookReponse{}
	result.List = append(result.List, mockWebHookWithID())
	mockWebHook.On("ListWebHooks").Return(result.List, nil)

	appContext.Repositories.WebHookDAO = mockWebHook

	req, err := http.NewRequest("GET", "/webhooks", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listWebHooks)
	handler.ServeHTTP(rr, req)

	mockWebHook.AssertNumberOfCalls(t, "ListWebHooks", 1)
	assert.Equal(t, http.StatusOK, rr.Code, "Response should be ok")

	response := string(rr.Body.Bytes())
	assert.Contains(t, response, `{"list":[{"ID":999,`)
	assert.Contains(t, response, `"name":"Product Deploy",`)
	assert.Contains(t, response, `"type":"HOOK_DEPLOY_PRODUCT"`)
	assert.Contains(t, response, `"url":"http://example.com"`)
	assert.Contains(t, response, `"environmentId":999}]}`)
}

func TestListWebHook_Error(t *testing.T) {
	appContext := AppContext{}

	mockWebHook := &mockRepo.WebHookDAOInterface{}
	mockWebHook.On("ListWebHooks").Return(nil, errors.New("some error"))

	appContext.Repositories.WebHookDAO = mockWebHook

	req, err := http.NewRequest("GET", "/webhooks", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listWebHooks)
	handler.ServeHTTP(rr, req)

	mockWebHook.AssertNumberOfCalls(t, "ListWebHooks", 1)
	assert.Equal(t, http.StatusInternalServerError, rr.Code, "Response should be 500")
}
