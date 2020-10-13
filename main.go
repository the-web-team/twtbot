package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"twtbot/karma"
	"twtbot/points"
	"twtbot/rearrange"
)

const Prefix = "!b"

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
	discord, sessionError := discordgo.New("Bot " + AuthToken)
	if sessionError != nil {
		fmt.Println("error creating discord session")
		log.Fatal(sessionError)
	}

	discord.AddHandler(messageCreate)

	discord.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

	openError := discord.Open()
	if openError != nil {
		fmt.Println("error opening discord connection")
		log.Fatal(openError)
	}
	defer func() {
		_ = discord.Close()
	}()

	if statusError := discord.UpdateListeningStatus("you via your Google Home"); statusError != nil {
		log.Fatal(statusError)
	}

	// Start Services
	points.StartService()

	fmt.Println("Discord Bot is running...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	handleError := func(err error) {
		errorText := ":exclamation::exclamation::exclamation::exclamation: Error ```%s```"
		_, sendError := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(errorText, err.Error()))
		if sendError != nil {
			log.Fatal(sendError)
		}
	}

	if strings.HasPrefix(m.Content, Prefix) {
		// None yet
	} else {
		// Points Manager
		go points.HandleMessageCreate(s, m)

		// Rearrange
		go func() {
			if rearrangeError := rearrange.HandleMessageCreate(s, m); rearrangeError != nil {
				handleError(rearrangeError)
			}
		}()

		// Karma
		go func() {
			if karmaError := karma.HandleMessageCreate(s, m); karmaError != nil {
				handleError(karmaError)
			}
		}()
	}
}
