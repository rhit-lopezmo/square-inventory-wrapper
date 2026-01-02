package config

import (
	"aoa-inventory/utils"
	"github.com/joho/godotenv"
	"os"
	"strings"
)

var (
	Port              string
	AllowedOrigins    []string
	SquareAccessToken string
	SquareEnv         string
	SquareLocationID  string
)

var log = utils.NewLogger("CONFIG")

func Load() {
	log.Println("Loading config...")

	if err := godotenv.Load(); err == nil {
		log.Println(".env loaded")
	} else {
		log.Println("No local .env file, using real environment variables")
	}

	Port = os.Getenv("PORT")
	if Port == "" {
		log.Fatalln("ERROR: Could not find 'PORT' in env file.")
	}

	AllowedOrigins = strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
	if AllowedOrigins[0] == "" {
		log.Fatalln("ERROR: Could not find 'ALLOWED_ORIGINS' in env file.")
	}

	SquareAccessToken = os.Getenv("SQUARE_ACCESS_TOKEN")
	if SquareAccessToken == "" {
		log.Fatalln("ERROR: Could not find 'SQUARE_ACCESS_TOKEN' in env file.")
	}

	SquareEnv = os.Getenv("SQUARE_ENV")
	if SquareEnv == "" {
		log.Fatalln("ERROR: Could not find 'SQUARE_ENV' in env file.")
	}

	SquareLocationID = os.Getenv("SQUARE_LOCATION_ID")
	if SquareLocationID == "" {
		log.Fatalln("ERROR: Could not find 'SQUARE_LOCATION_ID' in env file.")
	}
}
