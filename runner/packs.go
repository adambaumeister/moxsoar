package runner

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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

func GetPackIndex(contentDir string) PackIndex {
	// Lookup the pack index
	b, err := ioutil.ReadFile(path.Join(contentDir, PACK_INDEX_FILE))
	if err != nil {
		panic(err)
	}

	pi := PackIndex{}
	err = json.Unmarshal(b, &pi)
	if err != nil {
		panic(err)
	}

	for _, pack := range pi.Packs {
		pack.FullPath = path.Join(contentDir, pack.Path)

	}

	return pi

}
