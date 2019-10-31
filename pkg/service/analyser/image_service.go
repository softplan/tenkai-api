package analyser

import (
	"encoding/json"
	helmapi2 "github.com/softplan/tenkai-api/pkg/service/_helm"
	"sync"
)

//Image object
type Image struct {
	Repository string `json:"repository"`
}

//JSONObject Parent Object
type JSONObject struct {
	Image Image `json:"image"`
}

//GetImageFromService Retrieve Image from Servic eChart
func GetImageFromService(hsi helmapi2.HelmServiceInterface, serviceName string, mutex *sync.Mutex) (string, error) {

	//Look at the chart
	bytes, err := hsi.GetValues(serviceName, "0")
	if err != nil {
		return "", err
	}

	var data JSONObject
	if err := json.Unmarshal(bytes, &data); err != nil {
		return "", err
	}

	return data.Image.Repository, nil

}
