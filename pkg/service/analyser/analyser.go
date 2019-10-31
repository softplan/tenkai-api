package analyser

import (
	"fmt"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"github.com/softplan/tenkai-api/pkg/dbms/repository"
	"github.com/softplan/tenkai-api/pkg/global"
	helmapi2 "github.com/softplan/tenkai-api/pkg/service/_helm"
	"strings"
)

//Analyse - Analyse dependencies
func Analyse(dao repository.EnvironmentDAOInterface, hsi helmapi2.HelmServiceInterface, dai repository.DependencyDAOInterface, payload model.DepAnalyseRequest, analyse *model.DepAnalyse) error {
	innerAnalyse(dai, "", payload.ChartName, payload.Tag, analyse)
	err := analyseIfDeployed(dao, hsi, payload, analyse)
	if err != nil {
		return err
	}
	return nil
}

func containsByID(nodes []model.Node, ID string) bool {
	result := false
	for _, node := range nodes {
		if node.ID == ID {
			result = true
			break
		}
	}
	return result
}

func innerAnalyse(dai repository.DependencyDAOInterface, parent string, chartName string, tag string, analyse *model.DepAnalyse) {

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

	dependencies, err := dai.GetDependencies(chartName, tag)
	if err != nil {
		fmt.Println("error")
	}

	for _, element := range dependencies {
		matchedDependencies := getMatchedVersions(element.ChartName, element.Version)
		for _, matched := range matchedDependencies {

			if !containsByID(analyse.Nodes, getNodeName(element.ChartName, element.Version)) {
				innerAnalyse(dai, getNodeName(chartName, tag), matched.ChartName, matched.Tag, analyse)
			}
		}
	}

}

func getNodeName(chartName string, version string) string {
	return chartName + ":" + version
}

func getMatchedVersions(chartName string, tag string) []model.DepAnalyseRequest {
	var result []model.DepAnalyseRequest
	result = make([]model.DepAnalyseRequest, 0)
	element := &model.DepAnalyseRequest{ChartName: chartName, Tag: tag}
	result = append(result, *element)
	return result
}

func analyseIfDeployed(dao repository.EnvironmentDAOInterface, hsi helmapi2.HelmServiceInterface, payload model.DepAnalyseRequest, analyse *model.DepAnalyse) error {

	//Find environment
	environment, _ := dao.GetByID(payload.EnvironmentID)

	for index, element := range analyse.Nodes {
		releaseName := removeTag(removeRepo(element.ID)) + "-" + environment.Namespace

		kubeConfig := global.KubeConfigBasePath + environment.Group + "_" + environment.Name
		err := identifyDeployedReleased(kubeConfig, hsi, analyse, environment.Namespace, releaseName, onlyTag(removeRepo(element.ID)), index)
		if err != nil {
			return err
		}
	}
	return nil

}

func identifyDeployedReleased(kubeconfig string, hsi helmapi2.HelmServiceInterface, analyse *model.DepAnalyse, namespace, releaseName string, tag string, index int) error {
	deployed, err := hsi.GetReleaseHistory(kubeconfig, releaseName)
	if err != nil {
		deployed = false
	}

	if !deployed {
		analyse.Nodes[index].Svg = "data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyIgZmlsbC1ydWxlPSJldmVub2RkIiBjbGlwLXJ1bGU9ImV2ZW5vZGQiPjxwYXRoIGQ9Ik0xMyA5aDlsLTE0IDE1IDMtOWgtOWwxNC0xNS0zIDl6bS04LjY5OSA1aDguMDg2bC0xLjk4NyA1Ljk2MyA5LjI5OS05Ljk2M2gtOC4wODZsMS45ODctNS45NjMtOS4yOTkgOS45NjN6Ii8+PC9zdmc+"
	} else {
		//Verify if version is OK.
		versionMatched, err := hsi.IsThereAnyPodWithThisVersion(kubeconfig, namespace, releaseName, tag)
		if err != nil {
			return err
		}
		if !versionMatched {
			analyse.Nodes[index].SymbolType = "triangle"
		} else {
			analyse.Nodes[index].SymbolType = "circle"
			analyse.Nodes[index].Color = "green"
		}
	}
	return nil
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
