package helmapi

import (
	"fmt"
	"io"
	"os"

	"k8s.io/helm/pkg/helm/helmpath"
	"k8s.io/helm/pkg/repo"
)

type repoRemoveCmd struct {
	out  io.Writer
	name string
	home helmpath.Home
}

//RemoveRepository - Remove a repository
func RemoveRepository(name string) error {
	remove := &repoRemoveCmd{out: os.Stdout}
	remove.home = settings.Home
	remove.name = name
	if err := remove.run(); err != nil {
		return err
	}
	return nil
}

func (r *repoRemoveCmd) run() error {
	return removeRepoLine(r.out, r.name, r.home)
}

func removeRepoLine(out io.Writer, name string, home helmpath.Home) error {
	repoFile := home.RepositoryFile()
	r, err := repo.LoadRepositoriesFile(repoFile)
	if err != nil {
		return err
	}

	if !r.Remove(name) {
		return fmt.Errorf("no repo named %q found", name)
	}
	if err := r.WriteFile(repoFile, 0644); err != nil {
		return err
	}

	if err := removeRepoCache(name, home); err != nil {
		return err
	}

	fmt.Fprintf(out, "%q has been removed from your repositories\n", name)

	return nil
}

func removeRepoCache(name string, home helmpath.Home) error {
	if _, err := os.Stat(home.CacheIndex(name)); err == nil {
		err = os.Remove(home.CacheIndex(name))
		if err != nil {
			return err
		}
	}
	return nil
}
