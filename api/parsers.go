package api

import (
	"fmt"
	"strings"
)

func (routeMessage *AddRoute) Parse() error {
	route := routeMessage.Route
	if route.Path == "" {
		return fmt.Errorf("Provided route is missing a path value.")
	}
	for _, method := range route.Methods {
		// If no response filename is proved, generate one based on the path name
		if method.ResponseFile == "" {
			method.ResponseFile = strings.Replace(route.Path, "/", "_", -1)[1:]
		}
	}
	return nil
}
