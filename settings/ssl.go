package settings

import (
	"errors"
	"fmt"
	"os"
)

func (s *Settings) CheckForCerts() error {
	if s.SSLCertificatePath == "" {
		return errors.New("path to SSL Certificate file (public key) not found in settings")
	}
	if s.SSLKeyPath == "" {
		return errors.New("path to SSL Key file (private key) not found in settings")
	}

	if _, err := os.Stat(s.SSLCertificatePath); os.IsNotExist(err) {
		return fmt.Errorf("SSL certificate not found at %v", s.SSLCertificatePath)
	}

	if _, err := os.Stat(s.SSLKeyPath); os.IsNotExist(err) {
		return fmt.Errorf("SSL key not found at %v", s.SSLKeyPath)
	}
	return nil
}
