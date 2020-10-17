package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"twtbot/interfaces"
)

type DiscordClient struct {
	session      *discordgo.Session
	handlers     []interface{}
	services     []interfaces.BotService
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

	// Shutdown Services
	d.shutdownServicesHandler()

	return d.session.Close()
}

func (d *DiscordClient) AttachMessageCreateHandler(handler interfaces.MessageHandlerInterface) {
	d.session.AddHandler(interfaces.CreateMessageHandler(handler))
}

func (d *DiscordClient) AttachService(service interfaces.BotService) interfaces.BotService {
	d.services = append(d.services, service)
	return service
}

func (d *DiscordClient) launchServices() {
	for _, service := range d.services {
		go func(service interfaces.BotService) {
			if startError := service.StartService(); startError != nil {
				d.errorChannel <- startError
			}
		}(service)
	}
}

func (d *DiscordClient) shutdownServicesHandler() {
	var wg sync.WaitGroup

	for _, service := range d.services {
		wg.Add(1)
		go func(service interfaces.BotService) {
			service.StopService()
			wg.Done()
		}(service)
	}

	wg.Wait()
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
