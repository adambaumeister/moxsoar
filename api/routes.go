package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
)

func (a *api) PackRequest(writer http.ResponseWriter, request *http.Request) {
	// Validate the user is authenticated
	_, tkn := checkAuth(writer, request)
	if tkn == nil {
		return
	}

	var packName string
	var integrationName string
	var unused string
	var command string

	parseArray := []*string{&unused, &unused, &unused, &packName, &integrationName, &command}
	var err error

	err = parsePath(request.URL.Path, parseArray)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	rc := a.RunConfig
	var r interface{}

	// Command to use
	if command != "" {
		switch command {
		// Route commands, add, get, etc.
		case "route":
			switch request.Method {
			// GET
			case http.MethodGet:
				routeReq := GetRouteRequest{}
				err := json.NewDecoder(request.Body).Decode(&routeReq)
				if err != nil {
					writer.WriteHeader(http.StatusBadRequest)
					r := ErrorMessage("Malformed request.")
					_, _ = writer.Write(r)
					return
				}

				routeId := routeReq.routeid
				i := getIntegrationObject(integrationName, rc)

				if i == nil {
					writer.WriteHeader(http.StatusBadRequest)
					r = Error{Message: "Integration not found"}
					return
				}
				for _, m := range i.Routes[routeId].Methods {
					fn := m.ResponseFile
					fb, err := ioutil.ReadFile(path.Join(i.PackDir, integrationName, fn))
					if err != nil {
						writer.WriteHeader(http.StatusInternalServerError)
						r = Error{Message: err.Error()}
						return
					}

					m.ResponseString = string(fb)

				}

				r = GetRoute{
					Route: i.Routes[routeId],
				}
			// POST
			case http.MethodPost:
				writer.WriteHeader(http.StatusNotImplemented)
				r := ErrorMessage("Not yet implemented.")
				_, _ = writer.Write(r)
				return
			}
		}
		// We've requested the entire integration
	} else if integrationName != "" {
		r, err = getIntegration(integrationName, rc)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			r = Error{Message: err.Error()}
			return
		}
	} else {
		r = GetRunnerResponse{
			RunConfig: rc,
		}
	}
	b := MarshalToJson(r)
	_, err = writer.Write(b)
	if err != nil {
		panic("Failed to write response http")
	}
}

func parsePath(str string, parseArray []*string) error {
	s := strings.Split(str, "/")
	ps := parseArray
	idx := 0
	if len(parseArray) < len(s) {
		return fmt.Errorf("Invalid path.")
	}

	for _, pathStr := range s {
		*ps[idx] = pathStr
		idx = idx + 1
	}
	return nil
}
