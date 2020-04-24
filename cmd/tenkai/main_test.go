package main

import (
	"testing"

	dbms2 "github.com/softplan/tenkai-api/pkg/dbms"
	"github.com/softplan/tenkai-api/pkg/handlers"
	"github.com/stretchr/testify/assert"
)

func TestInitRepository(t *testing.T) {
	dbms := dbms2.Database{}
	repos := initRepository(&dbms)
	assert.NotNil(t, repos)
	assert.NotNil(t, repos.ConfigDAO)
	assert.NotNil(t, repos.VariableDAO)
	assert.NotNil(t, repos.ConfigDAO)
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
