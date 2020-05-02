package api

import (
	"encoding/json"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"os"
)

type Hash struct {
}

type User struct {
	Credentials Credentials

	Name string
}

func (c *Hash) Generate(s string) (string, error) {
	saltedBytes := []byte(s)
	hashedBytes, err := bcrypt.GenerateFromPassword(saltedBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	hash := string(hashedBytes[:])
	return hash, nil
}

//Compare string to generated hash
func (c *Hash) Compare(hash string, s string) error {
	incoming := []byte(s)
	existing := []byte(hash)

	return bcrypt.CompareHashAndPassword(existing, incoming)
}

type JSONPasswordDB struct {
	Path string
}

func (jpdb *JSONPasswordDB) Write(db map[string]*User) error {
	b, err := json.Marshal(db)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(jpdb.Path, b, os.FileMode(600))
	if err != nil {
		return err
	}
	return nil
}
