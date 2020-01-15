package handlers

import (
	"encoding/json"
	"hash/fnv"
	"net/http"
	"sort"

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

	if len(payload.OnlyFields) > 0 && len(payload.ExceptFields) > 0 {
		http.Error(w, "Choose only one kind of filter fields: only or except", http.StatusInternalServerError)
		return
	}

	if len(payload.OnlyCharts) > 0 && len(payload.ExceptCharts) > 0 {
		http.Error(w, "Choose only one kind of filter charts: only or except", http.StatusInternalServerError)
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
	rmap := make(map[uint32]model.EnvironmentsDiff)

	appContext.compare(rmap, payload, toMap(sourceVars), toMap(targetVars), false)
	appContext.compare(rmap, payload, toMap(targetVars), toMap(sourceVars), true)

	for _, v := range rmap {
		resp.List = append(resp.List, v)
	}

	sort.Slice(resp.List, func(i int, j int) bool {
		return resp.List[i].SourceScope < resp.List[j].SourceScope
	})

	data, _ := json.Marshal(resp)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (appContext *AppContext) compare(
	rmap map[uint32]model.EnvironmentsDiff,
	filter model.CompareEnvironments,
	source map[string]map[string]string,
	target map[string]map[string]string, reverse bool) {

	for scope, srcVars := range source {
		if ignore := shouldFilterChart(filter, scope); ignore {
			continue
		}

		if _, ok := target[scope]; ok { // Target possui este chart?

			for srcVarName, srcValue := range srcVars {
				if ignore := shouldFilterVar(filter, srcVarName); ignore {
					continue
				}
				if _, ok := target[scope][srcVarName]; !ok { // Chart target não possui esta variável?
					addToResp(rmap, filter.SourceEnvID, filter.TargetEnvID, scope, scope, srcVarName, "", srcValue, "", reverse)
					continue
				}
				for tarVarName, tarValue := range target[scope] {
					if ignore := shouldFilterVar(filter, tarVarName); ignore {
						continue
					}
					if srcVarName == tarVarName {
						if srcValue == tarValue {
							continue
						} else {
							addToResp(rmap, filter.SourceEnvID, filter.TargetEnvID, scope, scope, srcVarName, tarVarName, srcValue, tarValue, reverse)
						}
					} else {
						continue
					}
				}

			}

		} else {
			addToResp(rmap, filter.SourceEnvID, filter.TargetEnvID, scope, "", "", "", "", "", reverse)
		}
	}
}

func shouldFilterVar(filter model.CompareEnvironments, varName string) bool {
	if len(filter.ExceptFields) > 0 {
		for _, e := range filter.ExceptFields {
			if e == varName {
				return true
			}
		}
		return false
	}

	if len(filter.OnlyFields) > 0 {
		found := false
		for _, e := range filter.OnlyFields {
			if e == varName {
				found = true
				continue
			}
		}
		return !found
	}

	return false
}

func shouldFilterChart(filter model.CompareEnvironments, scope string) bool {
	if len(filter.ExceptCharts) > 0 {
		for _, e := range filter.ExceptCharts {
			if e == scope {
				return true
			}
		}
		return false
	}

	if len(filter.OnlyCharts) > 0 {
		found := false
		for _, e := range filter.OnlyCharts {
			if e == scope {
				found = true
				continue
			}
		}
		return !found
	}

	return false
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

func addToResp(rmap map[uint32]model.EnvironmentsDiff, srcEnvID int,
	tarEnvID int, srcScope string, tarScope string, srcVarName string,
	tarVarName string, srcValue string, tarValue string, reverse bool) {

	var e model.EnvironmentsDiff
	e.SourceEnvID = srcEnvID
	e.TargetEnvID = tarEnvID

	if reverse {
		e.SourceScope = tarScope
		e.TargetScope = srcScope
		e.SourceName = tarVarName
		e.TargetName = srcVarName
		e.SourceValue = tarValue
		e.TargetValue = srcValue
	} else {
		e.SourceScope = srcScope
		e.TargetScope = tarScope
		e.SourceName = srcVarName
		e.TargetName = tarVarName
		e.SourceValue = srcValue
		e.TargetValue = tarValue
	}

	// Avoid duplicate entries
	h := hash(e)
	if _, ok := rmap[h]; !ok {
		rmap[h] = e
	}
}

func hash(e model.EnvironmentsDiff) uint32 {
	key := e.SourceScope + e.TargetScope + e.SourceName +
		e.TargetName + e.SourceValue + e.TargetValue

	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}
