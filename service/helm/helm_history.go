package helmapi

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	"github.com/softplan/tenkai-api/global"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/timeconv"
	"os"
	"strings"
)

type releaseInfo struct {
	Revision    int32  `json:"revision"`
	Updated     string `json:"updated"`
	Status      string `json:"status"`
	Chart       string `json:"chart"`
	Description string `json:"description"`
}

type releaseHistory []releaseInfo

type historyCmd struct {
	max          int32
	rls          string
	out          io.Writer
	helmc        helm.Interface
	colWidth     uint
	outputFormat string
}

//IsThereAnyPodWithThisVersion - Verify if is there a pod with a specific version deployed
func IsThereAnyPodWithThisVersion(kubeconfig string, namespace string, releaseName string, tag string) (bool, error) {

	_, client, err := getKubeClient(settings.KubeContext, kubeconfig)
	if err != nil {
		return false, err
	}

	deployment, error := client.AppsV1().Deployments(namespace).Get(releaseName, metav1.GetOptions{})
	if error != nil {
		return false, error
	}

	image := deployment.Spec.Template.Spec.Containers[0].Image
	containerTag := image[strings.Index(image, ":")+1:]
	if containerTag != tag {
		return false, nil
	}

	return true, nil

}

//GetReleaseHistory - Retrieve Release History
func GetReleaseHistory(kubeconfig string, releaseName string) (bool, error) {
	settings.KubeConfig = kubeconfig
	settings.Home = global.HelmDir
	settings.TillerNamespace = "kube-system"
	settings.TLSEnable = false
	settings.TLSVerify = false
	settings.TillerConnectionTimeout = 1200
	err := setupConnection()
	deployed := false
	if err == nil {
		his := &historyCmd{out: os.Stdout, helmc: newClient()}
		his.rls = releaseName
		his.max = 1
		deployed, err = his.verifyItDeployed()
		teardown()
		settings.TillerHost = ""
	}
	return deployed, err
}

//GetHelmReleaseHistory - Get helm release history
func GetHelmReleaseHistory(kubeconfig string, releaseName string) (releaseHistory, error) {

	var result releaseHistory

	settings.KubeConfig = kubeconfig
	settings.Home = global.HelmDir
	settings.TillerNamespace = "kube-system"
	settings.TLSEnable = false
	settings.TLSVerify = false
	settings.TillerConnectionTimeout = 1200
	err := setupConnection()
	if err == nil {
		his := &historyCmd{out: os.Stdout, helmc: newClient()}
		his.rls = releaseName

		r, err := his.helmc.ReleaseHistory(his.rls, helm.WithMaxHistory(256))
		if err != nil {
			return nil, prettyError(err)
		}

		if len(r.Releases) == 0 {
			return nil, nil
		}

		result = getReleaseHistory(r.Releases)

		teardown()
		settings.TillerHost = ""
	}
	return result, err
}

func (cmd *historyCmd) verifyItDeployed() (bool, error) {

	r, err := cmd.helmc.ReleaseHistory(cmd.rls, helm.WithMaxHistory(cmd.max))

	if err != nil {
		return false, prettyError(err)
	}

	if len(r.Releases) == 0 {
		return false, nil
	}

	releaseHistory := getReleaseHistory(r.Releases)

	for i := 0; i <= len(releaseHistory)-1; i++ {
		r := releaseHistory[i]
		if r.Status != "DEPLOYED" {
			return false, nil
		}
	}

	return true, nil

}

func (cmd *historyCmd) run() error {

	r, err := cmd.helmc.ReleaseHistory(cmd.rls, helm.WithMaxHistory(cmd.max))

	if err != nil {
		return prettyError(err)
	}
	if len(r.Releases) == 0 {
		return nil
	}

	releaseHistory := getReleaseHistory(r.Releases)

	var history []byte
	var formattingError error

	switch cmd.outputFormat {
	case "yaml":
		history, formattingError = yaml.Marshal(releaseHistory)
	case "json":
		history, formattingError = json.Marshal(releaseHistory)
	case "table":
		history = formatAsTable(releaseHistory, cmd.colWidth)
	default:
		return fmt.Errorf("unknown output format %q", cmd.outputFormat)
	}

	if formattingError != nil {
		return prettyError(formattingError)
	}

	fmt.Fprintln(cmd.out, string(history))
	return nil
}

func getReleaseHistory(rls []*release.Release) (history releaseHistory) {
	for i := len(rls) - 1; i >= 0; i-- {
		r := rls[i]
		c := formatChartname(r.Chart)
		t := timeconv.String(r.Info.LastDeployed)
		s := r.Info.Status.Code.String()
		v := r.Version
		d := r.Info.Description

		rInfo := releaseInfo{
			Revision:    v,
			Updated:     t,
			Status:      s,
			Chart:       c,
			Description: d,
		}
		history = append(history, rInfo)
	}

	return history
}

func formatAsTable(releases releaseHistory, colWidth uint) []byte {
	tbl := uitable.New()

	tbl.MaxColWidth = colWidth
	tbl.AddRow("REVISION", "UPDATED", "STATUS", "CHART", "DESCRIPTION")
	for i := 0; i <= len(releases)-1; i++ {
		r := releases[i]
		tbl.AddRow(r.Revision, r.Updated, r.Status, r.Chart, r.Description)
	}
	return tbl.Bytes()
}

func formatChartname(c *chart.Chart) string {
	if c == nil || c.Metadata == nil {
		// This is an edge case that has happened in prod, though we don't
		// know how: https://github.com/kubernetes/helm/issues/1347
		return "MISSING"
	}
	return fmt.Sprintf("%s-%s", c.Metadata.Name, c.Metadata.Version)
}
