package pack

import (
	"encoding/json"
	"fmt"
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
	_, err := GetPackFromGit(path.Join(p.ContentDir, packName), repopath)
	if err != nil {
		return err
	}
	newPack := Pack{
		Name: packName,
		Path: path.Join(p.ContentDir, packName),
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
