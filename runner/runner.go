package runner

import (
	"fmt"
	"github.com/adambaumeister/moxsoar/integrations"
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
	Run    []Run
}

/*
Runner handles executing the Mock handlers and distributing them across the available system ports
*/
type Runner struct {
	Address    string
	PortMin    int
	PortMax    int
	ContentDir string
	PackDir    string

	currentPort int
}

/*
Run definitions
*/
type Run struct {
	Integration string
}

func (r *Runner) GetAddress() string {
	// Get the next port/address combo
	a := fmt.Sprintf("%v:%v", r.Address, r.currentPort)
	r.currentPort = r.currentPort + 1

	return a
}

func GetRunConfig(packDir string) RunConfig {
	// Get the runner configuration

	b, err := ioutil.ReadFile(path.Join(packDir, DEFAULT_RUNNER_CONFIG))
	if err != nil {
		log.Fatal(err)
	}

	rc := RunConfig{}
	rc.Runner.PackDir = packDir
	err = yaml.Unmarshal(b, &rc)

	if err != nil {
		log.Fatal(err)
	}
	rc.Runner.currentPort = rc.Runner.PortMin

	return rc
}

func (rc *RunConfig) RunAll() {
	/*
		Start all the configured mock integrations
	*/
	for _, run := range rc.Run {
		switch run.Integration {
		case "minemeld":
			i := integrations.BaseIntegration{}
			i.Start("minemeld", rc.Runner.PackDir, rc.Runner.GetAddress())
		}
	}
}
