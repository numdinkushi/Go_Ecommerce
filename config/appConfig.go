package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	ServerPort               string
	DBHost                   string
	DBPort                   string
	DBUser                   string
	DBPassword               string
	DBName                   string
	JwtSecret                string
	TwilioAccountSid         string
	TwilioAuthToken          string
	TwilioPhoneNumber        string
	FlutterwaveClientID      string
	FlutterwaveSecretKey     string
	FlutterwaveEncryptionKey string
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

	dbHost := os.Getenv("DB_HOST")
	if len(dbHost) < 1 {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if len(dbPort) < 1 {
		dbPort = "5432"
	}

	dbUser := os.Getenv("DB_USER")
	if len(dbUser) < 1 {
		return AppConfig{}, errors.New("DB_USER is not set, env variable is not found")
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if len(dbPassword) < 1 {
		return AppConfig{}, errors.New("DB_PASSWORD is not set, env variable is not found")
	}

	dbName := os.Getenv("DB_NAME")
	if len(dbName) < 1 {
		return AppConfig{}, errors.New("DB_NAME is not set, env variable is not found")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 1 {
		return AppConfig{}, errors.New("JWT_SECRET is not set, env variable is not found")
	}

	twilioAccountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	if len(twilioAccountSid) < 1 {
		return AppConfig{}, errors.New("TWILIO_ACCOUNT_SID is not set, env variable is not found")
	}

	twilioAuthToken := os.Getenv("TWILIO_AUTH_TOKEN")
	if len(twilioAuthToken) < 1 {
		return AppConfig{}, errors.New("TWILIO_AUTH_TOKEN is not set, env variable is not found")
	}

	twilioPhoneNumber := os.Getenv("TWILIO_PHONE_NUMBER")
	if len(twilioPhoneNumber) < 1 {
		return AppConfig{}, errors.New("TWILIO_PHONE_NUMBER is not set, env variable is not found")
	}

	flutterwaveClientID := os.Getenv("FLUTTERWAVE_CLIENT_ID")
	flutterwaveSecretKey := os.Getenv("FLUTTERWAVE_SECRET_KEY")
	flutterwaveEncryptionKey := os.Getenv("FLUTTERWAVE_ENCRYPTION_KEY")
	// Note: FLUTTERWAVE_SECRET_KEY is required for bank verification features
	// Get your keys from: https://dashboard.flutterwave.com (Settings > API Keys)

	return AppConfig{
		ServerPort:               httpPort,
		DBHost:                   dbHost,
		DBPort:                   dbPort,
		DBUser:                   dbUser,
		DBPassword:               dbPassword,
		DBName:                   dbName,
		JwtSecret:                jwtSecret,
		TwilioAccountSid:         twilioAccountSid,
		TwilioAuthToken:          twilioAuthToken,
		TwilioPhoneNumber:        twilioPhoneNumber,
		FlutterwaveClientID:      flutterwaveClientID,
		FlutterwaveSecretKey:     flutterwaveSecretKey,
		FlutterwaveEncryptionKey: flutterwaveEncryptionKey,
	}, nil
}
