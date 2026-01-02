package client

import (
	"aoa-inventory/utils"

	square "github.com/square/square-go-sdk"
	client "github.com/square/square-go-sdk/client"
	option "github.com/square/square-go-sdk/option"
)

var log = utils.NewLogger("SQUARE-CLIENT")

var SquareClient *client.Client

func Init(accessToken, env string) {
	if SquareClient != nil {
		log.Println("WARNING: Square client already initialized, skipping...")
		return
	}

	var envUrl string
	switch env {
	case "production":
		envUrl = square.Environments.Production
	case "sandbox":
		envUrl = square.Environments.Sandbox
	default:
		log.Fatalln("ERROR: Invalid Square environment, exiting...")
	}

	SquareClient = client.NewClient(
		option.WithToken(accessToken),
		option.WithBaseURL(envUrl),
	)

	log.Printf("Initialized Square client in %s environment", env)
}
