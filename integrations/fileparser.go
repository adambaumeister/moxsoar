package integrations

import (
	"fmt"
	"github.com/adambaumeister/moxsoar/settings"
	"strings"
)

const VAR_BUILTINS_DISPLAYHOST = "$(VAR_DISPLAYHOST)"

func SubVariables(fb []byte, settings *settings.Settings) []byte {
	// This function replaces any strings that match the variable syntax $(blah) in teh response body

	s := string(fb)
	// Sub the default stuff
	s = strings.Replace(s, VAR_BUILTINS_DISPLAYHOST, settings.DisplayHost, -1)

	for k, v := range settings.Variables {
		kString := fmt.Sprintf("$(%v)", k)
		s = strings.Replace(s, kString, v, -1)
	}
	return []byte(s)
}
