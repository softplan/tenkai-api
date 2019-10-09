package dockerapi

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"github.com/softplan/tenkai-api/dbms"
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

func getBaseDomainFromRepo(dbms *dbms.Database, imageName string) (*model.DockerRepo, error) {
	firstBarIndex := strings.Index(imageName, "/")
	host := imageName[0:firstBarIndex]
	result, err := dbms.GetDockerRepositoryByHost(host)
	return &result, err
}

func cacheDockerTags(tags []string, imageName string, result *model.ListDockerTagsResult, ds DockerServiceInterface,
	repo *model.DockerRepo, matchFromDate bool, dateFrom time.Time, globalCache map[string]time.Time) error {

	for _, tag := range tags {
		img := imageName + ":" + tag

		if _, exists := globalCache[img]; exists {
			if matchFromDate {
				if globalCache[img].After(dateFrom) {
					result.TagResponse = append(result.TagResponse, model.TagResponse{Tag: tag, Created: globalCache[img]})
				}
			} else {
				result.TagResponse = append(result.TagResponse, model.TagResponse{Tag: tag, Created: globalCache[img]})
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
			globalCache[img] = *date
		}
	}
	return nil
}

//GetDockerTagsWithDate Method
func GetDockerTagsWithDate(payload model.ListDockerTagsRequest, testMode bool,
	dbms dbms.Database, globalCache map[string]time.Time) (*model.ListDockerTagsResult, error) {

	var dateFrom time.Time
	matchFromDate := false
	if payload.From != "" {
		layout := "2006-01-02"
		dateFrom, _ = time.Parse(layout, payload.From)
		matchFromDate = true

	}

	ds := GetDockerService(testMode)

	repo, err := getBaseDomainFromRepo(&dbms, payload.ImageName)

	tagResult, err := ds.GetTags(repo, payload.ImageName)
	if err != nil {
		return nil, err
	}

	result := &model.ListDockerTagsResult{}

	cacheErr := cacheDockerTags(tagResult.Tags, payload.ImageName, result, ds, repo, matchFromDate, dateFrom, globalCache)
	if cacheErr != nil {
		return nil, cacheErr
	}

	sort.Slice(result.TagResponse, func(i, j int) bool {
		return result.TagResponse[i].Created.Before(result.TagResponse[j].Created)
	})

	return result, nil

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
