package analyser

import (
	"encoding/json"
	helmapi "github.com/softplan/tenkai-api/service/helm"
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
func GetImageFromService(serviceName string) (string, error) {

	//Look at the chart
	bytes, _ := helmapi.GetValues(serviceName, "0")

	var data JSONObject
	if err := json.Unmarshal(bytes, &data); err != nil {
		return "", err
	}

	return data.Image.Repository, nil

}
