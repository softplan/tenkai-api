package helmapi

import (
	"fmt"
	"github.com/softplan/tenkai-api/pkg/global"
	"io"
	"k8s.io/helm/pkg/helm"
	"os"
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

//RollbackRelease - Rollback a release
func (svc HelmServiceImpl) RollbackRelease(kubeconfig string, releaseName string, revision int) error {

	settings.KubeConfig = kubeconfig
	settings.Home = global.HelmDir
	settings.TillerNamespace = "kube-system"
	settings.TLSEnable = false
	settings.TLSVerify = false
	settings.TillerConnectionTimeout = 1200
	err := setupConnection()
	defer teardown()

	if err != nil {
		settings.TillerHost = ""
		return err
	}

	cmd := &rollbackCmd{out: os.Stdout}
	cmd.client = newClient()
	cmd.name = releaseName
	cmd.revision = int32(revision)

	err = cmd.run()

	settings.TillerHost = ""
	return err

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
