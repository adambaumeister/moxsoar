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

	for _, route := range i.Routes {
		http.HandleFunc(route.Path, func(writer http.ResponseWriter, request *http.Request) {
			fb, err := ioutil.ReadFile(path.Join(contentDir, "minemeld", route.ResponseFile))
			if err != nil {
				log.Fatal(err)
			}
			_, err = writer.Write(fb)
		})

	}

	err = http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal(err)
	}

}
