package pack

import (
	"encoding/json"
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"io/ioutil"
	"os"
	"path"
)

const PACK_INDEX_FILE = "index.json"

/*
Packs are groups of content
	* runner.yml: Indicates which integrations are to run and any config specific to them
	* any number of integration directories: The actual content per integration

The pack index (index.json) maintains the index.
*/

type PackIndex struct {
	Packs      []*Pack
	ContentDir string
	indexfile  string
}

type Pack struct {
	Name    string
	Comment string
	Version string
	Path    string
	Active  bool

	Repo string
}

func (p *PackIndex) ActivatePack(pn string) (*Pack, error) {
	pack, err := p.GetPackName(pn)
	if err != nil {
		return nil, err
	}

	for _, oldPack := range p.Packs {
		if oldPack.Active {
			oldPack.Active = false
		}
	}

	pack.Active = true
	return pack, nil
}
func (p *PackIndex) GetPackName(pn string) (*Pack, error) {
	p.Reindex()
	for _, pack := range p.Packs {
		if pack.Name == pn {
			return pack, nil
		}
	}
	return nil, fmt.Errorf("Invalid pack name %v supplied", pn)
}

func (pi *PackIndex) GetOrClone(packName string, repopath string) (*Pack, error) {
	p, _ := pi.GetPackName(packName)
	if p == nil {
		err := pi.Get(packName, repopath)
		if err != nil {
			return nil, err
		}
	}
	p, err := pi.GetPackName(packName)
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve pack.")
	}

	return p, nil
}

func (pi *PackIndex) Update(packName string) (*string, error) {
	p, _ := pi.GetPackName(packName)
	if p == nil {
		return nil, fmt.Errorf("Invalid pack name.")
	}

	gp := GitPack{
		ContentDir: pi.ContentDir,
	}

	s, err := gp.Update(packName)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (pi *PackIndex) Status(packName string) (git.Status, error) {
	p, _ := pi.GetPackName(packName)
	if p == nil {
		return nil, fmt.Errorf("Invalid pack name.")
	}

	gp := GitPack{
		ContentDir: pi.ContentDir,
	}

	s, err := gp.Status(packName)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (pi *PackIndex) Save(packName string, commitmsg string, author object.Signature) error {
	p, _ := pi.GetPackName(packName)
	if p == nil {
		return fmt.Errorf("Invalid pack name.")
	}

	gp := GitPack{
		ContentDir: pi.ContentDir,
	}

	err := gp.Save(packName, commitmsg, &author)
	if err != nil {
		return err
	}
	return nil
}

func (pi *PackIndex) Reindex() {
	rePacks := []*Pack{}
	for _, pack := range pi.Packs {
		// This re-indexes by removing packs that are no longer on the system from the index
		if _, err := os.Stat(pack.Path); !os.IsNotExist(err) {
			rePacks = append(rePacks, pack)
		}
	}

	pi.Packs = rePacks

}

func (p *PackIndex) Get(packName string, repopath string) error {
	// Retrieve a pack from the given git url
	// If this pack already exists but has fallen out of the index just update it
	if _, err := os.Stat(path.Join(p.ContentDir, packName)); !os.IsNotExist(err) {
		gp := GitPack{
			ContentDir: p.ContentDir,
		}
		_, err := gp.Update(packName)
		if err != nil {
			return err
		}
	} else {
		_, err := GetPackFromGit(path.Join(p.ContentDir, packName), repopath)
		if err != nil {
			return err
		}
	}

	newPack := Pack{
		Name: packName,
		Path: path.Join(p.ContentDir, packName),
		Repo: repopath,
	}

	p.Packs = append(p.Packs, &newPack)
	p.WritePackIndex(p.indexfile)
	return nil
}

func GetPackIndex(contentDir string) *PackIndex {
	// Lookup the pack index
	packs := []*Pack{}
	pi := PackIndex{
		ContentDir: contentDir,
		Packs:      packs,
		indexfile:  path.Join(contentDir, PACK_INDEX_FILE),
	}

	b, err := ioutil.ReadFile(pi.indexfile)
	// If it doesn't exist, create an empty index and return immediately.
	if err != nil {
		pi.WritePackIndex(pi.indexfile)
		return &pi
	}

	err = json.Unmarshal(b, &pi)
	if err != nil {
		panic(err)
	}

	return &pi

}

func (pi *PackIndex) WritePackIndex(p string) {

	b, err := json.Marshal(pi)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(p, b, os.FileMode(644))
	if err != nil {
		panic(err)
	}
}
