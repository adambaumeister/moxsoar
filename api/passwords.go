package api

import (
	"golang.org/x/crypto/bcrypt"
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

func GetJsonPasswordDB(path string) *JSONPasswordDB {
	jpdb := JSONPasswordDB{
		Path: path,
	}

	return &jpdb
}

func (jpdb *JSONPasswordDB) Write(map[string]User) {

}
