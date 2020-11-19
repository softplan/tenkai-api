package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getRequestDeploymentAppContext() *AppContext {
	appContext := AppContext{}
	requestDeploymentMock := mocks.RequestDeploymentDAOInterface{}

	requestDeployment := []model.RequestDeployments{}

	requestDeploymentMock.On(
		"ListRequestDeployments",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		-1,
		1,
		100,
	).Return(requestDeployment, nil)

	requestDeploymentMock.On(
		"CountRequestDeployments",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(int64(1), nil)

	appContext.Repositories.RequestDeploymentDAO = &requestDeploymentMock

	return &appContext
}


func TestListRequestDeploymentsWithRightParams(test *testing.T) {

	req, err := http.NewRequest(
		"GET",
		"/requestDeployments?start_date=2020-01-01&end_date=2020-01-01",
		nil,
	)
	if err != nil {
		test.Fatal(err)
	}

	appContext := getRequestDeploymentAppContext()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listRequestDeployments)
	handler.ServeHTTP(rr, req)

	assert.Equal(test, http.StatusOK, rr.Result().StatusCode)
}

func TestListRequestDeploymentsWithoutParams(test *testing.T) {

	req, err := http.NewRequest(
		"GET",
		"/requestDeployments",
		nil,
	)
	if err != nil {
		test.Fatal(err)
	}

	appContext := getRequestDeploymentAppContext()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listRequestDeployments)
	handler.ServeHTTP(rr, req)

	assert.Equal(test, http.StatusBadRequest, rr.Result().StatusCode)
}

func TestListRequestDeploymentsWithWrongStartDate(test *testing.T) {

	req, err := http.NewRequest(
		"GET",
		"/requestDeployments?start_date=2020-01&end_date=2020-01-01",
		nil,
	)
	if err != nil {
		test.Fatal(err)
	}

	appContext := getRequestDeploymentAppContext()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listRequestDeployments)
	handler.ServeHTTP(rr, req)

	assert.Equal(test, http.StatusBadRequest, rr.Result().StatusCode)
}

func TestListRequestDeploymentsWithWrongEndDate(test *testing.T) {

	req, err := http.NewRequest(
		"GET",
		"/requestDeployments?start_date=2020-01-01&end_date=2020-01",
		nil,
	)
	if err != nil {
		test.Fatal(err)
	}

	appContext := getRequestDeploymentAppContext()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listRequestDeployments)
	handler.ServeHTTP(rr, req)

	assert.Equal(test, http.StatusBadRequest, rr.Result().StatusCode)
}

func TestListRequestDeploymentsWithoutEndDate(test *testing.T) {

	req, err := http.NewRequest(
		"GET",
		"/requestDeployments?start_date=2020-01-01",
		nil,
	)
	if err != nil {
		test.Fatal(err)
	}

	appContext := getRequestDeploymentAppContext()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listRequestDeployments)
	handler.ServeHTTP(rr, req)

	assert.Equal(test, http.StatusBadRequest, rr.Result().StatusCode)
}

func TestListRequestDeploymentsWithPageSizeNotANumber(test *testing.T) {

	req, err := http.NewRequest(
		"GET",
		"/requestDeployments?start_date=2020-01-01&end_date=2020-01-01&pageSize=a",
		nil,
	)
	if err != nil {
		test.Fatal(err)
	}

	appContext := getRequestDeploymentAppContext()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listRequestDeployments)
	handler.ServeHTTP(rr, req)

	assert.Equal(test, http.StatusBadRequest, rr.Result().StatusCode)
}

func TestListRequestDeploymentsWithUserIDNotANumber(test *testing.T) {

	req, err := http.NewRequest(
		"GET",
		"/requestDeployments?start_date=2020-01-01&end_date=2020-01-01&user_id=a",
		nil,
	)
	if err != nil {
		test.Fatal(err)
	}

	appContext := getRequestDeploymentAppContext()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listRequestDeployments)
	handler.ServeHTTP(rr, req)

	assert.Equal(test, http.StatusBadRequest, rr.Result().StatusCode)
}

func TestListRequestDeploymentsErrorDatabaseQuery(test *testing.T) {

	req, err := http.NewRequest(
		"GET",
		"/requestDeployments?start_date=2020-01-01&end_date=2020-01-01",
		nil,
	)
	if err != nil {
		test.Fatal(err)
	}

	appContext := getRequestDeploymentAppContextWithError()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listRequestDeployments)
	handler.ServeHTTP(rr, req)

	assert.Equal(test, http.StatusInternalServerError, rr.Result().StatusCode)
}
func getRequestDeploymentAppContextWithError() *AppContext {
	appContext := AppContext{}
	requestDeploymentMock := mocks.RequestDeploymentDAOInterface{}

	requestDeployment := []model.RequestDeployments{}

	requestDeploymentMock.On(
		"ListRequestDeployments",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		-1,
		1,
		100,
	).Return(requestDeployment, errors.New("error"))

	appContext.Repositories.RequestDeploymentDAO = &requestDeploymentMock

	return &appContext
}
