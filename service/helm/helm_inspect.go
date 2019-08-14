package helmapi

import (
	"errors"
	"fmt"
	"github.com/helm/helm/pkg/downloader"
	"github.com/softplan/tenkai-api/global"
	"io"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/repo"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
	"sync"
)

type inspectCmd struct {
	chartpath string
	verify    bool
	keyring   string
	out       io.Writer
	version   string
	repoURL   string
	username  string
	password  string
	devel     bool

	certFile string
	keyFile  string
	caFile   string
}

//GetTemplate - Retrieve a chart template called deployment.yaml
func GetTemplate(mutex *sync.Mutex, chartName string, version string, kind string) ([]byte, error) {

	var result []byte
	var err error
	mutex.Lock()
	if kind == "values" {
		result, err = GetValues(chartName, version)
	} else {
		if kind == "deployment" {
			result, err = GetDeployment(chartName, version)
		}
	}
	mutex.Unlock()
	return result, err
}

//GetDeployment - Retrieve a deployment chart template
func GetDeployment(chartName string, version string) ([]byte, error) {

	logFields := global.AppFields{global.Function: "GetValues"}

	insp := &inspectCmd{
		out: os.Stdout,
	}

	if len(version) > 0 {
		insp.version = version
	}

	settings.Home = global.HelmDir

	global.Logger.Info(logFields, "insp.prepare()")
	if err := insp.prepare(chartName); err != nil {
		return nil, err
	}

	global.Logger.Info(logFields, "before insp.run()")
	values, err := insp.runGetDeployment()
	if err != nil {
		global.Logger.Error(logFields, err.Error())
		return nil, err
	}
	var result []byte
	if values != nil {
		result = values.Data
	}
	return result, nil

}

//GetValues Method
func GetValues(chartName string, version string) ([]byte, error) {

	logFields := global.AppFields{global.Function: "GetValues"}

	insp := &inspectCmd{
		out: os.Stdout,
	}

	if len(version) > 0 {
		insp.version = version
	}

	settings.Home = global.HelmDir

	global.Logger.Info(logFields, "insp.prepare()")
	if err := insp.prepare(chartName); err != nil {
		return nil, err
	}

	global.Logger.Info(logFields, "before insp.run()")
	values, err := insp.run()
	if err != nil {
		global.Logger.Error(logFields, err.Error())
		return nil, err
	}

	j2, err := yaml.YAMLToJSON([]byte(values.Raw))

	if err != nil {
		global.Logger.Error(logFields, err.Error())
		return nil, err
	}

	return j2, nil

}

func (i *inspectCmd) prepare(chart string) error {

	if i.version == "" && i.devel {
		i.version = ">0.0.0-0"
	}

	cp, err := locateChartPath(i.repoURL, i.username, i.password, chart, i.version, i.verify, i.keyring,
		i.certFile, i.keyFile, i.caFile)
	if err != nil {
		return err
	}
	i.chartpath = cp

	return nil
}

func (i *inspectCmd) runGetDeployment() (*chart.Template, error) {
	var result *chart.Template

	chrt, err := chartutil.Load(i.chartpath)

	if err != nil {
		return nil, err
	}
	for _, e := range chrt.Templates {
		if e.Name == "templates/deployment.yaml" {
			result = e
			break
		}
	}

	return result, nil
}

func (i *inspectCmd) run() (*chart.Config, error) {
	chrt, err := chartutil.Load(i.chartpath)
	if err != nil {
		return nil, err
	}
	return chrt.Values, nil
}

func locateChartPath(repoURL, username, password, name, version string, verify bool, keyring,
	certFile, keyFile, caFile string) (string, error) {

	logFields := global.AppFields{global.Function: "locateChartPath"}

	name = strings.TrimSpace(name)
	version = strings.TrimSpace(version)
	if fi, err := os.Stat(name); err == nil {
		abs, err := filepath.Abs(name)
		if err != nil {
			return abs, err
		}
		if verify {
			if fi.IsDir() {
				return "", errors.New("cannot verify a directory")
			}
			if _, err := downloader.VerifyChart(abs, keyring); err != nil {
				return "", err
			}
		}
		return abs, nil
	}
	if filepath.IsAbs(name) || strings.HasPrefix(name, ".") {
		return name, fmt.Errorf("path %q not found", name)
	}

	crepo := filepath.Join(settings.Home.Repository(), name)
	if _, err := os.Stat(crepo); err == nil {
		return filepath.Abs(crepo)
	}

	dl := downloader.ChartDownloader{
		HelmHome: settings.Home,
		Out:      os.Stdout,
		Keyring:  keyring,
		Getters:  getter.All(settings),
		Username: username,
		Password: password,
	}
	if verify {
		dl.Verify = downloader.VerifyAlways
	}
	if repoURL != "" {

		global.Logger.Info(global.AppFields{global.Function: "locateChartPath", "repoURL": repoURL}, "before FindChartInAuthRepoURL")

		chartURL, err := repo.FindChartInAuthRepoURL(repoURL, username, password, name, version,
			certFile, keyFile, caFile, getter.All(settings))
		if err != nil {
			global.Logger.Error(logFields, err.Error())
			return "", err
		}
		name = chartURL
	}

	if _, err := os.Stat(settings.Home.Archive()); os.IsNotExist(err) {
		os.MkdirAll(settings.Home.Archive(), 0744)
	}

	global.Logger.Info(global.AppFields{global.Function: "locateChartPath", "name": name, "version": version}, "before DownloadTo")
	filename, _, err := dl.DownloadTo(name, version, settings.Home.Archive())
	if err == nil {
		lname, err := filepath.Abs(filename)
		if err != nil {
			global.Logger.Error(global.AppFields{global.Function: "locateChartPath", "lname": lname, "err": err.Error()}, "Error download")
			return filename, err
		}
		return lname, nil
	} else if settings.Debug {
		return filename, err
	}

	global.Logger.Error(global.AppFields{global.Function: "locateChartPath", "filename": filename, "err": err.Error()}, "Before final return")
	return filename, fmt.Errorf("failed to download %q (hint: running `helm repo update` may help)", name)
}
