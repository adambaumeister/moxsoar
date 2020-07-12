package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/adambaumeister/moxsoar/tracker"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
)

const ROUTE_FILE = "routes.json"

type BaseIntegration struct {
	Routes   []*Route
	Addr     string               `json:"none"`
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
	s := &http.Server{Addr: addr, Handler: httpMux}

	httpMux.HandleFunc("/", defaultHandler)
	for _, route := range bi.Routes {
		httpMux.HandleFunc(route.Path, func(writer http.ResponseWriter, request *http.Request) {

			t := tracker.GetDebugTracker()
			t.Track(request)

			// Read the route table within the http handler, such that it is dynamic
			bi.ReadRoutes(path.Join(packDir, integrationName, ROUTE_FILE))

			// HandleFunc gets defined when the server starts, dispatch runs when a request is received
			r := bi.Dispatch(request, packDir)

			// If any cookies, write those firsts
			for _, c := range r.Cookies {
				http.SetCookie(writer, c)
			}

			// Write the response data
			fb, err := ioutil.ReadFile(path.Join(packDir, integrationName, r.ResponseFile))
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
			}
			_, err = writer.Write(fb)
		})

	}

	// This starts a func in the background that just sits listening for input on that channel, then executes
	// very cool
	go func() {
		<-bi.ExitChan

		if err := s.Shutdown(context.Background()); err != nil {
			log.Fatalf("Could not gracefully shutdown the server: %v\n", err)
		}
		bi.ExitChan <- true
	}()

	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Printf("HTTP server ListenAndServe: %v", err)
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

func (bi *BaseIntegration) AddRoute(route *Route) error {
	// First, check that this path doesn't already exist
	for _, r := range bi.Routes {
		if route.Path == r.Path {
			return fmt.Errorf("Path %v already exists.", route.Path)
		}
	}

	// Add a route{} object to both the routes for this integration, and write the string as a file
	for _, method := range route.Methods {
		jsonFile := path.Join(bi.PackDir, bi.Name, method.ResponseFile)
		err := ioutil.WriteFile(jsonFile, []byte(method.ResponseString), 755)
		if err != nil {
			return err
		}
	}

	routeFile := path.Join(bi.PackDir, bi.Name, ROUTE_FILE)
	bi.Routes = append(bi.Routes, route)
	b, err := json.Marshal(bi)
	if err != nil {
		return fmt.Errorf("Failed to marshal provided route object.")
	}
	err = ioutil.WriteFile(routeFile, b, 755)
	if err != nil {
		return fmt.Errorf("Could not save route file.")
	}
	return nil
}

func (bi *BaseIntegration) DeleteRoute(pathName string) error {
	// Replace the route list with the list - the one we are deleting
	var route *Route
	var newRoutes []*Route
	for _, r := range bi.Routes {
		if pathName == r.Path {
			route = r
		} else {
			newRoutes = append(newRoutes, r)
		}
	}

	if route == nil {
		return (fmt.Errorf("Path not found: %v", pathName))
	}

	// Find all the response files
	var files []string
	for _, method := range route.Methods {
		files = append(files, path.Join(bi.PackDir, bi.Name, method.ResponseFile))
	}

	// Delete all the response files
	for _, f := range files {
		_, err := os.Stat(f)
		// Only try unlinking if they exist, duh
		if !os.IsNotExist(err) {
			err := os.Remove(f)
			if err != nil {
				return err
			}
		}

	}

	// write the new integration (route) file
	bi.Routes = newRoutes
	routeFile := path.Join(bi.PackDir, bi.Name, ROUTE_FILE)
	b, err := json.Marshal(bi)
	if err != nil {
		return fmt.Errorf("Failed to marshal provided route object.")
	}
	err = ioutil.WriteFile(routeFile, b, 755)
	if err != nil {
		return fmt.Errorf("Could not save route file.")
	}
	return nil
}

func (bi *BaseIntegration) Dispatch(request *http.Request, packDir string) *Method {
	// Used at runtime
	m := bi.GetRoute(request.URL.Path, request.Method)
	return m
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

	Cookies []*http.Cookie
}
