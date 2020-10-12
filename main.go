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
	defer discord.Close()

	fmt.Println("Discord Bot is running...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(m.Content, Prefix) {
		// None yet
	} else {
		go karma.HandleMessageCreate(s, m)
	}
}
