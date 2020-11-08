package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"twtbot/discord"
	"twtbot/handlers/getuserpointshandler"
	"twtbot/handlers/givepointshandler"
	"twtbot/handlers/karmahandler"
	"twtbot/handlers/rearrangerhandler"
	"twtbot/interfaces"
	"twtbot/services/points"
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
	client, clientError := discord.NewDiscordClient(AuthToken)
	if clientError != nil {
		log.Fatal(clientError)
	}

	// Run Services
	pointsManager := &points.Manager{}
	client.AttachService(pointsManager)

	// Handlers
	client.AttachHandler(func() interfaces.MessageHandlerInterface {
		return &givepointshandler.Handler{
			PointsManager: pointsManager,
		}
	})
	client.AttachHandler(func() interfaces.MessageHandlerInterface {
		return &getuserpointshandler.Handler{
			PointsManager: pointsManager,
		}
	})
	client.AttachHandler(func() interfaces.MessageHandlerInterface {
		return &rearrangerhandler.Handler{}
	})
	client.AttachHandler(func() interfaces.MessageHandlerInterface {
		return &karmahandler.Handler{}
	})

	if runError := client.Run(); runError != nil {
		log.Fatal(runError)
	}
}
