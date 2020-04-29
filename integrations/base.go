package integrations

import (
	"encoding/json"
	"fmt"
	"github.com/adambaumeister/moxsoar/api"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"regexp"
)

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

func (bi *BaseIntegration) Start(integrationName string, packDir string, addr string) {
	b, err := ioutil.ReadFile(path.Join(packDir, integrationName, "routes.json"))
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(b, &bi.Routes)

	if err != nil {
		log.Fatal(err)
	}

	httpMux := http.NewServeMux()
	for _, route := range bi.Routes {
		httpMux.HandleFunc(route.Path, func(writer http.ResponseWriter, request *http.Request) {
			// HandleFunc gets defined when the server starts, dispatch runs when a request is received
			r := bi.Dispatch(request, packDir)
			fb, err := ioutil.ReadFile(path.Join(packDir, integrationName, r.ResponseFile))
			if err != nil {
				sendError(writer, api.ErrorMessage(fmt.Sprintf("Failed to read: %v", r.ResponseFile)))
			}
			_, err = writer.Write(fb)
		})

	}

	err = http.ListenAndServe(addr, httpMux)
	if err != nil {
		log.Fatal(err)
	}
}

func (bi *BaseIntegration) Dispatch(request *http.Request, packDir string) Route {
	// Used at runtime
	r := bi.GetRoute(request.URL.Path)
	return r
}

func sendError(writer http.ResponseWriter, b []byte) {
	writer.WriteHeader(500)
	_, _ = writer.Write(b)
}

type Route struct {
	Path         string
	ResponseFile string
	ResponseCode int
}
