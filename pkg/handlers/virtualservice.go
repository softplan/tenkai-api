package handlers

import (
	"encoding/json"
	"errors"
	"github.com/softplan/tenkai-api/pkg/util"
	"net/http"
	"sort"
	"strconv"
)

func (appContext *AppContext) handleEnvironment(r *http.Request) (string, string, error) {

	principal := util.GetPrincipal(r)

	environmentIDs, ok := r.URL.Query()["environmentID"]
	if !ok || len(environmentIDs[0]) < 1 {
		return "", "", errors.New("param environmentID is required")
	}

	//Locate Environment
	envID, _ := strconv.Atoi(environmentIDs[0])

	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(int(envID))
	if err != nil {
		return "", "", err
	}

	has, err := appContext.hasAccess(principal.Email, envID)
	if err != nil || !has {
		return "", "", errors.New("Access Denied in this environment")
	}

	kubeConfig := appContext.ConventionInterface.GetKubeConfigFileName(environment.Group, environment.Name)

	return kubeConfig, environment.Name, nil

}

func (appContext *AppContext) getVirtualServices(w http.ResponseWriter, r *http.Request) {

	kubeConfig, name, err := appContext.handleEnvironment(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	services, err := appContext.HelmServiceAPI.GetVirtualServices(kubeConfig, name)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	sort.Slice(services, func(i, j int) bool {
		return services[i] < (services[j])
	})

	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(services)
	w.Write(data)

}
