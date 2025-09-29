package config

import (
	"errors"
	"os"
)

type EnvService struct{}

func NewEnvService() *EnvService {
	return &EnvService{}
}

func (EnvService) Get(key string) (string, error) {
	if value, ok := os.LookupEnv(key); ok {
		return value, nil
	}

	return "", errors.New(key + " not found")
}
