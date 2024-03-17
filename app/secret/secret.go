package secret

import (
	"errors"
	"os"
)

type SecretService interface {
	Get(key string) (string, error)
}

type EnvSecretService struct{}

func NewEnvSecretService() *EnvSecretService {
	return &EnvSecretService{}
}

func (EnvSecretService) Get(key string) (string, error) {
	if value, ok := os.LookupEnv(key); ok {
		return value, nil
	}

	return "", errors.New(key + " not found")
}
