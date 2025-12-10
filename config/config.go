package config

import (
	"aoa-inventory/utils"
	"github.com/joho/godotenv"
	"os"
	"strings"
)

var (
	Port           string
	AllowedOrigins []string
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
}
