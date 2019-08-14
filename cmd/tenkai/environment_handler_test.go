package main

import (
	"encoding/json"
	"github.com/softplan/tenkai-api/dbms/model"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetEnvironmentsAccessDenied(t *testing.T) {
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/environments", nil)
	checkTestingFatalError(t, err)
	appContext := GetAppContext()
	appContext.database.MockConnect()
	handler := http.HandlerFunc(appContext.getEnvironments)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}

func TestGetEnvironmentsNotFound(t *testing.T) {
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/environments", nil)

	data, _ := json.Marshal(&model.Principal{Email: "test@beta.com.br"})
	req.Header.Set("principal", string(data))

	checkTestingFatalError(t, err)
	appContext := GetAppContext()
	appContext.database.MockConnect()
	handler := http.HandlerFunc(appContext.getEnvironments)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}
