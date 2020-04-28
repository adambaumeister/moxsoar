package minemeld

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

func (i *Minemeld) Start(d string, addr string) {
	b, err := ioutil.ReadFile(path.Join(d, "minemeld", "routes.json"))
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(b, &i.BaseIntegration.Routes)

	if err != nil {
		log.Fatal(err)
	}

	for _, route := range i.Routes {
		http.HandleFunc(route.Path, func(writer http.ResponseWriter, request *http.Request) {
			fb, err := ioutil.ReadFile(path.Join(d, "minemeld", route.ResponseFile))
			if err != nil {
				log.Fatal(err)
			}
			_, err = writer.Write(fb)
		})

	}

	http.ListenAndServe(addr, nil)
}
