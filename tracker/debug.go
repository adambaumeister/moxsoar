package tracker

import (
	"fmt"
	"net/http"
)

/*
Debug tracker simply dumps request output to stdout.
uSeful for development but probably not in production.
*/

type DebugTracker struct {
}

func GetDebugTracker() *DebugTracker {
	return &DebugTracker{}
}

func (t *DebugTracker) Track(request *http.Request) {
	fmt.Printf("DEBUG: request recevied at path %v\n", request.URL)
	b := []byte{}
	_, err := request.Body.Read(b)
	// Unparsable request body
	if err == nil {
		// Log only if there is a body.
		fmt.Printf("## BODY ##\n\n")
		fmt.Printf("%v\n", string(b))
	}
}
