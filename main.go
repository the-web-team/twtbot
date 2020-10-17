package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"twtbot/message_handlers"
	"twtbot/points"
)

var AuthToken string

func init() {
	AuthToken = os.Getenv("DISCORD_AUTH_TOKEN")

	if AuthToken == "" {
		flag.StringVar(&AuthToken, "t", "", "Bot Token")
		flag.Parse()
	}

	if AuthToken == "" {
		log.Fatal(errors.New("no discord auth token supplied"))
	}
}

func main() {
	client, clientError := NewDiscordClient(AuthToken)
	if clientError != nil {
		log.Fatal(clientError)
	}

	// Run Services
	pointsManager := client.AttachService(new(points.Manager)).(*points.Manager)

	// Handlers
	client.AttachMessageCreateHandler(&message_handlers.GivePointsHandler{
		PointsManager: pointsManager,
	})
	client.AttachMessageCreateHandler(&message_handlers.GetUserPointsHandler{
		PointsManager: pointsManager,
	})
	client.AttachMessageCreateHandler(&message_handlers.RearrangerHandler{})
	client.AttachMessageCreateHandler(&message_handlers.KarmaHandler{})

	if runError := client.Run(); runError != nil {
		log.Fatal(runError)
	}
}
