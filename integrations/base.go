package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/adambaumeister/moxsoar/tracker"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"regexp"
	"strings"
)

const ROUTE_FILE = "routes.json"

type BaseIntegration struct {
	Routes   []*Route
	Addr     string
	ExitChan chan bool            `json:"none"`
	Tracker  tracker.DebugTracker `json:"none"`
	PackDir  string               `json:"none"`

	Name string
}

func (bi *BaseIntegration) GetRoute(url string, method string) *Method {

	// Get a route for a given request
	for _, route := range bi.Routes {
		if strings.Contains(url, route.Path) {

			// If the route doesn't specify methods, and the path matches, return it
			if route.Methods == nil {

				return &Method{
					ResponseFile: route.ResponseFile,
					ResponseCode: route.ResponseCode,
					HttpMethod:   method,
				}
			}

			// If the route does specify methods, try to match the method against the provided
			for _, rmethod := range route.Methods {
				if method == rmethod.HttpMethod {
					//  MatchRegex allows more granular (regex) matching for making routing decisions
					if rmethod.MatchRegex != "" {
						m, _ := regexp.MatchString(rmethod.MatchRegex, url)
						if m {
							return rmethod
						}
					}
					return rmethod
				}
			}
		}
	}

	// If nothing matches, return this.
	return &Method{
		ResponseFile: "default.json",
		ResponseCode: 200,
		HttpMethod:   method,
	}
}

func defaultHandler(_ http.ResponseWriter, request *http.Request) {
	t := tracker.GetDebugTracker()
	t.Track(request)
}

func (bi *BaseIntegration) Start(integrationName string) {
	/*
		Register the HTTP handlers and start the integration
	*/
	packDir := bi.PackDir
	addr := bi.Addr
	bi.ReadRoutes(path.Join(packDir, integrationName, ROUTE_FILE))

	httpMux := http.NewServeMux()
	s := http.Server{Addr: addr, Handler: httpMux}

	// This starts a func in the background that just sits listening for input on that channel, then executes
	// very cool
	go func() {
		<-bi.ExitChan
		fmt.Printf("requested shutdown\n")

		if err := s.Shutdown(context.Background()); err != nil {
			log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
	}()

	httpMux.HandleFunc("/", defaultHandler)
	for _, route := range bi.Routes {
		httpMux.HandleFunc(route.Path, func(writer http.ResponseWriter, request *http.Request) {

			t := tracker.GetDebugTracker()
			t.Track(request)

			// Read the route table within the http handler, such that it is dynamic
			bi.ReadRoutes(path.Join(packDir, integrationName, ROUTE_FILE))

			// HandleFunc gets defined when the server starts, dispatch runs when a request is received
			r := bi.Dispatch(request, packDir)
			fb, err := ioutil.ReadFile(path.Join(packDir, integrationName, r.ResponseFile))
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
			}
			_, err = writer.Write(fb)
		})

	}

	err := s.ListenAndServe()
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
	err = json.Unmarshal(b, &bi)

	if err != nil {
		log.Fatal(err)
	}

	for _, route := range bi.Routes {
		if route.Methods == nil {
			m := Method{
				ResponseFile: route.ResponseFile,
				ResponseCode: route.ResponseCode,
				HttpMethod:   "GET",
			}
			route.Methods = []*Method{&m}
		}
	}

}

func (bi *BaseIntegration) Dispatch(request *http.Request, packDir string) *Method {
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
	Methods      []*Method
}

type Method struct {
	HttpMethod   string
	ResponseFile string
	ResponseCode int
	MatchRegex   string

	ResponseString string
}
