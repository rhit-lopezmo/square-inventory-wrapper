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

	dotenvErr := godotenv.Load()
	if dotenvErr != nil {
		log.Fatalln("ERROR: Could not load env file.")
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
