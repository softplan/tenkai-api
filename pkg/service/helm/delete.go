package helmapi

import (
	"fmt"
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

	cmd := &deleteCmd{out: os.Stdout}
	cmd.purge = purge
	cmd.name = releaseName

	return svc.HelmCommandExecutor(svc.HelmExecutorFunc)(kubeconfig, cmd)

}

func (d *deleteCmd) SetNewClient(helmConnection HelmConnection, tillerHost string) {
	d.client = helmConnection.NewClient(tillerHost)
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
