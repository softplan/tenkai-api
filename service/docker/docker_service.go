package dockerapi

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"github.com/softplan/tenkai-api/dbms/model"
	"net/http"
	"sort"
	"strings"
	"time"
)

func getImageWithoutRepo(image string) string {
	result := ""
	index := strings.Index(image, "/")
	result = image[index+1:]
	return result
}

func getHTTPClient() *http.Client {
	tr := &http.Transport{}
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	return &http.Client{Transport: tr}
}

// GetDockerService is a builder to return DockerService or DockerMockService based on appContext.testMode value
func GetDockerService(testMode bool) DockerServiceInterface {
	var result DockerServiceInterface
	if testMode {
		result = &DockerMockService{}
	} else {
		result = &DockerService{}
	}
	return result
}

// DockerServiceInterface can be used to interact with a remote Docker repo
type DockerServiceInterface interface {
	GetDate(repo model.DockerRepo, imageName string, tag string) (*time.Time, error)
	GetTags(repo *model.DockerRepo, imageName string) (*model.TagsResult, error)
	GetDateCalledTimes() int
}

// MOCK IMPLEMENTATION

// DockerMockService is used to concretize DockerServiceInterface for test only
type DockerMockService struct {
	dateCalledTimes int
}

// GetDate returns a date using time.Now for test only
func (docker *DockerMockService) GetDate(repo model.DockerRepo, imageName string, tag string) (*time.Time, error) {
	list := make([]model.V1Compatibility, 0)
	list = append(list, model.V1Compatibility{Created: time.Now()})
	docker.dateCalledTimes = docker.dateCalledTimes + 1
	return &list[len(list)-1].Created, nil
}

// GetTags is a mock function for test only
func (docker *DockerMockService) GetTags(repo *model.DockerRepo, imageName string) (*model.TagsResult, error) {
	return nil, nil
}

// GetDateCalledTimes returns number of times the func was called.
func (docker *DockerMockService) GetDateCalledTimes() int {
	return docker.dateCalledTimes
}

// REAL IMPLEMENTATION

// DockerService is used to concretize DockerServiceInterface
type DockerService struct {
}

// GetDate fetches the docker image creation from a remote repo
func (docker *DockerService) GetDate(repo model.DockerRepo, imageName string, tag string) (*time.Time, error) {

	imageName = getImageWithoutRepo(imageName)

	url := "https://" + repo.Host + "/v2/" + imageName + "/manifests/" + tag

	client := getHTTPClient()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	sEnc := base64.StdEncoding.EncodeToString([]byte(repo.Username + ":" + repo.Password))
	req.Header.Add("Authorization", " Basic "+sEnc)

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	d := json.NewDecoder(resp.Body)
	var manifestResult model.ManifestResult

	err = d.Decode(&manifestResult)
	if err != nil {
		return nil, err
	}

	list := make([]model.V1Compatibility, 0)

	for _, e := range manifestResult.History {
		v1Compatibility := &model.V1Compatibility{}
		json.Unmarshal([]byte(e.V1Compatibility), &v1Compatibility)
		list = append(list, *v1Compatibility)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Created.Before(list[j].Created)
	})

	return &list[len(list)-1].Created, nil

}

// GetTags fetches the docker tags from a remote repo
func (docker *DockerService) GetTags(repo *model.DockerRepo, imageName string) (*model.TagsResult, error) {

	url := "https://" + repo.Host + "/v2/" + getImageWithoutRepo(imageName) + "/tags/list"

	client := getHTTPClient()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	sEnc := base64.StdEncoding.EncodeToString([]byte(repo.Username + ":" + repo.Password))

	req.Header.Add("Authorization", " Basic "+sEnc)

	// Fetch Request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	d := json.NewDecoder(resp.Body)
	var tagResult model.TagsResult

	err = d.Decode(&tagResult)
	if err != nil {
		return nil, err
	}
	return &tagResult, nil

}

// GetDateCalledTimes returns number of times the func was called.
func (docker *DockerService) GetDateCalledTimes() int {
	return 0
}
