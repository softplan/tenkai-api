package handlers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/softplan/tenkai-api/pkg/constraints"
	"github.com/softplan/tenkai-api/pkg/global"
	helmapi "github.com/softplan/tenkai-api/pkg/service/_helm"
	"github.com/softplan/tenkai-api/pkg/util"

	"strings"

	"github.com/gorilla/mux"
	"github.com/softplan/tenkai-api/pkg/dbms/model"

	"github.com/softplan/tenkai-api/pkg/rabbitmq"
	"github.com/streadway/amqp"
)

func (appContext *AppContext) listCharts(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

	vars := mux.Vars(r)
	repo := vars["repo"] + "?"

	all, ok := r.URL.Query()["all"]
	allVersions := "all=false"
	if ok && len(all[0]) > 0 {
		if all[0] == "true" {
			allVersions = "all=true"
		}
	}

	url := global.HelmURL + "charts/" + repo + allVersions

	data, err := appContext.HelmService.DoGetRequest(url)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *AppContext) deleteHelmRelease(w http.ResponseWriter, r *http.Request) {

	isAdmin := false
	principal := util.GetPrincipal(r)
	if util.Contains(principal.Roles, constraints.TenkaiAdmin) {
		isAdmin = true
	}

	environmentIDs, ok := r.URL.Query()["environmentID"]
	if !ok || len(environmentIDs[0]) < 1 {
		http.Error(w, errors.New("param environmentID is required").Error(), http.StatusInternalServerError)
		return
	}

	releasesName, ok := r.URL.Query()["releaseName"]
	if !ok || len(releasesName[0]) < 1 {
		http.Error(w, errors.New("param releasesName is required").Error(), http.StatusInternalServerError)
		return
	}

	purges, ok := r.URL.Query()["purge"]
	if !ok || len(purges[0]) < 1 {
		http.Error(w, errors.New("param purges, is required").Error(), http.StatusInternalServerError)
		return
	}

	//Locate Environment
	envID, _ := strconv.Atoi(environmentIDs[0])

	has, err := appContext.hasAccess(principal.Email, envID)
	if err != nil || !has {
		http.Error(w, errors.New("Access Denied in this environment").Error(), http.StatusUnauthorized)
		return
	}

	//If not admin, verify authorization of user for specific environment
	if !isAdmin {
		auth, _ := appContext.hasEnvironmentRole(principal, uint(envID), "ACTION_HELM_PURGE")
		if !auth {
			http.Error(w, errors.New(global.AccessDenied).Error(), http.StatusUnauthorized)
			return
		}
	}

	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(int(envID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	kubeConfig := appContext.ConventionInterface.GetKubeConfigFileName(environment.Group, environment.Name)

	purge, _ := strconv.ParseBool(purges[0])
	err = appContext.HelmServiceAPI.DeleteHelmRelease(kubeConfig, releasesName[0], purge)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	auditValues := make(map[string]string)
	auditValues["environment"] = environment.Name
	auditValues["purge"] = strconv.FormatBool(purge)
	auditValues["name"] = releasesName[0]

	appContext.Auditing.DoAudit(r.Context(), appContext.Elk, principal.Email, "deleteHelmRelease", auditValues)

	w.WriteHeader(http.StatusOK)

}

func (appContext *AppContext) rollback(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

	var payload model.GetRevisionRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Locate Environment
	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(payload.EnvironmentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	kubeConfig := appContext.ConventionInterface.GetKubeConfigFileName(environment.Group, environment.Name)

	err = appContext.HelmServiceAPI.RollbackRelease(kubeConfig, payload.ReleaseName, payload.Revision)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (appContext *AppContext) revision(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

	var payload model.GetRevisionRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Locate Environment
	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(payload.EnvironmentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	kubeConfig := appContext.ConventionInterface.GetKubeConfigFileName(environment.Group, environment.Name)

	yaml, err := appContext.HelmServiceAPI.Get(kubeConfig, payload.ReleaseName, payload.Revision)

	data, _ := json.Marshal(yaml)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *AppContext) listReleaseHistory(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

	var payload model.HistoryRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Locate Environment
	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(payload.EnvironmentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	kubeConfig := appContext.ConventionInterface.GetKubeConfigFileName(environment.Group, environment.Name)

	history, err := appContext.HelmServiceAPI.GetHelmReleaseHistory(kubeConfig, payload.ReleaseName)

	data, _ := json.Marshal(history)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *AppContext) listHelmDeploymentsByEnvironment(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Locate Environment
	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(global.ContentType, global.JSONContentType)

	kubeConfig := appContext.ConventionInterface.GetKubeConfigFileName(environment.Group, environment.Name)

	result, err := appContext.HelmServiceAPI.ListHelmDeployments(kubeConfig, environment.Namespace)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (appContext *AppContext) hasConfigMap(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

	var payload model.GetChartRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	hasConfigMap, err := appContext.hasConfigMapCached(payload.ChartName, payload.ChartVersion)

	w.WriteHeader(http.StatusOK)
	if err != nil {
		w.Write([]byte("{\"result\":\"false\"}"))
	} else {
		if hasConfigMap {
			w.Write([]byte("{\"result\":\"true\"}"))
		} else {
			w.Write([]byte("{\"result\":\"false\"}"))
		}
	}

}

func (appContext *AppContext) getChartVariables(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)
	var payload model.GetChartRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	chartName, err := appContext.getChartName(payload.ChartName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := appContext.HelmServiceAPI.GetTemplate(&appContext.Mutex, chartName, payload.ChartVersion, "values")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(result)

}

func (appContext *AppContext) getHelmCommand(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)
	var payload model.MultipleInstallPayload

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	out := &bytes.Buffer{}

	var fullCommand string
	for _, element := range payload.Deployables {

		//Locate Environment
		environment, err := appContext.Repositories.EnvironmentDAO.GetByID(element.EnvironmentID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		command, errX := appContext.simpleInstall(environment, element, out, false, true, "", -1)
		if errX != nil {
			http.Error(w, err.Error(), 501)
			return
		}

		fullCommand = fullCommand + "\n\n" + command

	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fullCommand))

}

func (appContext *AppContext) hasConfigMapCached(chart string, version string) (bool, error) {
	cmKey := chart + version
	cachedValue, ok := appContext.ConfigMapCache.Load(cmKey)
	var hasConfigMap bool
	if ok {
		hasConfigMap = cachedValue.(bool)
	} else if !ok || cachedValue == "" {
		result, err := appContext.HelmServiceAPI.GetTemplate(&appContext.Mutex, chart, version, "deployment")
		if err != nil {
			return false, err
		}

		deployment := string(result)
		hasConfigMap = strings.Index(deployment, "gcm") > 0

		appContext.ConfigMapCache.Store(cmKey, hasConfigMap)
	}
	return hasConfigMap, nil
}

func (appContext *AppContext) loadConfigMap(deployables []model.InstallPayload, environmentID int) ([]model.InstallPayload, error) {
	configMaps := make([]model.InstallPayload, 0)

	var config model.ConfigMap
	var err error
	if config, err = appContext.Repositories.ConfigDAO.GetConfigByName("commonValuesConfigMapChart"); err != nil {
		return configMaps, err
	}

	for _, d := range deployables {
		configMaps = append(configMaps, d)
		hasConfigMap, err := appContext.hasConfigMapCached(d.Chart, d.ChartVersion)
		if err != nil {
			return deployables, err
		}

		if hasConfigMap {
			var ip model.InstallPayload
			ip.Name = d.Name + "-gcm"
			ip.Chart = config.Value
			ip.EnvironmentID = environmentID

			configMaps = append(configMaps, ip)
		}
	}
	return configMaps, nil
}

func (appContext *AppContext) multipleInstall(w http.ResponseWriter, r *http.Request) {

	isAdmin := false
	principal := util.GetPrincipal(r)
	if util.Contains(principal.Roles, constraints.TenkaiAdmin) {
		isAdmin = true
	}

	w.Header().Set(global.ContentType, global.JSONContentType)
	var payload model.MultipleInstallPayload

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	var environments []*model.Environment

	for _, environmentID := range payload.EnvironmentIDs {
		//Locate Environment
		environment, err := appContext.Repositories.EnvironmentDAO.GetByID(environmentID)
		if err != nil {
			msg := err.Error()
			if err.Error() == "record not found" {
				msg = "Environment " + strconv.Itoa(environmentID) + " not found"
			}
			http.Error(w, msg, http.StatusInternalServerError)
			return
		}

		//If not admin, verify authorization of user for specific environment
		if !isAdmin {
			auth, _ := appContext.hasEnvironmentRole(principal, environment.ID, "ACTION_DEPLOY")
			if !auth {
				http.Error(w, errors.New(global.AccessDenied).Error(), http.StatusUnauthorized)
				return
			}
		}

		environments = append(environments, environment)
	}

	user, err := appContext.Repositories.UserDAO.FindByEmail(principal.Email)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	requestDeployment := model.RequestDeployment{}
	requestDeployment.Success = false
	requestDeployment.Processed = false
	requestDeployment.UserID = user.ID
	requestDeploymentID, _ := appContext.Repositories.RequestDeploymentDAO.CreateRequestDeployment(requestDeployment)

	for _, environment := range environments {
		configMaps, err := appContext.loadConfigMap(payload.Deployables, int(environment.ID))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		payload.Deployables = configMaps
		out := &bytes.Buffer{}

		for _, element := range payload.Deployables {
			if err = appContext.updateImageTagBeforeInstallProduct(payload.ProductVersionID,
				int(environment.ID), element.Chart); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}

			_, err = appContext.simpleInstall(environment, element, out, false, false, principal.Email, requestDeploymentID)
			if err != nil {
				http.Error(w, err.Error(), 501)
				return
			}

			auditValues := make(map[string]string)
			auditValues["environment"] = environment.Name
			auditValues["chartName"] = element.Chart
			auditValues["name"] = element.Name

			appContext.Auditing.DoAudit(r.Context(), appContext.Elk, principal.Email, "deploy", auditValues)

			if payload.ProductVersionID > 0 {
				pv, err := appContext.Repositories.ProductDAO.ListProductVersionsByID(payload.ProductVersionID)
				if err != nil {
					http.Error(w, err.Error(), 501)
					return
				}
				environment.ProductVersion = pv.Version
				if err := appContext.Repositories.EnvironmentDAO.EditEnvironment(*environment); err != nil {
					http.Error(w, err.Error(), 501)
					return
				}

				appContext.triggerProductDeploymentWebhook(int(environment.ID),
					pv.ProductID, environment.Name, pv.Version)
			}
		}

	}
	w.WriteHeader(http.StatusOK)
}

func (appContext *AppContext) triggerProductDeploymentWebhook(
	environmentID int, productID int, envName string, releaseVersion string) {

	var err error
	var webHooks []model.WebHook
	webHooks, err = appContext.Repositories.WebHookDAO.
		ListWebHooksByEnvAndType(environmentID, "HOOK_DEPLOY_PRODUCT")
	if err != nil {
		log.Println("Error trying to find webhooks", err)
		return
	}

	var product model.Product
	if product, err = appContext.Repositories.ProductDAO.FindProductByID(productID); err != nil {
		log.Println("Error trying to find product", err)
		return
	}

	for _, hook := range webHooks {
		var p model.WebHookPostPayload
		p.Environment = envName
		p.ProductName = product.Name
		p.Release = releaseVersion

		payloadStr, _ := json.Marshal(p)
		if _, err := http.Post(hook.URL, "application/json", bytes.NewBuffer(payloadStr)); err != nil {
			log.Println("Error trying to post to webhook: ", hook.URL, err)
			return
		}
	}
}

func (appContext *AppContext) updateImageTagBeforeInstallProduct(productVersionID int, envID int, chart string) error {
	if productVersionID > 0 {
		pvs, err := appContext.Repositories.ProductDAO.ListProductsVersionServices(productVersionID)
		if err != nil {
			return err
		}
		varImgTag, _ := appContext.Repositories.VariableDAO.GetVarImageTagByEnvAndScope(envID, chart)
		if varImgTag.ID > 0 {
			for _, pvsvc := range pvs {
				if varImgTag.Scope == strings.Split(pvsvc.ServiceName, " - ")[0] {
					varImgTag.Value = pvsvc.DockerImageTag
					if err := appContext.Repositories.VariableDAO.EditVariable(varImgTag); err != nil {
						return err
					}
				}
			}
		}

	}
	return nil
}

func (appContext *AppContext) install(w http.ResponseWriter, r *http.Request) {

	isAdmin := false
	principal := util.GetPrincipal(r)
	if util.Contains(principal.Roles, constraints.TenkaiAdmin) {
		isAdmin = true
	}

	w.Header().Set(global.ContentType, global.JSONContentType)
	var payload model.InstallPayload

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	out := &bytes.Buffer{}

	//Locate Environment
	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(payload.EnvironmentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//If not admin, verify authorization of user for specific environment
	if !isAdmin {

		auth, _ := appContext.hasEnvironmentRole(principal, environment.ID, "ACTION_DEPLOY")
		if !auth {
			http.Error(w, errors.New(global.AccessDenied).Error(), http.StatusUnauthorized)
			return
		}
	}

	_, err = appContext.simpleInstall(environment, payload, out, false, false, principal.Email, -1)
	if err != nil {
		fmt.Println(out.String())
		http.Error(w, err.Error(), 501)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (appContext *AppContext) helmDryRun(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)
	var payload model.InstallPayload

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	out := &bytes.Buffer{}

	//Locate Environment
	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(payload.EnvironmentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = appContext.simpleInstall(environment, payload, out, true, false, "", -1)

	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(out.Bytes())

}

func (appContext *AppContext) getArgsWithHelmDefault(variables []model.Variable, helmVars map[string]interface{}, globalVariables []model.Variable, environment *model.Environment) []string {

	var args []string
	var keys []string
	for i, item := range variables {
		if item.Secret {
			byteValues, _ := hex.DecodeString(item.Value)
			value, err := util.Decrypt(byteValues, appContext.Configuration.App.Passkey)
			if err == nil {
				variables[i].Value = string(value)
			}
		}
		if len(item.Name) > 0 && len(item.Value) > 0 {
			value := replace(item.Value, *environment, globalVariables)
			if value != "" {
				keys = append(keys, normalizeVariableName(item.Name))
			}
			if value == "T_EMPTY" {
				value = ""
			}
			args = append(args, normalizeVariableName(item.Name)+"="+value)
		}
	}

	dt := time.Now()
	args = append(args, "app.dateHour="+dt.String())
	keys = append(keys, "app.dateHour")

	for key, value := range helmVars {
		if !util.Contains(keys, normalizeVariableName(key)) {
			svalue, ok := value.(string)
			if ok {
				args = append(args, normalizeVariableName(key)+"="+replace(svalue, *environment, globalVariables))
			}
		}
	}

	return args
}

func (appContext *AppContext) simpleInstall(environment *model.Environment, installPayload model.InstallPayload, out *bytes.Buffer, dryRun bool, helmCommandOnly bool, userID string, requestDeploymentID int) (string, error) {

	//WARNING - VERIFY IF CONFIG FILE EXISTS !!! This is the cause of  u.client.ReleaseHistory fail sometimes.

	searchTerm := installPayload.Chart
	if strings.Index(installPayload.Name, "gcm") > -1 {
		searchTerm = installPayload.Name
	}
	variables, err := appContext.Repositories.VariableDAO.GetAllVariablesByEnvironmentAndScope(int(environment.ID), searchTerm)
	globalVariables := appContext.getGlobalVariables(int(environment.ID))

	helmVars, err := appContext.getHelmChartAppVars(installPayload.Chart, installPayload.ChartVersion)
	if err != nil {
		return "", err
	}
	args := appContext.getArgsWithHelmDefault(variables, helmVars, globalVariables, environment)

	//Add Default Gateway
	if len(environment.Gateway) > 0 {
		args = append(args, "istio.virtualservices.gateways[0]="+environment.Gateway)
	}

	if err == nil {
		name := installPayload.Name + "-" + environment.Namespace
		kubeConfig := appContext.ConventionInterface.GetKubeConfigFileName(environment.Group, environment.Name)

		if !helmCommandOnly {

			upgradeRequest := helmapi.UpgradeRequest{}
			upgradeRequest.Kubeconfig = kubeConfig
			upgradeRequest.Namespace = environment.Namespace
			upgradeRequest.ChartVersion = installPayload.ChartVersion
			upgradeRequest.Chart = installPayload.Chart
			upgradeRequest.Variables = args
			upgradeRequest.Dryrun = dryRun
			upgradeRequest.Release = name

			deployment := model.Deployment{}
			deployment.EnvironmentID = environment.ID
			deployment.RequestDeploymentID = uint(requestDeploymentID)
			deployment.Chart = installPayload.Chart
			deployment.Processed = false
			deploymentID, _ := appContext.Repositories.DeploymentDAO.CreateDeployment(deployment)

			queuePayload := rabbitmq.PayloadRabbit{
				UpgradeRequest: upgradeRequest,
				Name:           environment.Name,
				Token:          environment.Token,
				Filename:       appContext.K8sConfigPath + environment.Group + "_" + environment.Name,
				CACertificate:  environment.CACertificate,
				ClusterURI:     environment.ClusterURI,
				Namespace:      environment.Namespace,
				DeploymentID:   uint(deploymentID),
			}

			queuePayloadJSON, _ := json.Marshal(queuePayload)

			err := appContext.RabbitImpl.Publish(
				appContext.RabbitMQChannel,
				"",
				rabbitmq.InstallQueue,
				false,
				false,
				amqp.Publishing{
					ContentType: "application/json",
					Body:        queuePayloadJSON,
				},
			)
			return "", err
		}
		return getHelmMessage(name, args, environment, installPayload.Chart), nil
	}
	return "", nil
}

func (appContext *AppContext) getChartName(name string) (string, error) {

	searchTerms := []string{name}
	searchResult := appContext.HelmServiceAPI.SearchCharts(searchTerms, false)

	if len(*searchResult) > 0 {
		r := *searchResult
		return r[0].Name, nil
	}
	return "", errors.New("Chart does not exists")
}

func (appContext *AppContext) doUpgrade(upgradeRequest helmapi.UpgradeRequest, out *bytes.Buffer) (string, error) {
	var err error
	upgradeRequest.Chart, err = appContext.getChartName(upgradeRequest.Chart)
	if err != nil {
		return "", err
	}
	err = appContext.HelmServiceAPI.Upgrade(upgradeRequest, out)
	if err != nil {
		return "", err
	}
	return "", nil
}

func getHelmMessage(name string, args []string, environment *model.Environment, chart string) string {
	var message string

	message = "helm upgrade --install " + name + " \\\n"

	for _, e := range args {
		message = message + " --set \"" + e + "\" " + " \\\n"
	}
	message = message + " " + chart + " --namespace=" + environment.Namespace
	return message
}

func replace(value string, environment model.Environment, variables []model.Variable) string {
	newValue := strings.Replace(value, "${NAMESPACE}", environment.Namespace, -1)
	keywords := util.GetReplacebleKeyName(newValue)
	for _, keyword := range keywords {
		for _, element := range variables {
			if element.Name == keyword {
				newValue = strings.Replace(newValue, "${"+element.Name+"}", element.Value, -1)
				break
			}
		}
	}
	return newValue
}

func normalizeVariableName(value string) string {
	if strings.Index(value, "istio.") > -1 || (strings.Index(value, "image.")) > -1 || (strings.Index(value, "service.")) > -1 {
		return value
	}
	return "app." + value
}

func (appContext *AppContext) getGlobalVariables(id int) []model.Variable {
	variables, _ := appContext.Repositories.VariableDAO.GetAllVariablesByEnvironmentAndScope(id, "global")

	for i, e := range variables {
		if e.Secret {
			byteValues, _ := hex.DecodeString(e.Value)
			value, err := util.Decrypt(byteValues, appContext.Configuration.App.Passkey)
			if err == nil {
				variables[i].Value = string(value)
			}
		}
	}

	return variables
}
