package handlers

import (
	"encoding/json"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/global"
	"github.com/softplan/tenkai-api/pkg/util"
	"net/http"
)

func (appContext *AppContext) createOrUpdateUserEnvironmentRole(w http.ResponseWriter, r *http.Request) {

	w.Header().Set(global.ContentType, global.JSONContentType)
	var payload model.UserEnvironmentRole

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := appContext.Repositories.UserEnvironmentRoleDAO.CreateOrUpdate(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

func (appContext *AppContext) getUserPolicyByEnvironment(w http.ResponseWriter, r *http.Request) {

	var payload model.GetUserPolicyByEnvironmentRequest
	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var user model.User
	var err error
	if user, err = appContext.Repositories.UserDAO.FindByEmail(payload.Email); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := &model.SecurityOperation{}
	if result, err = appContext.Repositories.UserEnvironmentRoleDAO.GetRoleByUserAndEnvironment(user, uint(payload.EnvironmentID)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)


}