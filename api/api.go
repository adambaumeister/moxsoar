package api

import (
	"encoding/json"
	"github.com/adambaumeister/moxsoar/pack"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"time"
)

var users = map[string]string{
	"user1": "password1",
}

var jwtKey = []byte("FakeKeySon!")

type api struct {
	PackIndex *pack.PackIndex
}

func Start(addr string, pi *pack.PackIndex) {

	a := api{
		PackIndex: pi,
	}
	httpMux := http.NewServeMux()
	s := http.Server{Addr: addr, Handler: httpMux}

	httpMux.HandleFunc("/auth", a.auth)
	httpMux.HandleFunc("/packs", a.getPacks)

	s.ListenAndServe()
	err := s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}

func (a *api) auth(writer http.ResponseWriter, request *http.Request) {
	// Get an authentication request message JSON
	var creds Credentials
	err := json.NewDecoder(request.Body).Decode(&creds)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check the PW matches with what's in the DB
	expectedPassword, ok := users[creds.Username]

	if !ok || expectedPassword != creds.Password {
		writer.WriteHeader(http.StatusUnauthorized)
		return
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

func checkAuth(writer http.ResponseWriter, request *http.Request) string {
	c, err := request.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			writer.WriteHeader(http.StatusUnauthorized)
			return ""
		}
		// For any other type of error, return a bad request status
		writer.WriteHeader(http.StatusBadRequest)
		return ""
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
			return ""
		}
		writer.WriteHeader(http.StatusBadRequest)
		return ""
	}
	if !tkn.Valid {
		writer.WriteHeader(http.StatusUnauthorized)
		return ""
	}

	// Finally, return the welcome message to the user, along with their
	// username given in the token
	return claims.Username
}

func (a *api) getPacks(writer http.ResponseWriter, request *http.Request) {
	// Validate the user is authenticated
	checkAuth(writer, request)

	packs := a.PackIndex.Packs
	r := GetPacksResponse{
		Packs: packs,
	}

	b := MarshalToJson(r)
	writer.Write(b)
}
