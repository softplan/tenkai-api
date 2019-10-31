package handlers

import (
	"bytes"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/global"
	helmapi "github.com/softplan/tenkai-api/pkg/service/_helm"
	"github.com/softplan/tenkai-api/pkg/util"
	"net/http"
	"strconv"
	"strings"
)

func (appContext *AppContext) deployTrafficRule(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

	var payload model.TrafficRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	//Locate Environment
	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(payload.EnvironmentID)
	domain := strings.Replace(payload.Domain, "${NAMESPACE}", environment.Namespace, -1)
	serviceName := strings.Replace(payload.ServiceName, "${NAMESPACE}", environment.Namespace, -1)

	defaultReleaseName := strings.Replace(payload.DefaultReleaseName, "${NAMESPACE}", environment.Namespace, -1)
	headerReleaseName := strings.Replace(payload.HeaderReleaseName, "${NAMESPACE}", environment.Namespace, -1)

	//
	chart := "saj6/tenkai-canary"
	name := "canary-" + serviceName
	out := &bytes.Buffer{}

	kubeConfig := appContext.ConventionInterface.GetKubeConfigFileName(environment.Group, environment.Name)

	variables := make([]string, 1)
	variables = append(variables, "istio.virtualservices.hosts[0]="+domain)

	variables = append(variables, "istio.virtualservices.apiPath="+payload.ContextPath)

	appName := payload.ServiceName[:strings.Index(payload.ServiceName, "-")]

	variables = append(variables, "app.serviceName="+serviceName)
	variables = append(variables, "app.name="+appName)

	if payload.HeaderName != "" {
		variables = append(variables, "app.headerEnabled=true")
		variables = append(variables, "app.weightEnabled=false")
		variables = append(variables, "app.defaultReleaseName="+defaultReleaseName)
		variables = append(variables, "app.headerReleaseName="+headerReleaseName)
		variables = append(variables, "app.headers[0].name="+payload.HeaderName)
		variables = append(variables, "app.headers[0].value="+payload.HeaderValue)
	} else {
		variables = append(variables, "app.headerEnabled=false")
		variables = append(variables, "app.weightEnabled=true")
		for i, e := range payload.Releases {
			name := strings.Replace(e.Name, "${NAMESPACE}", environment.Namespace, -1)
			variables = append(variables, "app.releases["+strconv.Itoa(i)+"].name="+name)
			variables = append(variables, "app.releases["+strconv.Itoa(i)+"].value="+strconv.Itoa(e.Weight))
		}
	}

	//Retry 2 times (First time will fail because service already exists).
	//TODO - VERIFY HOW TO FIX IT

	upgradeRequest := helmapi.UpgradeRequest{}
	upgradeRequest.Kubeconfig = kubeConfig
	upgradeRequest.Namespace = environment.Namespace
	upgradeRequest.ChartVersion = ""
	upgradeRequest.Chart = chart
	upgradeRequest.Variables = variables
	upgradeRequest.Dryrun = false
	upgradeRequest.Release = name

	err = appContext.HelmServiceAPI.Upgrade(upgradeRequest, out)
	if err != nil {
		err = appContext.HelmServiceAPI.Upgrade(upgradeRequest, out)
		if err != nil {
			http.Error(w, err.Error(), 501)
			return
		}
	}

	w.WriteHeader(http.StatusOK)

}
