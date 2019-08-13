package helmapi

import (
	"fmt"
	"io"

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
