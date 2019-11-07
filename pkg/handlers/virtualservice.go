package handlers

import (
	"encoding/json"
	"errors"
	"github.com/softplan/tenkai-api/pkg/util"
	"net/http"
	"strconv"
)

func (appContext *AppContext) getVirtualServices(w http.ResponseWriter, r *http.Request) {

	principal := util.GetPrincipal(r)

	environmentIDs, ok := r.URL.Query()["environmentID"]
	if !ok || len(environmentIDs[0]) < 1 {
		http.Error(w, errors.New("param environmentID is required").Error(), 501)
		return
	}

	//Locate Environment
	envID, _ := strconv.Atoi(environmentIDs[0])

	has, err := appContext.hasAccess(principal.Email, envID)
	if err != nil || !has {
		http.Error(w, errors.New("Access Denied in this environment").Error(), http.StatusUnauthorized)
		return
	}

	environment, err := appContext.Repositories.EnvironmentDAO.GetByID(int(envID))
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	kubeConfig := appContext.ConventionInterface.GetKubeConfigFileName(environment.Group, environment.Name)

	services, err := appContext.HelmServiceAPI.GetVirtualServices(kubeConfig, environment.Name)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return

	}

	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(services)
	w.Write(data)

}
