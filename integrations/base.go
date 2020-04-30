package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/adambaumeister/moxsoar/api"
	"github.com/adambaumeister/moxsoar/tracker"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"regexp"
)

type BaseIntegration struct {
	Routes map[string]Route

	Ctx context.Context

	Tracker tracker.DebugTracker
}

func (bi *BaseIntegration) GetRoute(url string, method string) Method {

	// Get a route for a given request
	for _, route := range bi.Routes {
		// Url routing using regex. What will these crazy kooks think of next aye?
		m, _ := regexp.MatchString(route.Path, url)

		if m {

			// If the route doesn't specify methods, and the path matches, return it
			if route.Methods == nil {

				return Method{
					ResponseFile: route.ResponseFile,
					ResponseCode: route.ResponseCode,
					HttpMethod:   method,
				}
			}

			// If the route does specify methods, try to match the method against the provided
			for _, rmethod := range route.Methods {
				fmt.Printf("%v, %v try match %v\n", url, method, rmethod.HttpMethod)

				if method == rmethod.HttpMethod {
					return rmethod
				}
			}

		}
	}

	// If nothing matches, return this.
	return Method{
		ResponseFile: "default.json",
		ResponseCode: 200,
		HttpMethod:   method,
	}
}

func defaultHandler(_ http.ResponseWriter, request *http.Request) {
	t := tracker.GetDebugTracker()
	t.Track(request)
}

func (bi *BaseIntegration) Start(integrationName string, packDir string, addr string) {
	/*
		Register the HTTP handlers and start the integration
	*/
	bi.ReadRoutes(path.Join(packDir, integrationName, "routes.json"))

	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/", defaultHandler)
	for _, route := range bi.Routes {
		httpMux.HandleFunc(route.Path, func(writer http.ResponseWriter, request *http.Request) {

			// Read the route table within the http handler, such that it is dynamic
			bi.ReadRoutes(path.Join(packDir, integrationName, "routes.json"))

			// HandleFunc gets defined when the server starts, dispatch runs when a request is received
			r := bi.Dispatch(request, packDir)
			fb, err := ioutil.ReadFile(path.Join(packDir, integrationName, r.ResponseFile))
			if err != nil {
				sendError(writer, api.ErrorMessage(fmt.Sprintf("Failed to read: %v", r.ResponseFile)))
			}
			_, err = writer.Write(fb)
		})

	}

	err := http.ListenAndServe(addr, httpMux)
	if err != nil {
		log.Fatal(err)
	}
}

func (bi *BaseIntegration) ReadRoutes(routeFile string) {
	// Read the route table on invocation, such that it is dynamic
	b, err := ioutil.ReadFile(routeFile)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(b, &bi.Routes)

	if err != nil {
		log.Fatal(err)
	}

}

func (bi *BaseIntegration) Dispatch(request *http.Request, packDir string) Method {
	// Used at runtime
	m := bi.GetRoute(request.URL.Path, request.Method)
	return m
}

func sendError(writer http.ResponseWriter, b []byte) {
	writer.WriteHeader(500)
	_, _ = writer.Write(b)
}

type Route struct {
	Path         string
	ResponseFile string
	ResponseCode int
	Methods      []Method
}

type Method struct {
	HttpMethod   string
	ResponseFile string
	ResponseCode int
}
