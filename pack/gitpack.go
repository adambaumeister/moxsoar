package pack

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"os"
	"path"
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
	fmt.Printf("%v %v", gp.ContentDir, gp.RepoPath)

	_, err := git.PlainClone(gp.ContentDir, false, &git.CloneOptions{
		URL:      gp.RepoPath,
		Progress: os.Stdout,
	})

	if err != nil {
		return fmt.Errorf("Failed to clone %v (%v)", gp.RepoPath)
	}

	return nil
}

func (gp *GitPack) Update(pn string) (*string, error) {
	repo, err := git.PlainOpen(path.Join(gp.ContentDir, pn))
	if err != nil {
		return nil, fmt.Errorf("Failed to open repository for udpate.")
	}

	w, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("Couldn't checkout a worktree.")
	}

	err = w.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil {
		return nil, err
	}

	ref, _ := repo.Head()
	hs := ref.Hash().String()
	return &hs, nil
}

func (gp *GitPack) Status(pn string) (git.Status, error) {
	repo, err := git.PlainOpen(path.Join(gp.ContentDir, pn))
	if err != nil {
		return nil, fmt.Errorf("Failed to open repository for udpate.")
	}

	w, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("Couldn't checkout a worktree.")
	}

	s, err := w.Status()
	return s, err
}
