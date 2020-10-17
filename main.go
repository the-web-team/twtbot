package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"twtbot/karma"
	"twtbot/points"
	"twtbot/rearrange"
)

const Prefix = "!b"

var AuthToken string

var messageHandlers = []interface{}{
	pointsManager.HandleMessageCreate,
	rearrange.HandleMessageCreate,
	karma.HandleMessageCreate,
}

var pointsManager *points.Manager

var services = []func() error{
	pointsManager.StartService,
}

func init() {
	AuthToken = os.Getenv("DISCORD_AUTH_TOKEN")

	if AuthToken == "" {
		flag.StringVar(&AuthToken, "t", "", "Bot Token")
		flag.Parse()
	}

	if AuthToken == "" {
		log.Fatal(errors.New("no discord auth token supplied"))
	}

	pointsManager = &points.Manager{}
}

func main() {
	client, clientError := NewDiscordClient(AuthToken)
	if clientError != nil {
		log.Fatal(clientError)
	}

	client.AttachHandlers(messageHandlers)
	client.AttachServices(services)

	if runError := client.Run(); runError != nil {
		log.Fatal(runError)
	}
}
