package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
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
	var id string

	parseArray := []*string{&unused, &unused, &unused, &packName, &integrationName, &command, &id}
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
				routeId, err := strconv.Atoi(id)
				if err != nil {

					writer.WriteHeader(http.StatusBadRequest)
					r := ErrorMessage("No ID for Route specified!")
					_, _ = writer.Write(r)
					return
				}
				i := getIntegrationObject(integrationName, rc)

				if i == nil {
					writer.WriteHeader(http.StatusBadRequest)
					r := ErrorMessage(fmt.Sprintf("Integration %v not found.", integrationName))
					_, _ = writer.Write(r)
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

				routeMessage := AddRoute{}
				err := json.NewDecoder(request.Body).Decode(&routeMessage)

				if err != nil {
					writer.WriteHeader(http.StatusBadRequest)
					r := ErrorMessage(fmt.Sprintf("Invalid JSON route provided: %v", err))
					_, _ = writer.Write(r)
					return
				}
				// Validate the incoming JSON message.
				err = routeMessage.Parse()
				if err != nil {
					SendError(err, writer, 400)
					return
				}

				i := getIntegrationObject(integrationName, rc)
				err = i.AddRoute(routeMessage.Route)
				if err != nil {
					writer.WriteHeader(http.StatusBadRequest)
					r := ErrorMessage(err.Error())
					_, _ = writer.Write(r)
					return
				}

				a.RunConfig.Restart()

				r = StatusMessage{
					Message: "Sucessfully added route.",
				}
			case http.MethodDelete:
				routeMessage := DeleteRoute{}
				err := json.NewDecoder(request.Body).Decode(&routeMessage)
				if err != nil {
					writer.WriteHeader(http.StatusBadRequest)
					r := ErrorMessage(fmt.Sprintf("Invalid JSON delete message provided: %v", err))
					_, _ = writer.Write(r)
					return
				}
				i := getIntegrationObject(integrationName, rc)
				err = i.DeleteRoute(routeMessage.Path)
				if err != nil {
					writer.WriteHeader(http.StatusBadRequest)
					r := ErrorMessage(fmt.Sprintf("Delete route failed (%v)", err))
					_, _ = writer.Write(r)
					return
				}
				r = StatusMessage{
					Message: fmt.Sprintf("Sucessfully deleted %v", routeMessage.Path),
				}
			}
		}

		// We've requested the entire integration
	} else if integrationName != "" {
		switch request.Method {
		// GET
		case http.MethodGet:

			r, err = getIntegration(integrationName, rc)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				r = Error{Message: err.Error()}
				return
			}
		// POST: Add a new integration
		case http.MethodPost:
			err = rc.AddIntegration(integrationName)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				r := ErrorMessage(fmt.Sprintf("Could not add integration: %v", err))
				_, _ = writer.Write(r)
				return
			}
			r = StatusMessage{
				Message: fmt.Sprintf("Integration %v added!", integrationName),
			}
		case http.MethodDelete:
			err = rc.DeleteIntegration(integrationName)
			if err != nil {
				writer.WriteHeader(http.StatusBadRequest)
				r := ErrorMessage(fmt.Sprintf("Could not delete integration: %v", err))
				_, _ = writer.Write(r)
				return
			}
			r = StatusMessage{
				Message: fmt.Sprintf("Integration %v deleted.", integrationName),
			}

		}
	} else {
		r = GetRunnerResponse{
			RunConfig: rc,
		}
	}
	_ = SendJsonResponse(r, writer)
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

func (a *api) settings(writer http.ResponseWriter, request *http.Request) {
	// Validate the user is authenticated
	_, tkn := checkAuth(writer, request)
	if tkn == nil {
		return
	}

	var r interface{}
	switch request.Method {
	case http.MethodGet:
		r = a.SettingsDB.GetSettings()
	case http.MethodPost:
		s := a.SettingsDB.GetSettings()
		err := json.NewDecoder(request.Body).Decode(&s)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			r := ErrorMessage(fmt.Sprintf("Invalid JSON settings message provided: %v", err))
			_, _ = writer.Write(r)
			return
		}

		err = a.SettingsDB.Save(*s)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			r := ErrorMessage(fmt.Sprintf("Failed to write settings file: %v", err))
			_, _ = writer.Write(r)
			return
		}

		r = StatusMessage{
			Message: "Saved the server settings.",
		}
		a.RunConfig.UpdateSettings(*s)
		a.RunConfig.Restart()
	}
	_ = SendJsonResponse(r, writer)
}

func (a *api) VariablesRequest(writer http.ResponseWriter, request *http.Request) {
	// Validate the user is authenticated
	_, tkn := checkAuth(writer, request)
	if tkn == nil {
		return
	}

	var r interface{}
	switch request.Method {
	case http.MethodPost:
		v := AddVariableRequest{}
		err := json.NewDecoder(request.Body).Decode(&v)
		if err != nil {
			SendError(err, writer, http.StatusBadRequest)
			return
		}

		err = a.SettingsDB.AddVariable(v.Key, v.Value)
		if err != nil {
			SendError(err, writer, http.StatusInternalServerError)
			return
		}
		r = StatusMessage{
			Message: "Variable added.",
		}
		a.RunConfig.UpdateSettings(*a.SettingsDB.GetSettings())
		a.RunConfig.Restart()
	case http.MethodDelete:
		v := DeleteVaribleRequest{}
		err := json.NewDecoder(request.Body).Decode(&v)
		if err != nil {
			SendError(err, writer, http.StatusBadRequest)
			return
		}
		err = a.SettingsDB.DeleteVariable(v.Key)
		if err != nil {
			SendError(err, writer, http.StatusInternalServerError)
			return
		}
		r = StatusMessage{
			Message: "Variable Deleted.",
		}
		a.RunConfig.UpdateSettings(*a.SettingsDB.GetSettings())
		a.RunConfig.Restart()
	}
	_ = SendJsonResponse(r, writer)

}
