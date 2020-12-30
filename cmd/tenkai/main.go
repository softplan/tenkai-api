package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	audit2 "github.com/softplan/tenkai-api/pkg/audit"
	"github.com/softplan/tenkai-api/pkg/configs"
	"github.com/softplan/tenkai-api/pkg/dbms"
	"github.com/softplan/tenkai-api/pkg/dbms/repository"
	"github.com/softplan/tenkai-api/pkg/global"
	"github.com/softplan/tenkai-api/pkg/handlers"
	"github.com/softplan/tenkai-api/pkg/rabbitmq"
	helmapi "github.com/softplan/tenkai-api/pkg/service/_helm"
	"github.com/softplan/tenkai-api/pkg/service/core"
	dockerapi "github.com/softplan/tenkai-api/pkg/service/docker"
	"github.com/softplan/tenkai-api/pkg/tenkaihelm"
	"github.com/streadway/amqp"
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

	//RabbitMQ Connection
	appContext.RabbitImpl = rabbitmq.RabbitImpl{}
	appContext.RabbitMQConn = appContext.RabbitImpl.GetConnection(config.App.Rabbit.URI)
	appContext.RabbitMQChannel = appContext.RabbitImpl.GetChannel(appContext.RabbitMQConn)
	defer appContext.RabbitMQConn.Close()
	defer appContext.RabbitMQChannel.Close()
	createQueues(appContext)
	publishRepoToQueue(appContext)
	go handlers.StartConsumerQueue(appContext, rabbitmq.ResultInstallQueue)

	appContext.HelmService = tenkaihelm.HelmAPIImpl{}

	createEnvironmentFiles(appContext)

	global.Logger.Info(logFields, "http server started")
	handlers.StartHTTPServer(appContext)
}

func createEnvironmentFiles(appContext *handlers.AppContext) {
	envs, err := appContext.Repositories.EnvironmentDAO.GetAllEnvironments("")
	if err != nil {
		logFields := global.AppFields{global.Function: "createEnvironmentFiles"}
		global.Logger.Error(logFields, "Fail on create environment files")
		return
	}
	for _, env := range envs {
		handlers.CreateEnvironmentFile(env.Name, env.Token, appContext.K8sConfigPath+env.Group+"_"+env.Name,
			env.CACertificate, env.ClusterURI, env.Namespace)
	}
}

func createQueues(appContext *handlers.AppContext) {
	createQueue(rabbitmq.InstallQueue, appContext)
	createQueue(rabbitmq.ResultInstallQueue, appContext)
	createQueue(rabbitmq.RepositoriesQueue, appContext)
	createQueue(rabbitmq.DeleteRepoQueue, appContext)
}

func publishRepoToQueue(appContext *handlers.AppContext) {
	repositories, err := appContext.HelmServiceAPI.GetRepositories()
	if err != nil {
		panic("Can not retrieve repositories from helm service API")
	}
	for _, repo := range repositories {
		if repo.Name != "local" && repo.Name != "stable" {
			queuePayloadJSON, _ := json.Marshal(repo)
			appContext.RabbitImpl.Publish(
				appContext.RabbitMQChannel,
				"",
				rabbitmq.RepositoriesQueue,
				false,
				false,
				amqp.Publishing{
					ContentType: "application/json",
					Body:        queuePayloadJSON,
				},
			)
		}
	}
}

func createQueue(queueName string, appContext *handlers.AppContext) {
	_, err := appContext.RabbitImpl.QueueDeclare(appContext.RabbitMQChannel, queueName, true, false, false, false, nil)
	if err != nil {
		global.Logger.Error(
			global.AppFields{global.Function: "createQueue"},
			"Could not declare "+queueName+" - "+err.Error())
	}
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

	appContext.DockerServiceAPI = dockerapi.DockerServiceBuilder()
	appContext.HelmServiceAPI = helmapi.HelmServiceBuilder()

	appContext.Auditing = audit2.AuditingBuilder()
	appContext.ConventionInterface = &core.ConventionImpl{}
}

func initRepository(database *dbms.Database) handlers.Repositories {
	repositories := handlers.Repositories{}
	repositories.ConfigDAO = &repository.ConfigDAOImpl{Db: database.Db}
	repositories.DockerDAO = &repository.DockerDAOImpl{Db: database.Db}
	repositories.EnvironmentDAO = &repository.EnvironmentDAOImpl{Db: database.Db}
	repositories.ProductDAO = &repository.ProductDAOImpl{Db: database.Db}
	repositories.SolutionDAO = &repository.SolutionDAOImpl{Db: database.Db}
	repositories.SolutionChartDAO = &repository.SolutionChartDAOImpl{Db: database.Db}
	repositories.UserDAO = &repository.UserDAOImpl{Db: database.Db}
	repositories.VariableDAO = &repository.VariableDAOImpl{Db: database.Db}
	repositories.VariableRuleDAO = &repository.VariableRuleDAOImpl{Db: database.Db}
	repositories.ValueRuleDAO = &repository.ValueRuleDAOImpl{Db: database.Db}
	repositories.CompareEnvsQueryDAO = &repository.CompareEnvsQueryDAOImpl{Db: database.Db}
	repositories.SecurityOperationDAO = &repository.SecurityOperationDAOImpl{Db: database.Db}
	repositories.UserEnvironmentRoleDAO = &repository.UserEnvironmentRoleDAOImpl{Db: database.Db}
	repositories.NotesDAO = &repository.NotesDAOImpl{Db: database.Db}
	repositories.WebHookDAO = &repository.WebHookDAOImpl{Db: database.Db}
	repositories.DeploymentDAO = &repository.DeploymentDAOImpl{Db: database.Db}
	repositories.RequestDeploymentDAO = &repository.RequestDeploymentDAOImpl{Db: database.Db}

	return repositories
}

func checkFatalError(err error) {
	if err != nil {
		global.Logger.Error(global.AppFields{global.Function: "upload", "error": err}, "erro fatal")
		log.Fatal(err)
	}
}
