package main

import (
	"errors"
	"testing"

	dbms2 "github.com/softplan/tenkai-api/pkg/dbms"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/handlers"
	"github.com/softplan/tenkai-api/pkg/rabbitmq/mocks"
	mockHelm "github.com/softplan/tenkai-api/pkg/service/_helm/mocks"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInitRepository(t *testing.T) {
	dbms := dbms2.Database{}
	repos := initRepository(&dbms)
	assert.NotNil(t, repos)
	assert.NotNil(t, repos.ConfigDAO)
	assert.NotNil(t, repos.VariableDAO)
	assert.NotNil(t, repos.ConfigDAO)
	assert.NotNil(t, repos.DockerDAO)
	assert.NotNil(t, repos.EnvironmentDAO)
	assert.NotNil(t, repos.ProductDAO)
	assert.NotNil(t, repos.UserDAO)
	assert.NotNil(t, repos.SolutionChartDAO)
	assert.NotNil(t, repos.SolutionDAO)
}

func TestInitAPIs(t *testing.T) {
	appContext := handlers.AppContext{}
	initAPIs(&appContext)
}

func TestInitCache(t *testing.T) {
	appContext := handlers.AppContext{}
	initCache(&appContext)
}

func TestCheckError(t *testing.T) {
	checkFatalError(nil)
}

func TestCreateQueue(t *testing.T) {
	appContext := handlers.AppContext{}

	mockRabbitMQ := mocks.RabbitInterface{}
	conn := &amqp.Connection{}
	channel := &amqp.Channel{}
	mockRabbitMQ.Mock.On("GetConnection", mock.Anything).Return(conn)
	mockRabbitMQ.Mock.On("GetChannel", mock.Anything).Return(channel)
	queue := amqp.Queue{}

	mockRabbitMQ.Mock.On(
		"QueueDeclare", mock.Anything, mock.Anything, true, false, false, false, mock.Anything).Return(queue, nil)
	appContext.RabbitImpl = &mockRabbitMQ
	createQueues(&appContext)
}

func TestCreateQueueWithError(t *testing.T) {
	appContext := handlers.AppContext{}

	mockRabbitMQ := mocks.RabbitInterface{}
	conn := &amqp.Connection{}
	channel := &amqp.Channel{}
	mockRabbitMQ.Mock.On("GetConnection", mock.Anything).Return(conn)
	mockRabbitMQ.Mock.On("GetChannel", mock.Anything).Return(channel)
	queue := amqp.Queue{}
	err := errors.New("Error")

	mockRabbitMQ.Mock.On(
		"QueueDeclare", mock.Anything, mock.Anything, true, false, false, false, mock.Anything).Return(queue, err)
	appContext.RabbitImpl = &mockRabbitMQ
	createQueues(&appContext)
}

func TestPublishRepoToQueueWithRepos(test *testing.T) {
	helmMock := mockHelm.HelmServiceInterface{}
	var repos []model.Repository
	var repo model.Repository
	repo.Name = "Teste"
	repo.URL = "Teste"
	repo.Username = "Teste"
	repo.Password = "Teste"
	repos = append(repos, repo)
	helmMock.Mock.On("GetRepositories").Return(repos, nil)
	mockRabbitMQ := mocks.RabbitInterface{}

	mockRabbitMQ.Mock.On("Publish",
		mock.Anything,
		"",
		mock.Anything,
		false,
		false,
		mock.Anything,
	).Return(nil)

	appContext := &handlers.AppContext{}
	appContext.RabbitImpl = &mockRabbitMQ
	appContext.HelmServiceAPI = &helmMock

	publishRepoToQueue(appContext)
}

func TestPublishRepoToQueueWithoutRepos(test *testing.T) {
	helmMock := mockHelm.HelmServiceInterface{}
	var repos []model.Repository
	err := errors.New("Error")
	helmMock.Mock.On("GetRepositories").Return(repos, err)
	appContext := &handlers.AppContext{}
	appContext.HelmServiceAPI = &helmMock
	assert.Panics(test, func() { publishRepoToQueue(appContext) }, appContext)
}
