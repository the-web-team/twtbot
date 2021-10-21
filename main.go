package main

import (
	"errors"
	"flag"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"twtbot/discord"
	"twtbot/handlers/getstockprice"
	"twtbot/handlers/getuserpoints"
	"twtbot/handlers/givepoints"
	"twtbot/handlers/incrementkarma"
	"twtbot/handlers/rearrangerhandler"
	"twtbot/interfaces"
	"twtbot/services/bosstimers"
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
	pointsManager := &points.Service{}
	client.AttachService(pointsManager)

	bossTimersService := &bosstimers.Service{}
	client.AttachService(bossTimersService)

	// Handlers
	client.AttachHandler(func() interfaces.MessageHandlerInterface {
		return &givepoints.Handler{
			PointsManager: pointsManager,
		}
	})
	client.AttachHandler(func() interfaces.MessageHandlerInterface {
		return &getuserpoints.Handler{
			PointsManager: pointsManager,
		}
	})
	client.AttachHandler(func() interfaces.MessageHandlerInterface {
		return &rearrangerhandler.Handler{}
	})
	client.AttachHandler(func() interfaces.MessageHandlerInterface {
		return &incrementkarma.Handler{}
	})
	client.AttachHandler(func() interfaces.MessageHandlerInterface {
		return &getstockprice.Handler{}
	})

	// Slash Commands
	client.AttachCommand(&discordgo.ApplicationCommand{
		Type:        discordgo.ChatApplicationCommand,
		Name:        "registerboss",
		Description: "Command to register a new boss config",
	}, func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		respondError := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				// Note: this isn't documented, but you can use that if you want to.
				// This flag just allows you to create messages visible only for the caller of the command
				// (user who triggered the command)
				//Flags:   1 << 6,
				Content: "Test!",
			},
		})
		if respondError != nil {
			log.Fatal(respondError)
		}
	})

	if runError := client.Run(); runError != nil {
		log.Fatal(runError)
	}
}
