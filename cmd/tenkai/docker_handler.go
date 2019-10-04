package main

import (
	"encoding/json"
	"github.com/softplan/tenkai-api/dbms"
	"github.com/softplan/tenkai-api/dbms/model"
	dockerapi "github.com/softplan/tenkai-api/service/docker"
	"github.com/softplan/tenkai-api/util"
	"net/http"
	"sort"
	"strings"
	"time"
)

func (appContext *appContext) listDockerTags(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.ListDockerTagsRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}
	var dateFrom time.Time
	matchFromDate := false
	if payload.From != "" {
		layout := "2006-01-02"
		dateFrom, _ = time.Parse(layout, payload.From)
		matchFromDate = true

	}

	ds := dockerapi.GetDockerService(appContext.testMode)

	repo, err := getBaseDomainFromRepo(&appContext.database, payload.ImageName)

	tagResult, err := ds.GetTags(repo, payload.ImageName)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	result := &model.ListDockerTagsResult{}

	cacheErr := cacheDockerTags(tagResult.Tags, payload.ImageName, appContext, result, ds, repo, matchFromDate, dateFrom)
	if cacheErr != nil {
		http.Error(w, cacheErr.Error(), 501)
	}

	sort.Slice(result.TagResponse, func(i, j int) bool {
		return result.TagResponse[i].Created.Before(result.TagResponse[j].Created)
	})

	data, _ := json.Marshal(result)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func cacheDockerTags(tags []string, imageName string,
	appContext *appContext, result *model.ListDockerTagsResult, ds dockerapi.DockerServiceInterface,
	repo *model.DockerRepo, matchFromDate bool, dateFrom time.Time) error {

	for _, tag := range tags {
		img := imageName + ":" + tag

		if _, exists := appContext.dockerTagsCache[img]; exists {
			if matchFromDate {
				if appContext.dockerTagsCache[img].After(dateFrom) {
					result.TagResponse = append(result.TagResponse, model.TagResponse{Tag: tag, Created: appContext.dockerTagsCache[img]})
				}
			} else {
				result.TagResponse = append(result.TagResponse, model.TagResponse{Tag: tag, Created: appContext.dockerTagsCache[img]})
			}
		} else {
			date, err := ds.GetDate(*repo, imageName, tag)
			if err != nil {
				return err
			}
			if matchFromDate {
				if date.After(dateFrom) {
					result.TagResponse = append(result.TagResponse, model.TagResponse{Tag: tag, Created: *date})
				}
			} else {
				result.TagResponse = append(result.TagResponse, model.TagResponse{Tag: tag, Created: *date})
			}
			appContext.dockerTagsCache[img] = *date
		}
	}
	return nil
}

func getBaseDomainFromRepo(dbms *dbms.Database, imageName string) (*model.DockerRepo, error) {
	firstBarIndex := strings.Index(imageName, "/")
	host := imageName[0:firstBarIndex]
	result, err := dbms.GetDockerRepositoryByHost(host)
	return &result, err
}
