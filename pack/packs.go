package pack

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
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
	Packs []*Pack

	ContentDir string
}

type Pack struct {
	Name    string
	Comment string
	Version string
	Path    string

	FullPath string
}

func (p *PackIndex) GetPackName(pn string) (*Pack, error) {
	for _, pack := range p.Packs {
		if pack.Name == pn {
			return pack, nil
		}
	}
	return nil, fmt.Errorf("Invalid pack name %v supplied", pn)
}

func (p *PackIndex) Get(packName string, repopath string) {
	//gp, err := GetPackFromGit(p.ContentDir, repopath)
	_, err := GetPackFromGit(path.Join(p.ContentDir, packName), repopath)
	if err != nil {
		log.Fatal(err)
	}
}

func GetPackIndex(contentDir string) PackIndex {
	// Lookup the pack index
	packs := []*Pack{}
	pi := PackIndex{
		ContentDir: contentDir,
		Packs:      packs,
	}

	b, err := ioutil.ReadFile(path.Join(contentDir, PACK_INDEX_FILE))
	// If it doesn't exist, create an empty index and return immediately.
	if err != nil {
		pi.WritePackIndex(path.Join(contentDir, PACK_INDEX_FILE))
		return pi
	}

	err = json.Unmarshal(b, &pi)
	if err != nil {
		panic(err)
	}

	for _, pack := range pi.Packs {
		pack.FullPath = path.Join(contentDir, pack.Path)

	}

	return pi

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
