package mail

import (
	"github.com/jo-hoe/go-mail-service/internal/config"
)

// GetFieldOrDefault returns the user input if provided, otherwise retrieves the value from environment
func GetFieldOrDefault(userInput string, defaultEnvKey string) (string, error) {
	if userInput != "" {
		return userInput, nil
	}

	envService := config.NewEnvService()
	value, err := envService.Get(defaultEnvKey)
	if err != nil {
		return "", err
	}

	return value, nil
}
