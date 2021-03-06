package settings

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

/*
Simple settings file, for storing user configured stuff.
Also stores the elasticsearch server
*/

type SettingsDB struct {
	Path string
}

type Settings struct {
	DisplayHost        string
	Address            string
	Username           string
	Password           string
	SSLCertificatePath string
	SSLKeyPath         string

	Variables map[string]string
}

func (s *SettingsDB) GetSettings() *Settings {
	// Defaults
	// DisplayHost is set to Localhost for dev environments
	// Address is set to the docker-compose elasticsearch container
	settings := Settings{
		DisplayHost:        "localhost",
		Address:            "http://elasticsearch:9200",
		Variables:          map[string]string{},
		SSLCertificatePath: "/certs/server.crt",
		SSLKeyPath:         "/certs/server.key",
	}
	// If it already exists, read it and return it
	if fileExists(s.Path) {
		b, err := ioutil.ReadFile(s.Path)
		if err != nil {
			log.Fatal("Failed to open settings file: %v (%v)", s.Path, err)
		}
		err = json.Unmarshal(b, &settings)
		if err != nil {
			log.Fatal(fmt.Sprintf("Failed to unmarshal settings file: %v (%v)", s.Path, err))
		}
	}

	// Otherwise, just return the default.
	return &settings
}

func (s *SettingsDB) AddVariable(k string, v string) error {
	settings := s.GetSettings()
	settings.Variables[k] = v
	err := s.Save(*settings)
	return err
}

func (s *SettingsDB) DeleteVariable(k string) error {
	settings := s.GetSettings()
	delete(settings.Variables, k)
	err := s.Save(*settings)
	return err
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func (s *SettingsDB) Save(settings Settings) error {
	b, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(s.Path, b, 0755)
	if err != nil {
		return err
	}

	return nil
}
