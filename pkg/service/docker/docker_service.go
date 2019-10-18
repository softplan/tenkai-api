package dockerapi

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/dbms/repository"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

func DockerServiceBuilder() *DockerService {
	r := &DockerService{}
	return r
}

// DockerService is used to concretize DockerServiceInterface
type DockerService struct {
}

// DockerServiceInterface can be used to interact with a remote Docker repo
type DockerServiceInterface interface {
	GetDockerTagsWithDate(payload model.ListDockerTagsRequest, dao repository.DockerDAOInterface, globalCache *sync.Map) (*model.ListDockerTagsResult, error)
	GetDate(repo model.DockerRepo, imageName string, tag string) (*time.Time, error)
	GetTags(repo *model.DockerRepo, imageName string) (*model.TagsResult, error)
	GetDateCalledTimes() int
}

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

func getBaseDomainFromRepo(dao repository.DockerDAOInterface, imageName string) (*model.DockerRepo, error) {
	firstBarIndex := strings.Index(imageName, "/")
	if firstBarIndex <= 0 {
		return nil, errors.New("Repository expected")
	}
	host := imageName[0:firstBarIndex]
	result, err := dao.GetDockerRepositoryByHost(host)
	return &result, err
}

func cacheDockerTags(tags []string, imageName string, result *model.ListDockerTagsResult, ds DockerServiceInterface,
	repo *model.DockerRepo, matchFromDate bool, dateFrom time.Time, globalCache *sync.Map) error {

	for _, tag := range tags {

		img := imageName + ":" + tag
		createDate, ok := globalCache.Load(img)

		if ok {

			if matchFromDate {
				var object interface{}
				var dateTime time.Time
				object, _ = globalCache.Load(img)
				dateTime = object.(time.Time)
				major := dateTime.After(dateFrom)
				if major {
					result.TagResponse = append(result.TagResponse, model.TagResponse{Tag: tag, Created: createDate.(time.Time)})
				}
			} else {
				result.TagResponse = append(result.TagResponse, model.TagResponse{Tag: tag, Created: createDate.(time.Time)})
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
			globalCache.Store(img, *date)
		}
	}
	return nil
}

//GetDockerTagsWithDate Method
func (docker DockerService) GetDockerTagsWithDate(payload model.ListDockerTagsRequest, dao repository.DockerDAOInterface, globalCache *sync.Map) (*model.ListDockerTagsResult, error) {

	var dateFrom time.Time
	matchFromDate := false
	if payload.From != "" {
		layout := "2006-01-02"
		dateFrom, _ = time.Parse(layout, payload.From)
		matchFromDate = true

	}

	repo, err := getBaseDomainFromRepo(dao, payload.ImageName)
	if err != nil {
		return nil, err
	}

	tagResult, err := docker.GetTags(repo, payload.ImageName)
	if err != nil {
		return nil, err
	}

	result := &model.ListDockerTagsResult{}

	cacheErr := cacheDockerTags(tagResult.Tags, payload.ImageName, result, docker, repo, matchFromDate, dateFrom, globalCache)
	if cacheErr != nil {
		return nil, cacheErr
	}

	sort.Slice(result.TagResponse, func(i, j int) bool {
		return result.TagResponse[i].Created.Before(result.TagResponse[j].Created)
	})

	return result, nil

}

// GetDate fetches the docker image creation from a remote repo
func (docker DockerService) GetDate(repo model.DockerRepo, imageName string, tag string) (*time.Time, error) {

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
func (docker DockerService) GetTags(repo *model.DockerRepo, imageName string) (*model.TagsResult, error) {

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
func (docker DockerService) GetDateCalledTimes() int {
	return 0
}
