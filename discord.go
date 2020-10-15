package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
	"syscall"
)

type DiscordClient struct {
	AuthToken string
	session   *discordgo.Session
	handlers  []interface{}
	services  []func()
}

func NewDiscordClient(authToken string) *DiscordClient {
	discordClient := new(DiscordClient)
	discordClient.AuthToken = authToken
	return discordClient
}

func (d *DiscordClient) Run() error {
	discord, sessionError := discordgo.New("Bot " + d.AuthToken)
	if sessionError != nil {
		return sessionError
	}

	// Add Handlers
	for _, handler := range d.handlers {
		discord.AddHandler(d.wrapHandler(handler))
	}

	discord.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

	if openError := discord.Open(); openError != nil {
		return openError
	}

	// Set status
	if statusError := discord.UpdateListeningStatus("you via your Google Home"); statusError != nil {
		return statusError
	}

	for _, service := range d.services {
		go service()
	}

	fmt.Println("Discord Bot is running...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	return discord.Close()
}

func (d *DiscordClient) AttachHandlers(handlers []interface{}) {
	d.handlers = handlers
}

func (d *DiscordClient) AttachServices(services []func()) {
	d.services = services
}

func (d *DiscordClient) wrapHandler(handler interface{}) interface{} {
	switch v := handler.(type) {
	case func(*discordgo.Session, *discordgo.MessageCreate):
		return func(s *discordgo.Session, m *discordgo.MessageCreate) {
			if m.Author.ID == s.State.User.ID {
				return
			}
			v(s, m)
		}
	}

	return handler
}
