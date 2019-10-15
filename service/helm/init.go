package helmapi

import (
	"fmt"
	"github.com/softplan/tenkai-api/global"
	"io"
	"k8s.io/helm/pkg/getter"
	helm_env "k8s.io/helm/pkg/helm/environment"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/repo"
	"os"
)

const (
	stableRepositoryURL = "https://kubernetes-charts.storage.googleapis.com"

	localRepositoryURL = "http://127.0.0.1:8879/charts"

	stableRepository = "stable"

	// LocalRepository is the standard name of the local repository
	LocalRepository = "local"

	// LocalRepositoryIndexFile is the standard name of the local repository index file
	LocalRepositoryIndexFile = "index.yaml"
)

//InitializeHelm - Initialize a Helm repository
func (svc HelmServiceImpl) InitializeHelm() {
	settings.TillerNamespace = "kube-system"
	settings.TLSEnable = false
	settings.TLSVerify = false
	settings.TillerConnectionTimeout = 1200
	initialize(global.HelmDir, os.Stdout, true, settings)
}

// Initialize initializes local config
// Returns an error if the command failed.
func initialize(home helmpath.Home, out io.Writer, skipRefresh bool, settings helm_env.EnvSettings) error {
	if err := ensureDirectories(home, out); err != nil {
		return err
	}
	if err := ensureDefaultRepos(home, out, skipRefresh, settings, stableRepositoryURL, localRepositoryURL); err != nil {
		return err
	}

	return ensureRepoFileFormat(home.RepositoryFile(), out)
}

// ensureDirectories checks to see if $HELM_HOME exists.
//
// If $HELM_HOME does not exist, this function will create it.
func ensureDirectories(home helmpath.Home, out io.Writer) error {
	configDirectories := []string{
		home.String(),
		home.Repository(),
		home.Cache(),
		home.LocalRepository(),
		home.Plugins(),
		home.Starters(),
		home.Archive(),
	}
	for _, p := range configDirectories {
		if fi, err := os.Stat(p); err != nil {
			fmt.Fprintf(out, "Creating %s \n", p)
			if err := os.MkdirAll(p, 0755); err != nil {
				return fmt.Errorf("Could not create %s: %s", p, err)
			}
		} else if !fi.IsDir() {
			return fmt.Errorf("%s must be a directory", p)
		}
	}

	return nil
}

func ensureDefaultRepos(home helmpath.Home, out io.Writer, skipRefresh bool, settings helm_env.EnvSettings, stableRepositoryURL, localRepositoryURL string) error {
	repoFile := home.RepositoryFile()
	if fi, err := os.Stat(repoFile); err != nil {
		fmt.Fprintf(out, "Creating %s \n", repoFile)
		f := repo.NewRepoFile()
		sr, err := initStableRepo(home.CacheIndex(stableRepository), home, out, skipRefresh, settings, stableRepositoryURL)
		if err != nil {
			return err
		}
		lr, err := initLocalRepo(home.LocalRepository(LocalRepositoryIndexFile), home.CacheIndex("local"), home, out, settings, localRepositoryURL)
		if err != nil {
			return err
		}
		f.Add(sr)
		f.Add(lr)
		if err := f.WriteFile(repoFile, 0644); err != nil {
			return err
		}
	} else if fi.IsDir() {
		return fmt.Errorf("%s must be a file, not a directory", repoFile)
	}
	return nil
}

func initStableRepo(cacheFile string, home helmpath.Home, out io.Writer, skipRefresh bool, settings helm_env.EnvSettings, stableRepositoryURL string) (*repo.Entry, error) {
	fmt.Fprintf(out, "Adding %s repo with URL: %s \n", stableRepository, stableRepositoryURL)
	c := repo.Entry{
		Name:  stableRepository,
		URL:   stableRepositoryURL,
		Cache: cacheFile,
	}
	r, err := repo.NewChartRepository(&c, getter.All(settings))
	if err != nil {
		return nil, err
	}

	if skipRefresh {
		return &c, nil
	}

	// In this case, the cacheFile is always absolute. So passing empty string
	// is safe.
	if err := r.DownloadIndexFile(""); err != nil {
		return nil, fmt.Errorf("Looks like %q is not a valid chart repository or cannot be reached: %s", stableRepositoryURL, err.Error())
	}

	return &c, nil
}

func initLocalRepo(indexFile, cacheFile string, home helmpath.Home, out io.Writer, settings helm_env.EnvSettings, localRepositoryURL string) (*repo.Entry, error) {
	if fi, err := os.Stat(indexFile); err != nil {
		fmt.Fprintf(out, "Adding %s repo with URL: %s \n", LocalRepository, localRepositoryURL)
		i := repo.NewIndexFile()
		if err := i.WriteFile(indexFile, 0644); err != nil {
			return nil, err
		}

		//TODO: take this out and replace with helm update functionality
		if err := createLink(indexFile, cacheFile, home); err != nil {
			return nil, err
		}
	} else if fi.IsDir() {
		return nil, fmt.Errorf("%s must be a file, not a directory", indexFile)
	}

	return &repo.Entry{
		Name:  LocalRepository,
		URL:   localRepositoryURL,
		Cache: cacheFile,
	}, nil
}

func createLink(indexFile, cacheFile string, home helmpath.Home) error {
	return os.Symlink(indexFile, cacheFile)
}

func ensureRepoFileFormat(file string, out io.Writer) error {
	r, err := repo.LoadRepositoriesFile(file)
	if err == repo.ErrRepoOutOfDate {
		fmt.Fprintln(out, "Updating repository file format...")
		if err := r.WriteFile(file, 0644); err != nil {
			return err
		}
	}

	return nil
}
