package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	ServerPort string
}

func SetupEnv() (config AppConfig, err error) {
	// Load .env file first to make APP_ENV available
	godotenv.Load()

	if os.Getenv("APP_ENV") == "dev" {
		// .env already loaded above, but ensure it's loaded
		godotenv.Load()
	} else {
		godotenv.Load(".env.production")
	}

	httpPort := os.Getenv("HTTP_PORT")
	if len(httpPort) < 1 {
		return AppConfig{}, errors.New("HTTP_PORT is not set, env variable is not found")
	} 

	return AppConfig{ServerPort: httpPort}, nil
}
