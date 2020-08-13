package integrations

import (
	"github.com/adambaumeister/moxsoar/settings"
	"strings"
)

const VAR_BUILTINS_DISPLAYHOST = "$(VAR_DISPLAYHOST)"

func SubVariables(fb []byte, settings *settings.Settings) []byte {
	s := string(fb)
	// Sub the default stuff
	s = strings.Replace(s, VAR_BUILTINS_DISPLAYHOST, settings.DisplayHost, -1)

	return []byte(s)
}
