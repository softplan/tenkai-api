package helmapi

import (
	"fmt"
	"io"
	"k8s.io/helm/pkg/helm"
)

type rollbackCmd struct {
	name          string
	revision      int32
	dryRun        bool
	recreate      bool
	force         bool
	disableHooks  bool
	out           io.Writer
	client        helm.Interface
	timeout       int64
	wait          bool
	description   string
	cleanupOnFail bool
}

func (r *rollbackCmd) run() error {
	_, err := r.client.RollbackRelease(
		r.name,
		helm.RollbackDryRun(r.dryRun),
		helm.RollbackRecreate(r.recreate),
		helm.RollbackForce(r.force),
		helm.RollbackDisableHooks(r.disableHooks),
		helm.RollbackVersion(r.revision),
		helm.RollbackTimeout(r.timeout),
		helm.RollbackWait(r.wait),
		helm.RollbackDescription(r.description),
		helm.RollbackCleanupOnFail(r.cleanupOnFail))
	if err != nil {
		return prettyError(err)
	}

	fmt.Fprintf(r.out, "Rollback was a success.\n")

	return nil
}
