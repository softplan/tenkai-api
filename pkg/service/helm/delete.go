package helmapi

import (
	"fmt"
	"github.com/softplan/tenkai-api/pkg/global"
	"io"
	"os"

	"k8s.io/helm/pkg/helm"
)

type deleteCmd struct {
	name         string
	dryRun       bool
	disableHooks bool
	purge        bool
	timeout      int64
	description  string

	out    io.Writer
	client helm.Interface
}

//DeleteHelmRelease - Delete a Release
func (svc HelmServiceImpl) DeleteHelmRelease(kubeconfig string, releaseName string, purge bool) error {

	logFields := global.AppFields{global.Function: "ListHelmDeployments", releaseName: releaseName}

	settings.KubeConfig = kubeconfig
	settings.Home = global.HelmDir
	settings.TillerNamespace = "kube-system"
	settings.TLSEnable = false
	settings.TLSVerify = false
	settings.TillerConnectionTimeout = 1200

	cmd := &deleteCmd{out: os.Stdout}

	global.Logger.Info(logFields, "setupConnection")
	err := setupConnection()
	defer teardown()

	if err != nil {
		settings.TillerHost = ""
		return err
	}

	cmd.client = newClient()
	cmd.purge = purge
	cmd.name = releaseName

	global.Logger.Info(logFields, "cmd.run()")
	err = cmd.run()
	if err != nil {
		settings.TillerHost = ""
		return err
	}

	global.Logger.Info(logFields, "teardown()")
	settings.TillerHost = ""
	return nil

}

func (d *deleteCmd) run() error {
	opts := []helm.DeleteOption{
		helm.DeleteDryRun(d.dryRun),
		helm.DeleteDisableHooks(d.disableHooks),
		helm.DeletePurge(d.purge),
		helm.DeleteTimeout(d.timeout),
		helm.DeleteDescription(d.description),
	}
	res, err := d.client.DeleteRelease(d.name, opts...)
	if res != nil && res.Info != "" {
		fmt.Fprintln(d.out, res.Info)
	}

	return prettyError(err)
}
