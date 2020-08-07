package pack

import (
	"encoding/json"
	"fmt"
	"github.com/adambaumeister/moxsoar/integrations"
	"github.com/adambaumeister/moxsoar/settings"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
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

	Running []*integrations.BaseIntegration `yaml:"running,omitempty"`

	settings *settings.Settings
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

func (rc *RunConfig) Reread() {
	b, err := ioutil.ReadFile(path.Join(rc.Runner.PackDir, DEFAULT_RUNNER_CONFIG))
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(b, rc)
	if err != nil {
		log.Fatal(err)
	}
}

func GetRunConfig(packDir string, settings *settings.Settings) *RunConfig {
	// Get the runner configuration

	b, err := ioutil.ReadFile(path.Join(packDir, DEFAULT_RUNNER_CONFIG))
	if err != nil {
		log.Fatal(err)
	}

	rc := RunConfig{
		settings: settings,
	}

	rc.Runner.PackDir = packDir
	err = yaml.Unmarshal(b, &rc)

	if err != nil {
		log.Fatal(err)
	}
	rc.Runner.currentPort = rc.Runner.PortMin

	return &rc
}

func (rc *RunConfig) Save() error {
	b, err := yaml.Marshal(rc)
	if err != nil {
		return fmt.Errorf("Could not save the runenr config!")
	}
	err = ioutil.WriteFile(path.Join(rc.Runner.PackDir, DEFAULT_RUNNER_CONFIG), b, 755)
	if err != nil {
		return fmt.Errorf("Could not save the runenr config!")
	}
	return nil
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
			go i.Start(run.Integration, rc.settings)
			rc.Running = append(rc.Running, &i)
		}
	}
}

func (rc *RunConfig) Prepare() {
	// Prepares all the integrations to be run
	// This is a special function that is called as part of the test suite to create mock runners
	for _, run := range rc.Run {
		exitChan := make(chan bool)
		switch run.Integration {
		default:
			addr := rc.Runner.GetAddress()
			i := integrations.BaseIntegration{
				Name:     run.Integration,
				ExitChan: exitChan,
				Addr:     addr,
				PackDir:  rc.Runner.PackDir,
			}
			i.ReadRoutes(path.Join(i.PackDir, i.Name, integrations.ROUTE_FILE))
			rc.Running = append(rc.Running, &i)
		}
	}
}

func (rc *RunConfig) Shutdown() {
	// Shut down all the running integrations
	for _, running := range rc.Running {
		fmt.Printf("Shutting down %v\n", running.Name)
		running.ExitChan <- true
		// Wait for it to exit
		<-running.ExitChan
	}
	rc.Running = []*integrations.BaseIntegration{}
	rc.Runner.currentPort = rc.Runner.PortMin
}

func (rc *RunConfig) Restart() {
	// Restart all running integrations
	rc.Shutdown()
	rc.RunAll()
}

func (rc *RunConfig) AddIntegration(name string) error {
	// Copy the RunConfig, this ensures only marshalable stuff is in there
	nrc := GetRunConfig(path.Join(rc.Runner.PackDir), &settings.Settings{})

	nrc.Run = append(nrc.Run, Run{
		Integration: name,
	})
	err := nrc.Save()
	if err != nil {
		return err
	}

	// Create the integration directory
	err = os.Mkdir(path.Join(rc.Runner.PackDir, name), 755)
	if err != nil {
		return err
	}

	// create the default routes file
	dr := []*integrations.Route{}
	i := integrations.BaseIntegration{
		Routes: dr,
		Name:   name,
	}
	b, err := json.Marshal(i)
	err = ioutil.WriteFile(path.Join(rc.Runner.PackDir, name, integrations.ROUTE_FILE), b, 755)

	rc.Run = append(rc.Run, Run{
		Integration: name,
	})
	rc.Restart()
	return err
}

func (rc *RunConfig) DeleteIntegration(name string) error {
	intPath := path.Join(rc.Runner.PackDir, name)
	if _, err := os.Stat(intPath); os.IsNotExist(err) {
		return fmt.Errorf("Integration directory does not exist: %v", name)
	}

	directoryList, err := ioutil.ReadDir(intPath)
	if err != nil {
		return err
	}

	for _, f := range directoryList {
		p := path.Join(intPath, f.Name())
		err := os.Remove(p)
		if err != nil {
			return fmt.Errorf("Failed to delete %v: %v", p, err)
		}
	}
	// Clear the directory
	err = os.Remove(intPath)
	if err != nil {
		return err
	}

	// Clobber the integration out of the run config
	nrc := GetRunConfig(path.Join(rc.Runner.PackDir), &settings.Settings{})
	newRun := []Run{}
	for _, r := range nrc.Run {
		if r.Integration != name {
			newRun = append(newRun, r)
		}
	}
	nrc.Run = newRun
	err = nrc.Save()
	if err != nil {
		return err
	}
	// Also update the run register for the actual runconfig
	rc.Run = newRun

	rc.Restart()
	return nil
}
