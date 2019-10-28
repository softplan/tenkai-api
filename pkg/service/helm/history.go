package helmapi

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/gosuri/uitable"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/timeconv"
)

//ReleaseInfo ReleaseInfo
type ReleaseInfo struct {
	Revision    int32  `json:"revision"`
	Updated     string `json:"updated"`
	Status      string `json:"status"`
	Chart       string `json:"chart"`
	Description string `json:"description"`
}

//ReleaseHistory Structure
type ReleaseHistory []ReleaseInfo

type historyCmd struct {
	max          int32
	rls          string
	out          io.Writer
	helmc        helm.Interface
	colWidth     uint
	outputFormat string
}

//IsThereAnyPodWithThisVersion - Verify if is there a pod with a specific version deployed
func (svc HelmServiceImpl) IsThereAnyPodWithThisVersion(kubeconfig string, namespace string, releaseName string, tag string) (bool, error) {

	_, client, err := svc.GetHelmConnection().GetKubeClient("", kubeconfig)
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
func (svc HelmServiceImpl) GetReleaseHistory(kubeconfig string, releaseName string) (bool, error) {

	tillerHost, tunnel, err := svc.GetHelmConnection().SetupConnection(kubeconfig)
	defer svc.GetHelmConnection().Teardown(tunnel)

	deployed := false
	if err == nil {
		his := &historyCmd{out: os.Stdout, helmc: svc.GetHelmConnection().NewClient(tillerHost)}

		his.rls = releaseName
		his.max = 1
		deployed, err = his.verifyItDeployed()
	}
	return deployed, err
}

//GetHelmReleaseHistory - Get helm release history
func (svc HelmServiceImpl) GetHelmReleaseHistory(kubeconfig string, releaseName string) (ReleaseHistory, error) {

	var result ReleaseHistory
	tillerHost, tunnel, err := svc.GetHelmConnection().SetupConnection(kubeconfig)
	defer svc.GetHelmConnection().Teardown(tunnel)

	if err == nil {
		his := &historyCmd{out: os.Stdout, helmc: svc.GetHelmConnection().NewClient(tillerHost)}
		his.rls = releaseName
		r, err := his.helmc.ReleaseHistory(his.rls, helm.WithMaxHistory(256))
		if err != nil {
			return nil, prettyError(err)
		}
		if len(r.Releases) == 0 {
			return nil, nil
		}
		result = getReleaseHistory(r.Releases)
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

	releaseList := getReleaseHistory(r.Releases)

	for i := 0; i <= len(releaseList)-1; i++ {
		r := releaseList[i]
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

func getReleaseHistory(rls []*release.Release) (history ReleaseHistory) {
	for i := len(rls) - 1; i >= 0; i-- {
		r := rls[i]
		c := formatChartname(r.Chart)
		t := timeconv.String(r.Info.LastDeployed)
		s := r.Info.Status.Code.String()
		v := r.Version
		d := r.Info.Description

		rInfo := ReleaseInfo{
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

func formatAsTable(releases ReleaseHistory, colWidth uint) []byte {
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
