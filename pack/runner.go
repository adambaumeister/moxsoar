package pack

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
	Info   Info
	Runner Runner
	Run    []Run

	Running []*integrations.BaseIntegration
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
	// This we read from Json
	Integration string

	// This is the actual object we retrieve later.
	integration *integrations.BaseIntegration
}

func (rc *RunConfig) GetIntegrations() []*integrations.BaseIntegration {
	return rc.Running
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

func GetRunConfig(packDir string) *RunConfig {
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

	return &rc
}

func (rc *RunConfig) RunAll() {
	/*
		Start all the configured mock integrations
	*/

	for _, run := range rc.Run {
		exitChan := make(chan bool)
		switch run.Integration {
		default:
			addr := rc.Runner.GetAddress()
			fmt.Printf("Starting %v integration.\n", run.Integration)
			i := integrations.BaseIntegration{
				Name:     run.Integration,
				ExitChan: exitChan,
				Addr:     addr,
				PackDir:  rc.Runner.PackDir,
			}
			go i.Start(run.Integration)
			rc.Running = append(rc.Running, &i)
		}
	}

	// Here is an example of how this can work, we can tell the servers to exit with a channel
	// Need to change this to be a channel per server
	// Also need a status check
	//time.Sleep(5*time.Second)
	//exitChan <- true

}

func (rc *RunConfig) Shutdown() {
	// Shut down all the running integrations
	for _, running := range rc.Running {
		fmt.Printf("Shutting down %v\n", running.Name)
		running.ExitChan <- true
	}
}
