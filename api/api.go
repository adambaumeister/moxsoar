package api

import (
	"encoding/json"
	"fmt"
	"github.com/adambaumeister/moxsoar/pack"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"strings"
	"time"
)

var jwtKey = []byte("FakeKeySon!")

type api struct {
	PackIndex *pack.PackIndex

	Users map[string]*User

	UserDB *JSONPasswordDB
}

func Start(addr string, pi *pack.PackIndex, userfile string) {

	jpdb := JSONPasswordDB{
		Path: userfile,
	}

	defaultAdminUser := User{
		Credentials: Credentials{
			Username: "admin",
			Password: "admin",
		},
		Name: "Default Administrative User",
	}

	a := api{
		PackIndex: pi,
		Users: map[string]*User{
			"admin": &defaultAdminUser,
		},
		UserDB: &jpdb,
	}
	httpMux := http.NewServeMux()
	s := http.Server{Addr: addr, Handler: httpMux}

	httpMux.HandleFunc("/auth", a.auth)
	httpMux.HandleFunc("/packs", a.getPacks)
	httpMux.HandleFunc("/packs/", a.getPack)
	httpMux.HandleFunc("/adduser", a.addUser)
	httpMux.HandleFunc("/refreshauth", refreshAuth)

	err := s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}

func (a *api) auth(writer http.ResponseWriter, request *http.Request) {
	/*
		Authenticate the API client
	*/

	// Get an authentication request message JSON
	var creds Credentials
	err := json.NewDecoder(request.Body).Decode(&creds)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check the PW matches with what's in the DB
	user, ok := a.Users[creds.Username]
	c := Hash{}

	checkHashResult := c.Compare(user.Credentials.Password, creds.Password)
	if checkHashResult != nil {
		// If hash doesn't match, check cleartext
		// This lets us populate the default admin password easier
		if !ok || user.Credentials.Password != creds.Password {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	expirationTime := time.Now().Add(30 * time.Minute)
	// Create the JWT claims, which includes the username and expiry time
	claims := &Claims{
		Username: creds.Username,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// This is where we sign the token
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
	}

	http.SetCookie(writer, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

}

func checkAuth(writer http.ResponseWriter, request *http.Request) (*Claims, *jwt.Token) {
	/*
		Validate the auth ticket is still valid
	*/
	c, err := request.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			writer.WriteHeader(http.StatusUnauthorized)
			return nil, nil
		}
		// For any other type of error, return a bad request status
		writer.WriteHeader(http.StatusBadRequest)
		return nil, nil
	}

	// Get the JWT string from the cookie
	tknStr := c.Value

	// Initialize a new instance of `Claims`
	claims := &Claims{}

	// Parse the JWT string and store the result in `claims`.
	// Note that we are passing the key in this method as well. This method will return an error
	// if the token is invalid (if it has expired according to the expiry time we set on sign in),
	// or if the signature does not match
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			writer.WriteHeader(http.StatusUnauthorized)
			return nil, nil
		}
		writer.WriteHeader(http.StatusBadRequest)
		return nil, nil
	}
	if !tkn.Valid {
		writer.WriteHeader(http.StatusUnauthorized)
		return nil, nil
	}

	// Finally, return the welcome message to the user, along with their
	// username given in the token
	return claims, tkn
}

func refreshAuth(writer http.ResponseWriter, request *http.Request) {
	// Refresh the token attached to a user

	// First check the autth is actually valid and error out of it isn't
	_, tkn := checkAuth(writer, request)
	if tkn == nil {
		return
	}

	claims := &Claims{}
	newExpTime := time.Now().Add(30 * time.Minute)

	claims.ExpiresAt = newExpTime.Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(writer, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: newExpTime,
	})
}

func (a *api) getPacks(writer http.ResponseWriter, request *http.Request) {
	/*
		Get all the content packs on the system
	*/

	// Validate the user is authenticated
	_, tkn := checkAuth(writer, request)
	if tkn == nil {
		return
	}

	packs := a.PackIndex.Packs
	r := GetPacksResponse{
		Packs: packs,
	}

	b := MarshalToJson(r)
	writer.Write(b)
}

func (a *api) getPack(writer http.ResponseWriter, request *http.Request) {
	/*
		Functions related to pack manipulation
	*/

	// Validate the user is authenticated
	_, tkn := checkAuth(writer, request)
	if tkn == nil {
		return
	}

	s := strings.Split(request.URL.Path, "/")
	if len(s) < 2 {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	packName := s[2]

	p, err := a.PackIndex.GetPackName(packName)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	rc := pack.GetRunConfig(p.Path)

	var r interface{}
	// We've asked for something other than just the pack itself
	if len(s) == 4 {
		integrationName := s[3]
		r, err = getIntegration(integrationName, rc)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			r = Error{Message: err.Error()}
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

func (a *api) addUser(writer http.ResponseWriter, request *http.Request) {

	// Validate the user is authenticated
	_, tkn := checkAuth(writer, request)
	if tkn == nil {
		return
	}

	var creds Credentials
	err := json.NewDecoder(request.Body).Decode(&creds)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	c := Hash{}
	hpwd, err := c.Generate(creds.Password)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	user := User{
		Credentials: Credentials{
			Username: creds.Username,
			Password: hpwd,
		},
	}

	a.Users[creds.Username] = &user

	err = a.UserDB.Write(a.Users)
	if err != nil {
		fmt.Printf("error writing file: %v", err)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	r := AddUserMessage{
		Message: fmt.Sprintf("Added user: %v", creds.Username),
	}

	b := MarshalToJson(r)
	_, err = writer.Write(b)
	if err != nil {
		panic("Failed to write response http")
	}

}

func getIntegration(name string, rc pack.RunConfig) (*GetIntegration, error) {
	ints := rc.GetIntegrations()

	r := GetIntegration{}
	for _, integration := range ints {
		if integration.Name == name {
			r.Routes = integration.Routes
			r.Integration = name
			return &r, nil
		}
	}

	return nil, fmt.Errorf("Integration %v not found", name)
}
