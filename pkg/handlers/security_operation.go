package handlers

import (
	"encoding/json"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/global"
	"github.com/softplan/tenkai-api/pkg/util"
	"net/http"
)

func (appContext *AppContext) listSecurityOperation(w http.ResponseWriter, r *http.Request) {

	result := &model.SecurityOperationResponse{}
	var err error
	if result.List, err = appContext.Repositories.SecurityOperationDAO.List(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}


func (appContext *AppContext) createOrUpdateSecurityOperation(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)

	var payload model.SecurityOperation

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := appContext.Repositories.SecurityOperationDAO.CreateOrUpdate(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}