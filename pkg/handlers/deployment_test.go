package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func getAppContext() *AppContext {
	appContext := AppContext{}
	deploymentMock := mocks.DeploymentDAOInterface{}

	deployments := []model.Deployments{}

	deploymentMock.On(
		"ListDeployments",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(deployments, nil)

	deploymentMock.On(
		"CountDeployments",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(int64(1), nil)

	appContext.Repositories.DeploymentDAO = &deploymentMock

	return &appContext
}

func TestListDeploymentsWithRightParams(test *testing.T) {

	req, err := http.NewRequest(
		"GET",
		"/deployments?start_date=2020-01-01&end_date=2020-01-01&user_id=1&environment_id=1&pageSize=10",
		nil,
	)
	if err != nil {
		test.Fatal(err)
	}

	appContext := getAppContext()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listDeployments)
	handler.ServeHTTP(rr, req)

	assert.Equal(test, http.StatusOK, rr.Result().StatusCode)
}

func TestListDeploymentsWrongPageNumber(test *testing.T) {

	req, err := http.NewRequest(
		"GET",
		"/deployments?start_date=2020-01-01&end_date=2020-01-01&user_id=1&environment_id=1&page=a",
		nil,
	)
	if err != nil {
		test.Fatal(err)
	}

	appContext := getAppContext()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listDeployments)
	handler.ServeHTTP(rr, req)

	assert.Equal(test, http.StatusBadRequest, rr.Result().StatusCode)
}

func TestListDeploymentsWithoutEndDate(test *testing.T) {

	req, err := http.NewRequest(
		"GET",
		"/deployments?start_date=2020-01-01",
		nil,
	)
	if err != nil {
		test.Fatal(err)
	}

	appContext := getAppContext()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listDeployments)
	handler.ServeHTTP(rr, req)

	assert.Equal(test, http.StatusBadRequest, rr.Result().StatusCode)
}

func TestListDeploymentsWithoutStartDate(test *testing.T) {

	req, err := http.NewRequest(
		"GET",
		"/deployments?end_date=2020-01-01",
		nil,
	)
	if err != nil {
		test.Fatal(err)
	}

	appContext := getAppContext()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listDeployments)
	handler.ServeHTTP(rr, req)

	assert.Equal(test, http.StatusBadRequest, rr.Result().StatusCode)
}

func TestListDeploymentsWithoutParams(test *testing.T) {
	req, err := http.NewRequest(
		"GET",
		"/deployments",
		nil,
	)
	if err != nil {
		test.Fatal(err)
	}

	appContext := getAppContext()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listDeployments)
	handler.ServeHTTP(rr, req)

	assert.Equal(test, http.StatusBadRequest, rr.Result().StatusCode)
}

func TestListDeploymentsOnlyWithEnvironmentAndUser(test *testing.T) {
	req, err := http.NewRequest(
		"GET",
		"/deployments?user_id=1&environment_id=1",
		nil,
	)
	if err != nil {
		test.Fatal(err)
	}

	appContext := getAppContext()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listDeployments)
	handler.ServeHTTP(rr, req)

	assert.Equal(test, http.StatusBadRequest, rr.Result().StatusCode)
}
