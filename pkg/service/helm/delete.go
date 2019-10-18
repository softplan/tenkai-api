package helmapi

import (
	"fmt"
	"io"
	"os"

	"k8s.io/helm/pkg/helm"
)

type deleteCmdInterface interface {
	run() error
}

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

func deleteCmdBuilder(releaseName string, purge bool) *deleteCmd {
	cmd := &deleteCmd{out: os.Stdout}
	cmd.client = newClient()
	cmd.purge = purge
	cmd.name = releaseName
	return cmd
}

//DeleteHelmRelease - Delete a Release
func (svc HelmServiceImpl) DeleteHelmRelease(kubeconfig string, releaseName string, purge bool) error {
	svc.EnsureSettings(kubeconfig)
	cmd := deleteCmdBuilder(releaseName, purge)
	return doDeleteHelmRelease(*cmd)
}

func doDeleteHelmRelease(cmd deleteCmd) error {
	err := setupConnection()
	defer teardown()
	if err != nil {
		settings.TillerHost = ""
		return err
	}
	err = cmd.run()
	if err != nil {
		settings.TillerHost = ""
		return err
	}
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
