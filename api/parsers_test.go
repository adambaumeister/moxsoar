package api

import (
	"github.com/adambaumeister/moxsoar/integrations"
	"log"
	"testing"
)

func TestAddRoute_Parse(t *testing.T) {
	// parses a valid Addroute method
	testMethod := integrations.Method{
		HttpMethod:     "GET",
		ResponseFile:   "test.json",
		ResponseCode:   200,
		MatchRegex:     "",
		ResponseString: "{}",
		Cookies:        nil,
	}
	testMethods := []*integrations.Method{&testMethod}
	m := AddRoute{
		Route: &integrations.Route{
			Path:    "/test/path",
			Methods: testMethods,
		},
	}
	err := m.Parse()
	if err != nil {
		t.Fail()
	}
}

func TestAddRoute_ParseNoPath(t *testing.T) {
	// parses an invalid Addroute method
	testMethod := integrations.Method{
		HttpMethod:     "GET",
		ResponseFile:   "test.json",
		ResponseCode:   200,
		MatchRegex:     "",
		ResponseString: "{}",
		Cookies:        nil,
	}
	testMethods := []*integrations.Method{&testMethod}
	m := AddRoute{
		Route: &integrations.Route{
			Path:    "",
			Methods: testMethods,
		},
	}
	err := m.Parse()
	if err == nil {
		t.Fatal("Should raise an error about missing paht.")
	}
}

func TestAddRoute_ParseNoFilename(t *testing.T) {
	// parses an invalid Addroute method
	testMethod := integrations.Method{
		HttpMethod:     "GET",
		ResponseFile:   "",
		ResponseCode:   200,
		MatchRegex:     "",
		ResponseString: "{}",
		Cookies:        nil,
	}
	testMethods := []*integrations.Method{&testMethod}
	m := AddRoute{
		Route: &integrations.Route{
			Path:    "/test/path",
			Methods: testMethods,
		},
	}
	err := m.Parse()
	if err != nil {
		t.Fatal(err)
	}

	if testMethod.ResponseFile != "test_path" {
		log.Fatalf("Response filename not computed correctly: %v", testMethod.ResponseFile)
	}
}
