package helmapi

import (
	"bytes"
	"fmt"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/portforwarder"
	"k8s.io/helm/pkg/kube"
	"sync"
)

//HelmServiceInterface - Interface
type HelmServiceInterface interface {
	InitializeHelm()
	GetServices(kubeconfig string, namespace string) ([]model.Service, error)
	DeletePod(kubeconfig string, podName string, namespace string) error
	GetPods(kubeconfig string, namespace string) ([]model.Pod, error)
	AddRepository(repo model.Repository) error
	GetRepositories() ([]model.Repository, error)
	RemoveRepository(name string) error
	SearchCharts(searchTerms []string, allVersions bool) *[]model.SearchResult
	DeleteHelmRelease(kubeconfig string, releaseName string, purge bool) error
	Get(kubeconfig string, releaseName string, revision int) (string, error)
	IsThereAnyPodWithThisVersion(kubeconfig string, namespace string, releaseName string, tag string) (bool, error)
	GetReleaseHistory(kubeconfig string, releaseName string) (bool, error)
	GetHelmReleaseHistory(kubeconfig string, releaseName string) (ReleaseHistory, error)
	GetTemplate(mutex *sync.Mutex, chartName string, version string, kind string) ([]byte, error)
	GetDeployment(chartName string, version string) ([]byte, error)
	GetValues(chartName string, version string) ([]byte, error)
	ListHelmDeployments(kubeconfig string, namespace string) (*HelmListResult, error)
	RepoUpdate() error
	RollbackRelease(kubeconfig string, releaseName string, revision int) error
	Upgrade(upgradeRequest UpgradeRequest, out *bytes.Buffer) error
	HelmExecutorFunc(kubeconfig string, cmd HelmCommand) error
	GetHelmConnection() HelmConnection
	HelmCommandExecutor(fn HelmExecutorFunc) HelmExecutorFunc
}

//HelmServiceImpl - Concrete type
type HelmServiceImpl struct {
	HelmConnection HelmConnection
}

//HelmServiceBuilder HelmServiceBuilder
func HelmServiceBuilder() *HelmServiceImpl {
	r := HelmServiceImpl{}
	return &r
}

//HelmCommand HelmCommand
type HelmCommand interface {
	run() error
	SetNewClient(helmConnection HelmConnection, tillerHost string)
}

//HelmExecutorFunc HelmExecutorFunc
type HelmExecutorFunc func(kubeconfig string, cmd HelmCommand) error

//HelmCommandExecutor HelmCommandExecutor
func (svc HelmServiceImpl) HelmCommandExecutor(fn HelmExecutorFunc) HelmExecutorFunc {
	return func(kubeconfig string, cmd HelmCommand) error {
		tillerHost, tunnel, err := svc.GetHelmConnection().SetupConnection(kubeconfig)
		defer svc.GetHelmConnection().Teardown(tunnel)
		if err != nil {
			return err
		}
		cmd.SetNewClient(svc.GetHelmConnection(), tillerHost)
		err = fn(kubeconfig, cmd)
		return nil
	}
}

//HelmExecutorFunc HelmExecutorFunc
func (svc HelmServiceImpl) HelmExecutorFunc(kubeconfig string, cmd HelmCommand) error {
	return cmd.run()
}

//GetHelmConnection GetHelmConnection
func (svc HelmServiceImpl) GetHelmConnection() HelmConnection {
	if svc.HelmConnection == nil {
		svc.HelmConnection = HelmConnectionImpl{}
	}
	return svc.HelmConnection
}

const (
	tillerNamespace string = "kube-system"
)

//HelmConnection HelmConnection
type HelmConnection interface {
	SetupConnection(kubeConfig string) (string, *kube.Tunnel, error)
	Teardown(tillerTunnel *kube.Tunnel)
	ConfigForContext(context string, kubeconfig string) (*rest.Config, error)
	NewClient(tillerHost string) helm.Interface
	GetKubeClient(context string, kubeconfig string) (*rest.Config, kubernetes.Interface, error)
}

//HelmConnectionImpl HelmConnectionImpl
type HelmConnectionImpl struct {
}

//SetupConnection SetupConnection
func (h HelmConnectionImpl) SetupConnection(kubeConfig string) (string, *kube.Tunnel, error) {

	config, client, err := h.GetKubeClient("", kubeConfig)
	if err != nil {
		return "", nil, err
	}

	tillerTunnel, err := portforwarder.New(tillerNamespace, client, config)
	if err != nil {
		return "", nil, err
	}

	theTillerHost := fmt.Sprintf("127.0.0.1:%d", tillerTunnel.Local)

	// Plugin support.
	return theTillerHost, tillerTunnel, nil
}

//Teardown Teardown
func (h HelmConnectionImpl) Teardown(tillerTunnel *kube.Tunnel) {
	if tillerTunnel != nil {
		tillerTunnel.Close()
	}
}

// prettyError unwraps or rewrites certain errors to make them more user-friendly.
func prettyError(err error) error {
	// Add this check can prevent the object creation if err is nil.
	if err == nil {
		return nil
	}
	// If it's grpc's error, make it more user-friendly.
	if s, ok := status.FromError(err); ok {
		return fmt.Errorf(s.Message())
	}
	// Else return the original error.
	return err
}

// ConfigForContext creates a Kubernetes REST client configuration for a given kubeconfig context.
func (h HelmConnectionImpl) ConfigForContext(context string, kubeconfig string) (*rest.Config, error) {
	config, err := kube.GetConfig(context, kubeconfig).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("could not get Kubernetes config for context %q: %s", context, err)
	}
	return config, nil
}

// GetKubeClient creates a Kubernetes config and client for a given kubeconfig context.
func (h HelmConnectionImpl) GetKubeClient(context string, kubeconfig string) (*rest.Config, kubernetes.Interface, error) {
	config, err := h.ConfigForContext(context, kubeconfig)
	if err != nil {
		return nil, nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("could not get Kubernetes client: %s", err)
	}
	return config, client, nil
}

//NewClient NewClient
func (h HelmConnectionImpl) NewClient(tillerHost string) helm.Interface {
	options := []helm.Option{helm.Host(tillerHost), helm.ConnectTimeout(1200)}
	return helm.NewClient(options...)
}
