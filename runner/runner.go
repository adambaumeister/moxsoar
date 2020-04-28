package runner

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"path"
)

const DEFAULT_RUNNER_CONFIG = "runner.yml"

/*
RunConfig is the configuration passed to the runner object
*/
type RunConfig struct {
	Runner Runner
}

/*
Runner handles executing the Mock handlers and distributing them across the available system ports
*/
type Runner struct {
	Address    string
	PortMin    int
	PortMax    int
	ContentDir string

	currentPort int
}

func (r *Runner) GetAddress() string {
	// Get the next port/address combo
	a := fmt.Sprintf("%v:%v", r.Address, r.currentPort)
	r.currentPort = r.currentPort + 1

	return a
}

func GetRunConfig(contentDir string) RunConfig {
	// Get the runner configuration
	b, err := ioutil.ReadFile(path.Join(contentDir, DEFAULT_RUNNER_CONFIG))
	if err != nil {
		log.Fatal(err)
	}

	rc := RunConfig{}
	err = yaml.Unmarshal(b, &rc)

	if err != nil {
		log.Fatal(err)
	}
	rc.Runner.ContentDir = contentDir
	rc.Runner.currentPort = rc.Runner.PortMin

	return rc
}
