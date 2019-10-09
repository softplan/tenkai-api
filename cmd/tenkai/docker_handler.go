package main

import (
	"encoding/json"
	"github.com/softplan/tenkai-api/dbms/model"
	dockerapi "github.com/softplan/tenkai-api/service/docker"
	"github.com/softplan/tenkai-api/util"
	"net/http"
)

func (appContext *appContext) listDockerTags(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.ListDockerTagsRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	result, err := dockerapi.GetDockerTagsWithDate(payload, appContext.testMode, appContext.database, appContext.dockerTagsCache)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
