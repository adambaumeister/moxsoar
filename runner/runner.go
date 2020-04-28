package runner

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
	Address string
	PortMin int
	PortMax int
}
