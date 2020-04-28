package integrations

type BaseIntegration struct {
	Routes map[string]Route
}

type Route struct {
	Path         string
	ResponseFile string
	ResponseCode int
}
