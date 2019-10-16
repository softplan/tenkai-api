package helmapi

import (
	"bytes"
	"github.com/softplan/tenkai-api/pkg/dbms/model"
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
	Upgrade(kubeconfig string, release string, chart string, chartVersion string, namespace string, variables []string, out *bytes.Buffer, dryrun bool) error
}

//HelmServiceImpl - Concrete type
type HelmServiceImpl struct {
}
