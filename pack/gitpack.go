package pack

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"os"
)

type GitPack struct {
	RepoPath   string
	ContentDir string

	Packs []*Pack
}

func GetPackFromGit(contentdir string, repopath string) (*GitPack, error) {
	// Attempt to pull the content pack from GIT
	gp := GitPack{
		RepoPath:   repopath,
		ContentDir: contentdir,
	}
	err := gp.Clone()
	if err != nil {
		return nil, err
	}

	return &gp, nil
}

func (gp *GitPack) Clone() error {
	_, err := git.PlainClone(gp.ContentDir, false, &git.CloneOptions{
		URL:      gp.RepoPath,
		Progress: os.Stdout,
	})

	if err != nil {
		return fmt.Errorf("Failed to clone %v", gp.RepoPath)
	}

	return nil
}
