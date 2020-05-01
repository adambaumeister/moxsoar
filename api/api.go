package api

import (
	"encoding/json"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"time"
)

var users = map[string]string{
	"user1": "password1",
}

var jwtKey = []byte("FakeKeySon!")

func Start(addr string) {
	httpMux := http.NewServeMux()
	s := http.Server{Addr: addr, Handler: httpMux}

	httpMux.HandleFunc("/auth", auth)

	s.ListenAndServe()
	err := s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}

}

func auth(writer http.ResponseWriter, request *http.Request) {
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
