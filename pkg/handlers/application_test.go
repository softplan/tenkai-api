package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/configs"
	"github.com/softplan/tenkai-api/pkg/util"
	"github.com/stretchr/testify/assert"
)

func checkTestingFatalError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func GetAppContext() *AppContext {
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

	appContext := &AppContext{Configuration: &config}
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

func TestDefineRoutes(t *testing.T) {
	r := mux.NewRouter()
	defineRotes(r, &AppContext{})
}

func TestCommonHandler(t *testing.T) {
	appContext := GetAppContext()
	r := mux.NewRouter()
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	assert.NotNil(t, req)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICI2OUhmUHpYNDV0WGtGbzFPSDVtajNHeEtxMTBxTjdWeHNSaDZFSFBuRTZNIn0.eyJqdGkiOiIzNzA0YTg0Ni1kY2U5LTRiNGMtYWJkYy02YjhmMGQ1MTVhZjAiLCJleHAiOjE1NjQ3MTA5MjEsIm5iZiI6MCwiaWF0IjoxNTY0NzEwNjIxLCJpc3MiOiJodHRwOi8va2V5Y2xvYWstdG9vbHMuc2FqNi5zb2Z0cGxhbi5jb20uYnIvYXV0aC9yZWFsbXMvdGVua2FpIiwiYXVkIjpbInRlbmthaSIsImFjY291bnQiXSwic3ViIjoiMTg5OTNjNTktMGEwNi00N2NjLTljYmYtNmQwODU1MWZiMTI2IiwidHlwIjoiQmVhcmVyIiwiYXpwIjoidGVua2FpIiwibm9uY2UiOiJhYzM5MTJjZi03YzQxLTRiOGQtOTUwMS0zOTFjZmVlMzk1M2IiLCJhdXRoX3RpbWUiOjE1NjQ3MTA2MDAsInNlc3Npb25fc3RhdGUiOiJkZTQwNzRlMi00NjNhLTQ5NzktODljMy0xZDQ5NGQ4MWM2MWMiLCJhY3IiOiIwIiwiYWxsb3dlZC1vcmlnaW5zIjpbIioiXSwicmVhbG1fYWNjZXNzIjp7InJvbGVzIjpbIm9mZmxpbmVfYWNjZXNzIiwidGVua2FpLXVzZXIiLCJ1bWFfYXV0aG9yaXphdGlvbiJdfSwicmVzb3VyY2VfYWNjZXNzIjp7ImFjY291bnQiOnsicm9sZXMiOlsibWFuYWdlLWFjY291bnQiLCJtYW5hZ2UtYWNjb3VudC1saW5rcyIsInZpZXctcHJvZmlsZSJdfX0sInNjb3BlIjoib3BlbmlkIGVtYWlsIHByb2ZpbGUiLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwibmFtZSI6IkRlbm55IFZyaWVzbWFuIiwicHJlZmVycmVkX3VzZXJuYW1lIjoiZGVubnkiLCJnaXZlbl9uYW1lIjoiRGVubnkiLCJmYW1pbHlfbmFtZSI6IlZyaWVzbWFuIiwiZW1haWwiOiJkZW5ueUBzb2Z0cGxhbi5jb20uYnIifQ.OH2cQitgmvetEScv7pGo4X5jOtofzidYIbqUweQ6BgfEuYYWZJnmJLAqnv0TSybm42uWFoviLCp6YFhI9Q6cFF2uLuuFzqBb49ZT3oEyEBZCqFMVqLL82NS_gt0wr6ntPBN59XCDUaGQnTai1tI8td-rgMR5Dpy_oSK51G6EEvd3w5pDPRUdRgffvWH-MuE_m7M84QagABuIvwo3HmQe-VMFNRzCszkCp-FwHtS2ChxXBk8hKKI0SfVKRIDYdQF3tetUTwQ40TuNV3HJG9Xuu4jxI_1sSlYFl5YZTxojmtX9vPEAMcDyYVp0zkOgYiur9XYz5XdL89uSSM4H28mNqA")
	r.HandleFunc("/", appContext.rootHandler).Methods("GET")
	commonHandler(r).ServeHTTP(rr, req)
	principal := util.GetPrincipal(req)
	assert.NotNil(t, principal)
	assert.Equal(t, "denny@softplan.com.br", principal.Email)
	assert.Equal(t, 3, len(principal.Roles))
}
