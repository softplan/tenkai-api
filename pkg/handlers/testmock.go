package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/softplan/tenkai-api/pkg/dbms/model"
	mockRepo "github.com/softplan/tenkai-api/pkg/dbms/repository/mocks"
)

//MockPrincipal injects a http header with the specified role to be used only for testing.
func MockPrincipal(req *http.Request, roles []string) {
	principal := model.Principal{Name: "alfa", Email: "beta@alfa.com", Roles: roles}
	pSe, _ := json.Marshal(principal)
	req.Header.Set("principal", string(pSe))
}

//MockGetByID mocks a call to GetByID function to be used only for testing.
func MockGetByID(appContext *AppContext) *mockRepo.EnvironmentDAOInterface {
	mockEnvDao := &mockRepo.EnvironmentDAOInterface{}
	env := MockGetEnv()
	mockEnvDao.On("GetByID", int(env.ID)).Return(&env, nil)
	appContext.Repositories.EnvironmentDAO = mockEnvDao
	return mockEnvDao
}

//MockGetEnv returns an Environment struct to be used only for testing.
func MockGetEnv() model.Environment {
	var env model.Environment
	env.ID = 999
	env.Group = "foo"
	env.Name = "bar"
	env.ClusterURI = "https://rancher-k8s-my-domain.com/k8s/clusters/c-kbfxr"
	env.CACertificate = "my-certificate"
	env.Token = "my-token"
	env.Namespace = "dev"
	env.Gateway = "my-gateway.istio-system.svc.cluster.local"
	return env
}

//MockVariable returns a Variable struct to be used only for testing.
func MockVariable() model.Variable {
	var variable model.Variable
	variable.Scope = "global"
	variable.Name = "username"
	variable.Value = "user"
	variable.Secret = false
	variable.Description = "Login username."
	variable.EnvironmentID = 999
	return variable
}
