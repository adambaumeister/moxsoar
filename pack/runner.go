package pack

import (
	"context"
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
	Info   Info
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

type Info struct {
	Description string
	Author      string
	Version     string
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

	ctx := context.Background()
	exitChan := make(chan bool)
	for _, run := range rc.Run {
		switch run.Integration {
		case "minemeld":
			fmt.Printf("Starting minemeld integration.\n")
			i := integrations.BaseIntegration{
				Ctx:      ctx,
				ExitChan: exitChan,
			}
			go i.Start("minemeld", rc.Runner.PackDir, rc.Runner.GetAddress())
		case "servicenow":
			fmt.Printf("Starting SNOW integration.\n")
			i := integrations.BaseIntegration{
				Ctx:      ctx,
				ExitChan: exitChan,
			}
			go i.Start("servicenow", rc.Runner.PackDir, rc.Runner.GetAddress())
		}
	}

	// Here is an example of how this can work, we can tell the servers to exit with a channel
	// Need to change this to be a channel per server
	// Also need a status check
	//time.Sleep(5*time.Second)
	//exitChan <- true

}
