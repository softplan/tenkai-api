package main

import (
	"github.com/softplan/tenkai-api/configs"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func checkTestingFatalError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func GetAppContext() *appContext {
	config := configs.Configuration{
		Server: configs.Server{
			Port: "1010",
		},
		App: configs.App{
			Dbms: configs.Dbms{
				URI: "",
			},
		},
	}

	appContext := &appContext{configuration: &config}
	appContext.dockerTagsCache = sync.Map{}
	return appContext
}

func TestRoot(t *testing.T) {
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	checkTestingFatalError(t, err)
	appContext := GetAppContext()
	handler := http.HandlerFunc(appContext.rootHandler)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
