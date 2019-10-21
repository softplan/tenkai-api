package helmapi

import (
	"bytes"
	"fmt"
	"io"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"strings"

	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	rls "k8s.io/helm/pkg/proto/hapi/services"
	"k8s.io/helm/pkg/renderutil"
	storageerrors "k8s.io/helm/pkg/storage/errors"
)

const upgradeDesc = ``

type valueFiles []string

type upgradeCmd struct {
	release       string
	chart         string
	out           io.Writer
	client        helm.Interface
	dryRun        bool
	recreate      bool
	force         bool
	disableHooks  bool
	valueFiles    valueFiles
	values        []string
	stringValues  []string
	fileValues    []string
	verify        bool
	keyring       string
	install       bool
	namespace     string
	version       string
	timeout       int64
	resetValues   bool
	reuseValues   bool
	wait          bool
	atomic        bool
	repoURL       string
	username      string
	password      string
	devel         bool
	subNotes      bool
	description   string
	cleanupOnFail bool

	certFile string
	keyFile  string
	caFile   string
}

//UpgradeRequest UpgradeRequest
type UpgradeRequest struct {
	Kubeconfig   string
	Release      string
	Chart        string
	ChartVersion string
	Namespace    string
	Variables    []string
	Dryrun       bool
}

//Upgrade Method
func (svc HelmServiceImpl) Upgrade(upgradeRequest UpgradeRequest, out *bytes.Buffer) error {

	svc.EnsureSettings(upgradeRequest.Kubeconfig)

	if upgradeRequest.Dryrun {
		settings.Debug = true
	}

	upgrade := &upgradeCmd{out: out}

	err := setupConnection()
	defer teardown()

	if err == nil {
		upgrade.client = newClient()

		if upgradeRequest.Dryrun {
			upgrade.dryRun = true
		}

		if upgradeRequest.ChartVersion != "" {
			upgrade.version = upgradeRequest.ChartVersion
		} else {
			upgrade.version = ">0.0.0-0"
		}

		upgrade.install = true
		upgrade.recreate = false
		upgrade.force = true
		upgrade.release = upgradeRequest.Release
		upgrade.chart = upgradeRequest.Chart
		upgrade.values = upgradeRequest.Variables
		upgrade.client = ensureHelmClient(upgrade.client)
		upgrade.wait = upgrade.wait || upgrade.atomic
		upgrade.namespace = upgradeRequest.Namespace
		err = upgrade.run()
		settings.KubeConfig = ""
	}
	settings.TillerHost = ""
	settings.Debug = false
	return err
}

func (u *upgradeCmd) doInstall(err error, releaseHistory *rls.GetHistoryResponse, chartPath string) (bool, error) {

	if err == nil {
		if u.namespace == "" {
			u.namespace = defaultNamespace()
		}
		previousReleaseNamespace := releaseHistory.Releases[0].Namespace
		if previousReleaseNamespace != u.namespace {
			fmt.Fprintf(u.out,
				"WARNING: Namespace %q doesn't match with previous. Release will be deployed to %s\n",
				u.namespace, previousReleaseNamespace,
			)
		}
	}

	if err != nil && strings.Contains(err.Error(), storageerrors.ErrReleaseNotFound(u.release).Error()) {
		fmt.Fprintf(u.out, "Release %q does not exist. Installing it now.\n", u.release)
		ic := &installCmd{
			chartPath:    chartPath,
			client:       u.client,
			out:          u.out,
			name:         u.release,
			valueFiles:   u.valueFiles,
			dryRun:       u.dryRun,
			verify:       u.verify,
			disableHooks: u.disableHooks,
			keyring:      u.keyring,
			values:       u.values,
			stringValues: u.stringValues,
			fileValues:   u.fileValues,
			namespace:    u.namespace,
			timeout:      u.timeout,
			wait:         u.wait,
			description:  u.description,
			atomic:       u.atomic,
		}
		return true, ic.run()
	}

	return false, nil

}

func (u *upgradeCmd) checkChart(chartPath string) (*chart.Chart, error) {

	ch, err := chartutil.Load(chartPath)
	if err == nil {

		if req, err := chartutil.LoadRequirements(ch); err == nil {
			if err := renderutil.CheckDependencies(ch, req); err != nil {
				return ch, err
			}
			return ch, nil
		}

		if err != chartutil.ErrRequirementsNotFound {
			return nil, fmt.Errorf("cannot load requirements: %v", err)
		}
		return ch, nil

	}
	return nil, prettyError(err)
}

func (u *upgradeCmd) run() error {

	chartPath, err := locateChartPath(u.repoURL, u.username, u.password, u.chart, u.version, u.verify, u.keyring, u.certFile, u.keyFile, u.caFile)
	if err != nil {
		return err
	}

	//TODO - VERIFY IF CONFIG FILE EXISTS !!! This is the cause of  u.client.ReleaseHistory fail sometimes.
	releaseHistory, err := u.client.ReleaseHistory(u.release, helm.WithMaxHistory(1))

	if u.install {
		exec, err := u.doInstall(err, releaseHistory, chartPath)
		if exec {
			return err
		}
	}

	rawVals, err := vals(u.valueFiles, u.values, u.stringValues, u.fileValues, u.certFile, u.keyFile, u.caFile)
	if err != nil {
		return err
	}

	// Check chart requirements to make sure all dependencies are present in /charts
	ch, err := u.checkChart(chartPath)
	if err != nil {
		return err
	}

	_, err = u.client.UpdateReleaseFromChart(
		u.release,
		ch,
		helm.UpdateValueOverrides(rawVals),
		helm.UpgradeDryRun(u.dryRun),
		helm.UpgradeRecreate(u.recreate),
		helm.UpgradeForce(u.force),
		helm.UpgradeDisableHooks(u.disableHooks),
		helm.UpgradeTimeout(u.timeout),
		helm.ResetValues(u.resetValues),
		helm.ReuseValues(u.reuseValues),
		helm.UpgradeSubNotes(u.subNotes),
		helm.UpgradeWait(u.wait),
		helm.UpgradeDescription(u.description),
		helm.UpgradeCleanupOnFail(u.cleanupOnFail))

	if err != nil {

		fmt.Fprintf(u.out, "UPGRADE FAILED\nError: %v\n", prettyError(err))

		if u.atomic {
			fmt.Fprint(u.out, "ROLLING BACK")
			rollback := &rollbackCmd{
				out:           u.out,
				client:        u.client,
				name:          u.release,
				dryRun:        u.dryRun,
				recreate:      u.recreate,
				force:         u.force,
				timeout:       u.timeout,
				wait:          u.wait,
				description:   "",
				revision:      releaseHistory.Releases[0].Version,
				disableHooks:  u.disableHooks,
				cleanupOnFail: u.cleanupOnFail,
			}
			if err := rollback.run(); err != nil {
				return err
			}
		}

		return fmt.Errorf("UPGRADE FAILED: %v", prettyError(err))

	}

	fmt.Fprintf(u.out, "Release %q has been upgraded.\n", u.release)

	return nil
}
