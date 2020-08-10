package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/adambaumeister/moxsoar/integrations"
	"github.com/adambaumeister/moxsoar/pack"
	"github.com/adambaumeister/moxsoar/settings"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const DEFAULT_PACK = "moxsoar-content"
const DEFAULT_REPO = "https://github.com/adambaumeister/moxsoar-content.git"

func getApiTest() (*api, []*http.Cookie) {
	/*
		API Handler fixture
	*/
	defaultAdminUser := User{
		Credentials: Credentials{
			Username: "admin",
			Password: "admin",
		},
		Name: "Default Administrative User",
	}

	pi := pack.GetPackIndex(os.Getenv("TEST_CONTENTDIR"))
	// Pull the default content repository
	p, err := pi.GetOrClone(DEFAULT_PACK, DEFAULT_REPO)
	if err != nil {
		log.Fatal(fmt.Sprintf("Could not load default pack name %s during startup (%v)!", DEFAULT_PACK, err))
	}
	settings := settings.Settings{
		DisplayHost: "0.0.0.0",
		Address:     "http://127.0.0.1:9201",
	}

	rc := pack.GetRunConfig(p.Path, &settings)
	_, _ = pi.ActivatePack(p.Name)

	rc.Prepare()

	a := api{
		PackIndex: pi,
		Users: map[string]*User{
			"admin": &defaultAdminUser,
		},
		RunConfig: rc,
	}

	// Auth to the API
	authReq, err := json.Marshal(defaultAdminUser.Credentials)
	authReqBytes := bytes.NewBuffer(authReq)
	req, err := http.NewRequest("POST", "/api/auth", authReqBytes)
	if err != nil {
		log.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(a.auth)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		log.Fatal(fmt.Printf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK))
	}
	return &a, rr.Result().Cookies()
}

func TestSplitStringToPack(t *testing.T) {
	var packName string
	var integrationName string
	var unused string
	parseArray := []*string{&unused, &unused, &unused, &packName, &integrationName}

	err := parsePath("/api/packs/packname/intname", parseArray)
	if err != nil {
		t.Fail()
	}

	if packName != "packname" {
		t.Fail()
	}
}

func TestApi_GET_PackRequest(t *testing.T) {
	// Get cookies and the API object
	a, c := getApiTest()
	req, err := http.NewRequest("GET", "/api/packs/moxsoar-content/minemeld/route/0", nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, cookie := range c {
		req.AddCookie(cookie)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(a.PackRequest)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		e := Error{}
		_ = json.NewDecoder(rr.Body).Decode(&e)
		t.Errorf("handler returned wrong status code: got %v want %v\nResponse: %v\n",
			status, http.StatusOK, e.Message)
	}
}

func TestApi_POST_PackRequest(t *testing.T) {
	// Get cookies and the API object
	a, c := getApiTest()

	method := integrations.Method{
		HttpMethod:     "GET",
		ResponseFile:   "testmethod.json",
		ResponseCode:   200,
		ResponseString: "{\"testing\": \"this\"}",
	}
	route := integrations.Route{
		Path: "/test/path/again",
		Methods: []*integrations.Method{
			&method,
		},
	}
	routeMessage := AddRoute{
		Route: &route,
	}

	b, err := json.Marshal(routeMessage)
	req, err := http.NewRequest(http.MethodPost, "/api/packs/moxsoar-content/minemeld/route", bytes.NewBuffer(b))
	if err != nil {
		t.Fatal(err)
	}

	for _, cookie := range c {
		req.AddCookie(cookie)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(a.PackRequest)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		e := Error{}
		_ = json.NewDecoder(rr.Body).Decode(&e)
		t.Errorf("handler returned wrong status code: got %v want %v\nResponse: %v\n",
			status, http.StatusOK, e.Message)
	}
	sm := StatusMessage{}
	_ = json.NewDecoder(rr.Body).Decode(&sm)

	fmt.Printf(sm.Message + "\n")

}
