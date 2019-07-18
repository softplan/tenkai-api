package service_tenkai

import (
	"fmt"
	"github.com/softplan/tenkai-api/dbms"
	"github.com/softplan/tenkai-api/dbms/model"
	"github.com/softplan/tenkai-api/global"
	helmapi "github.com/softplan/tenkai-api/service/helm"
	"strings"
)

func Analyse(database dbms.Database, payload model.DepAnalyseRequest, analyse *model.DepAnalyse) {
	innerAnalyse(database, "", payload.ChartName, payload.Tag, analyse)
	analyseIfDeployed(database, payload, analyse)
}

func innerAnalyse(database dbms.Database, parent string, chartName string, tag string, analyse *model.DepAnalyse) {

	if analyse.Nodes == nil {
		analyse.Nodes = make([]model.Node, 0)
	}

	if len(parent) > 0 && analyse.Links == nil {
		analyse.Links = make([]model.DepLink, 0)
	}

	analyse.Nodes = append(analyse.Nodes, model.Node{ID: getNodeName(chartName, tag), Color: "blue", SymbolType: "circle"})
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
			innerAnalyse(database, getNodeName(chartName, tag), matched.ChartName, matched.Tag, analyse)
		}
	}

}

func getParentName(dependency model.Dependency) string {
	return getNodeName(dependency.ChartName, dependency.Version)
}

func getNodeName(chartName string, version string) string {
	return chartName + ":" + version
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

func analyseIfDeployed(database dbms.Database, payload model.DepAnalyseRequest, analyse *model.DepAnalyse) {

	//Find environment
	environment, _ := database.GetByID(payload.EnvironmentID)

	for index, element := range analyse.Nodes {
		releaseName := removeTag(removeRepo(element.ID)) + "-" + environment.Namespace

		kubeConfig := global.KUBECONFIG_BASE_PATH + environment.Group + "_" + environment.Name
		identifyDeployedReleased(kubeConfig, analyse, environment.Namespace, releaseName, onlyTag(removeRepo(element.ID)), index)
	}

}

func identifyDeployedReleased(kubeconfig string, analyse *model.DepAnalyse, namespace, releaseName string, tag string, index int) {
	deployed, _ := helmapi.GetReleaseHistory(kubeconfig, releaseName)
	if !deployed {
		analyse.Nodes[index].Svg = "https://dev.w3.org/SVG/tools/svgweb/samples/svg-files/no.svg"
	} else {
		//Verify if version is OK.
		versionMatched, _ := helmapi.IsThereAnyPodWithThisVersion(kubeconfig, namespace, releaseName, tag)
		if !versionMatched {
			analyse.Nodes[index].SymbolType = "triangle"
		} else {
			analyse.Nodes[index].SymbolType = "circle"
			analyse.Nodes[index].Color = "green"
		}
	}
}

func onlyTag(value string) string {
	return value[strings.Index(value, ":")+1:]
}

func removeTag(value string) string {
	return value[0:strings.Index(value, ":")]
}

func removeRepo(value string) string {
	return value[strings.Index(value, "/")+1:]
}
