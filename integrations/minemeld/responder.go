package minemeld

/*
Implements a mock for the Palo Alto Minemeld integration
*/

import (
	"encoding/json"
	"github.com/adambaumeister/moxsoar/integrations"
	"io/ioutil"
	"log"
	"net/http"
	"path"
)

type Minemeld struct {
	integrations.BaseIntegration
}

func (i *Minemeld) Start(contentDir string, addr string) {
	b, err := ioutil.ReadFile(path.Join(contentDir, "minemeld", "routes.json"))
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(b, &i.BaseIntegration.Routes)

	if err != nil {
		log.Fatal(err)
	}

	httpMux := http.NewServeMux()
	for _, route := range i.Routes {
		// This doesn't work - The routes overwrite the other ones
		httpMux.HandleFunc(route.Path, func(writer http.ResponseWriter, request *http.Request) {
			// HandleFunc gets defined when the server starts, dispatch runs when a request is received
			r := i.Dispatch(request, contentDir)
			fb, err := ioutil.ReadFile(path.Join(contentDir, "minemeld", r.ResponseFile))
			if err != nil {
				log.Fatal(err)
			}
			_, err = writer.Write(fb)
		})

	}

	err = http.ListenAndServe(addr, httpMux)
	if err != nil {
		log.Fatal(err)
	}

}

func (m *Minemeld) Dispatch(request *http.Request, contentdir string) integrations.Route {
	// Used at runtime
	r := m.BaseIntegration.GetRoute(request.URL.Path)
	return r
}
