package api

import (
	"encoding/json"
	"github.com/adambaumeister/moxsoar/pack"
	"github.com/dgrijalva/jwt-go"
	"log"
)

type Error struct {
	Message string
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

type GetRunnerResponse struct {
	RunConfig pack.RunConfig
}

func ErrorMessage(m string) []byte {
	e := Error{
		Message: m,
	}

	return MarshalToJson(e)

}

func MarshalToJson(m interface{}) []byte {
	b, err := json.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}

	return b
}
