package service_tenkai

import (
	"fmt"
	"github.com/softplan/tenkai-api/dbms"
	"github.com/softplan/tenkai-api/dbms/model"
)

func Analyse(database dbms.Database, parent string, chartName string, tag string, analyse *model.DepAnalyse) {

	if analyse.Nodes == nil {
		analyse.Nodes = make([]string, 0)
	}

	if len(parent) > 0 && analyse.Links == nil {
		analyse.Links = make([]model.DepLink, 0)
	}

	analyse.Nodes = append(analyse.Nodes, chartName)
	if len(parent) > 0 {
		analyse.Links = append(analyse.Links, model.DepLink{Source: parent, Target: getNodeName(chartName, tag)})
	}

	dependencies, err := database.GetDependencies(chartName, tag)
	if err != nil {
		fmt.Println("error")
	}

	for _, element := range dependencies {
		matchedDependencies := getMatchedVersions(element.ChartName, element.Version)
		for _, matched := range matchedDependencies {
			Analyse(database, getNodeName(chartName, tag), matched.ChartName,matched.Tag, analyse)
		}
	}

}

func getParentName(dependency model.Dependency) string {
	return getNodeName(dependency .ChartName, dependency .Version)
}

func getNodeName(chartName string, version string) string {
	return chartName + "-" + version
}

func getMatchedVersions(chartName string, tag string) []model.DepAnalyseRequest {

	var result []model.DepAnalyseRequest
	result = make([]model.DepAnalyseRequest, 0)

	element := &model.DepAnalyseRequest{ChartName: chartName, Tag: tag}
	result = append(result, *element)


	//TODO: https://github.com/coreos/go-semver
	//Here we must match using version semantics
	//https://semver.npmjs.com/

	return result


}