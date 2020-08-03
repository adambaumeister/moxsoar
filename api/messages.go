package api

import (
	"encoding/json"
	"github.com/adambaumeister/moxsoar/integrations"
	"github.com/adambaumeister/moxsoar/pack"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
)

type Error struct {
	Message string
}

type StatusMessage struct {
	Message string
}

type LoginMessage struct {
	Message  string
	Username string
	Settings Settings
}

/*
Auth messages
*/
type Credentials struct {
	Password string
	Username string
}

type Claims struct {
	Username string
	jwt.StandardClaims
}

type AddUserMessage struct {
	Message string
}

/*
Pack messages
*/
type GetPacksResponse struct {
	Packs []*pack.Pack
}

type CloneRequest struct {
	PackName string
	Repo     string
}

type CloneResponse struct {
	Message string
}

type ActivateRequest struct {
	PackName string
}

type ActivateResponse struct {
	Message string
}

type UpdateRequest struct {
	PackName string
}

/*
INtegration messages
*/

type GetIntegration struct {
	Integration string
	Addr        string
	Port        string
	Routes      []*integrations.Route
}

type GetRunnerResponse struct {
	RunConfig *pack.RunConfig
}

func ErrorMessage(m string) []byte {
	e := Error{
		Message: m,
	}

	return MarshalToJson(e)

}

type GetRouteRequest struct {
	routeid int
}

type GetRoute struct {
	Route          *integrations.Route
	ResponseString string
}

type AddRoute struct {
	Route *integrations.Route
}

type DeleteRoute struct {
	Path string
}

func MarshalToJson(m interface{}) []byte {
	b, err := json.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}

	return b
}

func SendJsonResponse(m interface{}, writer http.ResponseWriter) error {
	b, err := json.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(b)
	return err
}

func SendError(err error, writer http.ResponseWriter, errcode int) {
	if err != nil {
		writer.WriteHeader(errcode)
		r := ErrorMessage(err.Error())
		_, _ = writer.Write(r)
		return
	}
}
