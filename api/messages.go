package api

import (
	"encoding/json"
	"log"
)

type Error struct {
	Message string
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
