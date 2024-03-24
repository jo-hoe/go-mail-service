package config

import (
	"bufio"
	"io"
	"strings"
)

type SecretService interface {
	Get() (string, error)
}

type SecretFileService struct {}

func NewSecretFileService() *SecretFileService {
	return &SecretFileService{}
}

func (s *SecretFileService) Get(handle io.Reader) (secret string, err error) {
	var builder strings.Builder
	scanner := bufio.NewScanner(handle)
	
	for scanner.Scan() {
		line := scanner.Text() 
        builder.WriteString(line + "\n")
	}

	if err := scanner.Err(); err != nil {
        return "", err
    }

	secret = builder.String()
	if len(secret) != 0 {
		secret = strings.TrimSuffix(secret, "\n")
	}

	return secret, nil
}
