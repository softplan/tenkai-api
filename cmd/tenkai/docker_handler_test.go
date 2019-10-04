package main

import (
	"github.com/softplan/tenkai-api/dbms/model"
	dockerapi "github.com/softplan/tenkai-api/service/docker"
	"testing"
	"time"
)

func TestCacheDockerTags(t *testing.T) {
	tags := []string{"0.1.0", "0.2.0"}
	imageName := "myrepo.com/my-docker-image"
	appContext := GetAppContext()
	result := &model.ListDockerTagsResult{}
	ds := dockerapi.GetDockerService(appContext.testMode)
	repo, _ := getBaseDomainFromRepoMock()

	dateTime := time.Now()
	error := cacheDockerTags(tags, imageName, appContext, result, ds, repo, false, dateTime)
	if error != nil {
		t.Fatal(error)
	}

	ct := ds.GetDateCalledTimes()
	if ct != 2 {
		t.Errorf("First call to cacheDockerTags: GetDateCalledTimes should be %d, but was %d.", 2, ct)
	}

	// Call cacheDockerTags again to certify if cache works
	error = cacheDockerTags(tags, imageName, appContext, result, ds, repo, false, dateTime)
	if error != nil {
		t.Fatal(error)
	}

	ct = ds.GetDateCalledTimes()
	if ct != 2 {
		t.Errorf("Second call to cacheDockerTags: GetDateCalledTimes should be %d, but was %d.", 2, ct)
	}

	if len(appContext.dockerTagsCache) != 2 {
		t.Errorf("The map appContext.dockerTagsCache should contain %d items, but has %d.", 2, ct)
	}
}

func getBaseDomainFromRepoMock() (*model.DockerRepo, error) {
	x := model.DockerRepo{Host: "myrepo.com", Username: "user", Password: "pass"}
	return &x, nil
}
