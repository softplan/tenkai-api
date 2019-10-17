package main

import (
	audit2 "github.com/softplan/tenkai-api/pkg/audit"
	"github.com/softplan/tenkai-api/pkg/configs"
	"github.com/softplan/tenkai-api/pkg/dbms"
	"github.com/softplan/tenkai-api/pkg/dbms/repository"
	"github.com/softplan/tenkai-api/pkg/global"
	"github.com/softplan/tenkai-api/pkg/handlers"
	dockerapi "github.com/softplan/tenkai-api/pkg/service/docker"
	helmapi "github.com/softplan/tenkai-api/pkg/service/helm"
	"log"
	"os"
	"sync"
)

const (
	configFileName = "app"
)

func main() {
	logFields := global.AppFields{global.Function: "main"}
	_ = os.Mkdir(global.KubeConfigBasePath, 0777)

	global.Logger.Info(logFields, "loading config properties")

	config, error := configs.ReadConfig(configFileName)
	checkFatalError(error)

	appContext := &handlers.AppContext{Configuration: config}

	dbmsURI := config.App.Dbms.URI

	initCache(appContext)
	initAPIs(appContext)

	//Dbms connection
	appContext.Database.Connect(dbmsURI, dbmsURI == "")
	defer appContext.Database.Db.Close()

	appContext.K8sConfigPath = global.KubeConfigBasePath
	initializeHelm(appContext)

	appContext.Repositories = initRepository(&appContext.Database)

	//Elk setup
	appContext.Elk, _ = appContext.Auditing.ElkClient(config.App.Elastic.URL, config.App.Elastic.Username, config.App.Elastic.Password)

	global.Logger.Info(logFields, "http server started")
	handlers.StartHTTPServer(appContext)
}

func initializeHelm(appContext *handlers.AppContext) {
	if _, err := os.Stat(global.HelmDir + "/repository/repositories.yaml"); os.IsNotExist(err) {
		appContext.HelmServiceAPI.InitializeHelm()
	}
}

func initCache(appContext *handlers.AppContext) {
	appContext.DockerTagsCache = sync.Map{}
	appContext.ChartImageCache = sync.Map{}
}

func initAPIs(appContext *handlers.AppContext) {
	appContext.DockerServiceAPI = &dockerapi.DockerService{}
	appContext.HelmServiceAPI = &helmapi.HelmServiceImpl{}
	appContext.Auditing = &audit2.AuditingImpl{}
}

func initRepository(database *dbms.Database) handlers.Repositories {
	repositories := handlers.Repositories{}
	repositories.ConfigDAO = &repository.ConfigDAOImpl{Db: database.Db}
	repositories.DependencyDAO = &repository.DependencyDAOImpl{Db: database.Db}
	repositories.DockerDAO = &repository.DockerDAOImpl{Db: database.Db}
	repositories.EnvironmentDAO = &repository.EnvironmentDAOImpl{Db: database.Db}
	repositories.ProductDAO = &repository.ProductDAOImpl{Db: database.Db}
	repositories.ReleaseDAO = &repository.ReleaseDAOImpl{Db: database.Db}
	repositories.SolutionDAO = &repository.SolutionDAOImpl{Db: database.Db}
	repositories.SolutionChartDAO = &repository.SolutionChartDAOImpl{Db: database.Db}
	repositories.UserDAO = &repository.UserDAOImpl{Db: database.Db}
	repositories.VariableDAO = &repository.VariableDAOImpl{Db: database.Db}
	return repositories
}

func checkFatalError(err error) {
	if err != nil {
		global.Logger.Error(global.AppFields{global.Function: "upload", "error": err}, "erro fatal")
		log.Fatal(err)
	}
}
