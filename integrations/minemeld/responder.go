package minemeld

/*
Implements a mock for the Palo Alto Minemeld integration
*/

import (
	"encoding/json"
	"github.com/adambaumeister/moxsoar/integrations"
	"github.com/adambaumeister/moxsoar/runner"
	"io/ioutil"
	"log"
	"net/http"
	"path"
)

type Minemeld struct {
	integrations.BaseIntegration
}

func (i *Minemeld) Start(config runner.RunConfig) {
	b, err := ioutil.ReadFile(path.Join(config.Runner.ContentDir, "minemeld", "routes.json"))
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(b, &i.BaseIntegration.Routes)

	if err != nil {
		log.Fatal(err)
	}

	for _, route := range i.Routes {
		http.HandleFunc(route.Path, func(writer http.ResponseWriter, request *http.Request) {
			fb, err := ioutil.ReadFile(path.Join(config.Runner.ContentDir, "minemeld", route.ResponseFile))
			if err != nil {
				log.Fatal(err)
			}
			_, err = writer.Write(fb)
		})

	}

	a := config.Runner.GetAddress()
	http.ListenAndServe(a, nil)

}
