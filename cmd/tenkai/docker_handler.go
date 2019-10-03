package main

import (
	"encoding/json"
	"log"
	"github.com/softplan/tenkai-api/dbms"
	"github.com/softplan/tenkai-api/dbms/model"
	dockerapi "github.com/softplan/tenkai-api/service/docker"
	"github.com/softplan/tenkai-api/util"
	"net/http"
	"sort"
	"strings"
)

func (appContext *appContext) listDockerTags(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var payload model.ListDockerTagsRequest

	if err := util.UnmarshalPayload(r, &payload); err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	ds := dockerapi.GetDockerService(appContext.testMode)

	repo, err := getBaseDomainFromRepo(&appContext.database, payload.ImageName)

	tagResult, err := ds.GetTags(repo, payload.ImageName)
	if err != nil {
		http.Error(w, err.Error(), 501)
		return
	}

	result := &model.ListDockerTagsResult{}

	cacheErr := cacheDockerTags(tagResult.Tags, payload.ImageName, appContext, result, ds, repo)
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
	repo *model.DockerRepo) error {

	for _, tag := range tags {
		img := imageName + ":" + tag

		if _, exists := appContext.dockerTagsCache[img]; exists {
			log.Printf("Value %s already exists.", img)
			result.TagResponse = append(result.TagResponse, model.TagResponse{Tag: tag, Created: appContext.dockerTagsCache[img]})
			continue
		} else {
			date, err := ds.GetDate(*repo, imageName, tag)
			if err != nil {
				return err
			}
			result.TagResponse = append(result.TagResponse, model.TagResponse{Tag: tag, Created: *date})
			appContext.dockerTagsCache[img] = *date
			log.Printf("Value %s added.", img)
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
