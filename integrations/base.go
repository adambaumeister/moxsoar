package integrations

import "regexp"

type BaseIntegration struct {
	Routes map[string]Route
}

func (bi *BaseIntegration) GetRoute(url string) Route {
	if route, ok := bi.Routes[url]; ok {
		return route
	}

	for _, route := range bi.Routes {
		m, _ := regexp.MatchString(route.Path, url)
		if m {
			return route
		}
	}

	// This is pretty hacky for now, need to improve this
	return Route{
		Path:         "/",
		ResponseFile: "default.json",
		ResponseCode: 200,
	}
}

type Route struct {
	Path         string
	ResponseFile string
	ResponseCode int
}
