package helmapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"os"
)

var errReleaseRequired = errors.New("release name is required")

type getValuesCmd struct {
	release   string
	allValues bool
	out       io.Writer
	client    helm.Interface
	version   int32
	output    string
}

//Get - All
func (svc HelmServiceImpl) Get(kubeconfig string, releaseName string, revision int) (string, error) {

	tillerHost, tunnel, err := svc.GetHelmConnection().SetupConnection(kubeconfig)
	defer svc.GetHelmConnection().Teardown(tunnel)
	if err != nil {
		return "", err
	}

	cmd := &getValuesCmd{out: os.Stdout}
	cmd.allValues = false
	cmd.release = releaseName
	cmd.version = int32(revision)
	cmd.client = svc.GetHelmConnection().NewClient(tillerHost)

	res, err := cmd.client.ReleaseContent(cmd.release, helm.ContentReleaseVersion(cmd.version))
	if err != nil {
		return "", err
	}

	values, err := chartutil.ReadValues([]byte(res.Release.Config.Raw))
	if err != nil {
		return "", err
	}

	result, err := formatValues(cmd.output, values)
	if err != nil {
		return "", err
	}
	return result, nil
}

func formatValues(format string, values chartutil.Values) (string, error) {
	switch format {
	case "", "yaml":
		out, err := values.YAML()
		if err != nil {
			return "", err
		}
		return out, nil
	case "json":
		out, err := json.Marshal(values)
		if err != nil {
			return "", fmt.Errorf("Failed to Marshal JSON output: %s", err)
		}
		return string(out), nil
	default:
		return "", fmt.Errorf("Unknown output format %q", format)
	}
}
