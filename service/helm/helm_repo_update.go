package helmapi

import (
	"errors"
	"fmt"
	"github.com/softplan/tenkai-api/global"
	"io"
	"os"
	"sync"

	"k8s.io/helm/pkg/getter"
	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/repo"
)

var errNoRepositories = errors.New("no repositories found. You must add one before updating")

type repoUpdateCmd struct {
	update func([]*repo.ChartRepository, io.Writer, helmpath.Home, bool) error
	home   helmpath.Home
	out    io.Writer
	strict bool
}

//RepoUpdate - Update a repository
func RepoUpdate() error {

	settings.Home = global.HELM_DIR

	u := &repoUpdateCmd{
		out:    os.Stdout,
		update: updateCharts,
	}

	u.home = settings.Home

	return u.run()

}

func (u *repoUpdateCmd) run() error {
	f, err := repo.LoadRepositoriesFile(u.home.RepositoryFile())
	if err != nil {
		return err
	}

	if len(f.Repositories) == 0 {
		return errNoRepositories
	}
	var repos []*repo.ChartRepository
	for _, cfg := range f.Repositories {
		r, err := repo.NewChartRepository(cfg, getter.All(settings))
		if err != nil {
			return err
		}
		repos = append(repos, r)
	}
	return u.update(repos, u.out, u.home, u.strict)
}

func updateCharts(repos []*repo.ChartRepository, out io.Writer, home helmpath.Home, strict bool) error {
	fmt.Fprintln(out, "Hang tight while we grab the latest from your chart repositories...")
	var (
		errorCounter int
		wg           sync.WaitGroup
		mu           sync.Mutex
	)
	for _, re := range repos {
		wg.Add(1)
		go func(re *repo.ChartRepository) {
			defer wg.Done()
			if re.Config.Name == "local" {
				mu.Lock()
				fmt.Fprintf(out, "...Skip %s chart repository\n", re.Config.Name)
				mu.Unlock()
				return
			}
			err := re.DownloadIndexFile(home.Cache())
			if err != nil {
				mu.Lock()
				errorCounter++
				fmt.Fprintf(out, "...Unable to get an update from the %q chart repository (%s):\n\t%s\n", re.Config.Name, re.Config.URL, err)
				mu.Unlock()
			} else {
				mu.Lock()
				fmt.Fprintf(out, "...Successfully got an update from the %q chart repository\n", re.Config.Name)
				mu.Unlock()
			}
		}(re)
	}
	wg.Wait()

	if errorCounter != 0 && strict {
		return errors.New("Update Failed. Check log for details")
	}

	fmt.Fprintln(out, "Update Complete.")
	return nil
}
