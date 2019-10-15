package main

import (
	dockerapi "github.com/softplan/tenkai-api/service/docker"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/olivere/elastic"
	"github.com/softplan/tenkai-api/audit"
	"github.com/softplan/tenkai-api/dbms"
	helmapi "github.com/softplan/tenkai-api/service/helm"

	"github.com/softplan/tenkai-api/configs"
	"github.com/softplan/tenkai-api/global"

	"github.com/gorilla/mux"
)

const (
	configFileName = "app"
)

type appContext struct {
	dockerServiceApi dockerapi.DockerServiceInterface
	k8sConfigPath    string
	configuration    *configs.Configuration
	configDAO        dbms.ConfigDAOInterface
	environmentDAO   dbms.EnvironmentDAOInterface
	database         dbms.Database
	elk              *elastic.Client
	mutex            sync.Mutex
	chartImageCache  sync.Map
	dockerTagsCache  sync.Map
}

func main() {
	logFields := global.AppFields{global.Function: "main"}

	_ = os.Mkdir(global.KubeConfigBasePath, 0777)

	if _, err := os.Stat(global.HelmDir + "/repository/repositories.yaml"); os.IsNotExist(err) {
		helmapi.InitializeHelm()
	}

	global.Logger.Info(logFields, "carregando configurações")

	config, error := configs.ReadConfig(configFileName)
	checkFatalError(error)

	appContext := &appContext{configuration: config}

	appContext.dockerTagsCache = sync.Map{}
	appContext.chartImageCache = sync.Map{}

	//appContext.dockerTagsCache = make(map[string]time.Time)
	//appContext.chartImageCache = make(map[string]string)

	dbmsURI := config.App.Dbms.URI

	//Dbms connection
	appContext.database.Connect(dbmsURI, dbmsURI == "")
	defer appContext.database.Db.Close()

	appContext.k8sConfigPath = global.KubeConfigBasePath
	appContext.dockerServiceApi = &dockerapi.DockerService{}

	//Init DAO
	appContext.configDAO = &dbms.ConfigDAOImpl{Db: appContext.database.Db}
	appContext.environmentDAO = &dbms.EnvironmentDAOImpl{Db: appContext.database.Db}

	//Elk setup
	appContext.elk, _ = audit.ElkClient(config.App.Elastic.URL, config.App.Elastic.Username, config.App.Elastic.Password)

	global.Logger.Info(logFields, "iniciando o servidor http")
	startHTTPServer(appContext)
}

func startHTTPServer(appContext *appContext) {

	//===

	port := appContext.configuration.Server.Port
	global.Logger.Info(global.AppFields{global.Function: "startHTTPServer", "port": port}, "online - listen and server")

	r := mux.NewRouter()

	r.HandleFunc("/install", appContext.install).Methods("POST")
	r.HandleFunc("/multipleInstall", appContext.multipleInstall).Methods("POST")
	r.HandleFunc("/getHelmCommand", appContext.getHelmCommand).Methods("POST")

	r.HandleFunc("/getVariablesNotUsed/{id}", appContext.getVariablesNotUsed).Methods("GET")

	r.HandleFunc("/listVariables", appContext.getVariablesByEnvironmentAndScope).Methods("POST")
	r.HandleFunc("/saveVariableValues", appContext.saveVariableValues).Methods("POST")
	r.HandleFunc("/getChartVariables", appContext.getChartVariables).Methods("POST")
	r.HandleFunc("/listHelmDeploymentsByEnvironment/{id}", appContext.listHelmDeploymentsByEnvironment).Methods("GET")
	r.HandleFunc("/listReleaseHistory", appContext.listReleaseHistory).Methods("POST")
	r.HandleFunc("/rollback", appContext.rollback).Methods("POST")

	r.HandleFunc("/charts/{repo}", appContext.listCharts).Methods("GET")
	r.HandleFunc("/listPods/{id}", appContext.pods).Methods("GET")
	r.HandleFunc("/listServices/{id}", appContext.services).Methods("GET")

	r.HandleFunc("/variables", appContext.addVariables).Methods("POST")
	r.HandleFunc("/variables/{envId}", appContext.getVariables).Methods("GET")
	r.HandleFunc("/variables/delete/{id}", appContext.deleteVariable).Methods("DELETE")
	r.HandleFunc("/deletePod", appContext.deletePod).Methods("DELETE")

	r.HandleFunc("/variables/edit", appContext.editVariable).Methods("POST")

	r.HandleFunc("/environments/delete/{id}", appContext.deleteEnvironment).Methods("DELETE")
	r.HandleFunc("/environments/edit", appContext.editEnvironment).Methods("POST")
	r.HandleFunc("/environments", appContext.addEnvironments).Methods("POST")
	r.HandleFunc("/environments", appContext.getEnvironments).Methods("GET")
	r.HandleFunc("/environments/all", appContext.getAllEnvironments).Methods("GET")
	r.HandleFunc("/environments/export/{id}", appContext.export).Methods("GET")
	r.HandleFunc("/hasConfigMap", appContext.hasConfigMap).Methods("POST")

	r.HandleFunc("/revision", appContext.revision).Methods("POST")

	r.HandleFunc("/environments/duplicate/{id}", appContext.duplicateEnvironments).Methods("GET")

	r.HandleFunc("/repositories", appContext.listRepositories).Methods("GET")
	r.HandleFunc("/repositories", appContext.newRepository).Methods("POST")
	r.HandleFunc("/repositories/{name}", appContext.deleteRepository).Methods("DELETE")

	r.HandleFunc("/deleteHelmRelease", appContext.deleteHelmRelease).Methods("DELETE")
	r.HandleFunc("/helmDryRun", appContext.helmDryRun).Methods("POST")

	r.HandleFunc("/releases", appContext.listReleases).Methods("GET")
	r.HandleFunc("/releases", appContext.newRelease).Methods("POST")
	r.HandleFunc("/releases/{id}", appContext.deleteRelease).Methods("DELETE")

	r.HandleFunc("/dependencies", appContext.listDependencies).Methods("GET")
	r.HandleFunc("/dependencies", appContext.newDependency).Methods("POST")
	r.HandleFunc("/dependencies/{id}", appContext.deleteDependency).Methods("DELETE")

	r.HandleFunc("/solutions", appContext.listSolution).Methods("GET")
	r.HandleFunc("/solutions", appContext.newSolution).Methods("POST")
	r.HandleFunc("/solutions/edit", appContext.editSolution).Methods("POST")
	r.HandleFunc("/solutions/{id}", appContext.deleteSolution).Methods("DELETE")

	r.HandleFunc("/products", appContext.listProducts).Methods("GET")
	r.HandleFunc("/products", appContext.newProduct).Methods("POST")
	r.HandleFunc("/products/edit", appContext.editProduct).Methods("POST")
	r.HandleFunc("/products/{id}", appContext.deleteProduct).Methods("DELETE")

	r.HandleFunc("/productVersions", appContext.listProductVersions).Methods("GET")
	r.HandleFunc("/productVersions", appContext.newProductVersion).Methods("POST")
	r.HandleFunc("/productVersions/{id}", appContext.deleteProductVersion).Methods("DELETE")

	r.HandleFunc("/productVersionServices", appContext.listProductVersionServices).Methods("GET")
	r.HandleFunc("/productVersionServices", appContext.newProductVersionService).Methods("POST")
	r.HandleFunc("/productVersionServices/edit", appContext.editProductVersionService).Methods("POST")
	r.HandleFunc("/productVersionServices/{id}", appContext.deleteProductVersionService).Methods("DELETE")

	r.HandleFunc("/dockerRepo", appContext.listDockerRepositories).Methods("GET")
	r.HandleFunc("/dockerRepo", appContext.newDockerRepository).Methods("POST")
	r.HandleFunc("/dockerRepo/{id}", appContext.deleteDockerRepository).Methods("DELETE")

	r.HandleFunc("/solutionCharts", appContext.listSolutionCharts).Methods("GET")
	r.HandleFunc("/solutionCharts", appContext.newSolutionChart).Methods("POST")
	r.HandleFunc("/solutionCharts/{id}", appContext.deleteSolutionChart).Methods("DELETE")

	r.HandleFunc("/analyse", appContext.analyse).Methods("POST")

	r.HandleFunc("/deployTrafficRule", appContext.deployTrafficRule).Methods("POST")

	r.HandleFunc("/repoUpdate", appContext.repoUpdate).Methods("GET")

	r.HandleFunc("/repo/default", appContext.setDefaultRepo).Methods("POST")
	r.HandleFunc("/repo/default", appContext.getDefaultRepo).Methods("GET")

	r.HandleFunc("/users/createOrUpdate", appContext.createOrUpdateUser).Methods("POST")

	r.HandleFunc("/users", appContext.newUser).Methods("POST")
	r.HandleFunc("/users", appContext.listUsers).Methods("GET")
	r.HandleFunc("/users/{id}", appContext.deleteUser).Methods("DELETE")

	r.HandleFunc("/promote", appContext.promote).Methods("GET")

	r.HandleFunc("/listDockerTags", appContext.listDockerTags).Methods("POST")

	r.HandleFunc("/permissions/users/{userId}/environments/{environmentId}", appContext.newEnvironmentPermission).Methods("GET")

	r.HandleFunc("/settings", appContext.addSettings).Methods("POST")
	r.HandleFunc("/getSettingList", appContext.getSettingList).Methods("POST")

	r.HandleFunc("/", appContext.rootHandler)

	log.Fatal(http.ListenAndServe(":"+port, commonHandler(r)))

}
