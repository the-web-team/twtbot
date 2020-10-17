package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
	"syscall"
)

type DiscordClient struct {
	session      *discordgo.Session
	handlers     []interface{}
	services     []func() error
	errorChannel chan error
}

func NewDiscordClient(authToken string) (*DiscordClient, error) {
	discordClient := new(DiscordClient)
	discord, sessionError := discordgo.New("Bot " + authToken)
	if sessionError != nil {
		return nil, sessionError
	}
	discordClient.session = discord
	return discordClient, nil
}

func (d *DiscordClient) Run() error {
	d.session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

	if openError := d.session.Open(); openError != nil {
		return openError
	}

	// Set status
	if statusError := d.session.UpdateListeningStatus("you via your Google Home"); statusError != nil {
		return statusError
	}

	go d.launchServices()

	fmt.Println("Discord Bot is running...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	return d.session.Close()
}

func (d *DiscordClient) AttachHandlers(handlers []interface{}) {
	for _, handler := range handlers {
		wrappedHandler := d.wrapHandler(handler)
		d.session.AddHandler(wrappedHandler)
	}
}

func (d *DiscordClient) AttachServices(services []func() error) {
	d.services = services
}

func (d *DiscordClient) launchServices() {
	for _, service := range d.services {
		go func(service func() error) {
			if serviceError := service(); serviceError != nil {
				d.errorChannel <- serviceError
			}
		}(service)
	}
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
