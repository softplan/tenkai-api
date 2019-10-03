package main

import (
	"github.com/softplan/tenkai-api/configs"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
	dockerTagsCache := make(map[string]time.Time)
	appContext := &appContext{configuration: &config, dockerTagsCache: dockerTagsCache, testMode: true}
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
