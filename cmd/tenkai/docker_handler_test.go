package main

import (
	"github.com/softplan/tenkai-api/dbms/model"
	dockerapi "github.com/softplan/tenkai-api/service/docker"
	"testing"
)

func TestCacheDockerTags(t *testing.T) {
	tags := []string{"0.1.0", "0.2.0"}
	imageName := "myrepo.com/my-docker-image"
	appContext := GetAppContext()
	result := &model.ListDockerTagsResult{}
	ds := dockerapi.GetDockerService(appContext.testMode)
	repo, _ := getBaseDomainFromRepoMock()

	error := cacheDockerTags(tags, imageName, appContext, result, ds, repo)

	if error != nil {
		t.Fatal(error)
	}

}

func getBaseDomainFromRepoMock() (*model.DockerRepo, error) {
	x := model.DockerRepo{Host: "myrepo.com", Username: "user", Password: "pass"}
	return &x, nil
}
