package handlers

import (
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
	"github.com/stretchr/testify/mock"
)


func getRequestDeploymentAppContext() *AppContext {
	appContext := AppContext{}
	requestDeploymentMock := mocks.RequestDeploymentDAOInterface{}

	requestDeployment := []model.RequestDeployment{}

	requestDeploymentMock.On(
		"ListRequestDeployments",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		-1,
		mock.Anything,
		mock.Anything,
	).Return(requestDeployment, nil)

	requestDeploymentMock.On(
		"CountRequestDeployments",
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(int64(1), nil)

	appContext.Repositories.RequestDeploymentDAO = &requestDeploymentMock

	return &appContext
}

/*
func TestListRequestDeploymentsWithRightParams(test *testing.T) {

	req, err := http.NewRequest(
		"GET",
		"/requestDeployments?start_date=2020-01-01&end_date=2020-01-01",
		nil,
	)
	if err != nil {
		test.Fatal(err)
	}

	appContext := getAppContext()

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(appContext.listRequestDeployments)
	handler.ServeHTTP(rr, req)

	assert.Equal(test, http.StatusOK, rr.Result().StatusCode)
}*/