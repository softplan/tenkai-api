package handlers

import (
	"encoding/json"
	"net/http"
)

func (appContext *AppContext) getVirtualServices(w http.ResponseWriter, r *http.Request) {

	kubeConfig := appContext.ConventionInterface.GetKubeConfigFileName("unj", "master")

	services, err := appContext.HelmServiceAPI.GetVirtualServices(kubeConfig, "master")
	if err != nil {
		http.Error(w, err.Error(), 501)
		return

	}

	w.WriteHeader(http.StatusOK)
	data, _ := json.Marshal(services)
	w.Write(data)

}
