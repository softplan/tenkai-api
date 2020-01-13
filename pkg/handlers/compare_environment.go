package handlers

import (
	"fmt"
	"net/http"

	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/global"
	"github.com/softplan/tenkai-api/pkg/util"
)

func (appContext *AppContext) compareEnvironments(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(global.ContentType, global.JSONContentType)

	var payload model.CompareEnvironments

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var err error
	var sourceVars []model.Variable
	if sourceVars, err = appContext.Repositories.VariableDAO.
		GetAllVariablesByEnvironment(payload.SourceEnvID); err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var targetVars []model.Variable
	if targetVars, err = appContext.Repositories.VariableDAO.
		GetAllVariablesByEnvironment(payload.TargetEnvID); err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var resp model.CompareEnvsResponse
	appContext.compare(&resp, toMap(sourceVars), toMap(targetVars))
	appContext.compare(&resp, toMap(targetVars), toMap(sourceVars))
}

func (appContext *AppContext) compare(
	resp *model.CompareEnvsResponse,
	source map[string]map[string]string,
	target map[string]map[string]string) {

	for scope, varMap := range source {
		if target[scope] != nil {

			for sk, sv := range varMap {
				for tk, tv := range target[scope] {
					if sk == tk {
						if sv == tv {
							continue
						} else {
							var x model.EnvironmentsDiff
							x.SourceScope = scope
							x.TargetScope = scope
							x.SourceName = sk
							x.TargetName = tk
							x.SourceValue = sv
							x.TargetValue = tv
							resp.List = append(resp.List, x)
						}
					} else {
						continue
					}
				}

			}

		} else {
			fmt.Println("diff: ", scope, nil, nil, nil, nil)
		}
	}
}

func toMap(vars []model.Variable) map[string]map[string]string {
	sm := make(map[string]map[string]string)

	for _, e := range vars {
		if sm[e.Scope] == nil {
			sm[e.Scope] = map[string]string{e.Name: e.Value}
		} else {
			sm[e.Scope][e.Name] = e.Value
		}
	}

	return sm
}
