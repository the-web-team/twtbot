package discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"twtbot/interfaces"
)

type Client struct {
	session      *discordgo.Session
	handlers     []interface{}
	services     []interfaces.BotService
	errorChannel chan error
}

func NewDiscordClient(authToken string) (*Client, error) {
	discordClient := new(Client)
	discord, sessionError := discordgo.New("Bot " + authToken)
	if sessionError != nil {
		return nil, sessionError
	}
	discordClient.session = discord
	return discordClient, nil
}

func (d *Client) Run() error {
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

func (d *Client) AttachHandler(handler interface{}) {
	if messageHandler, ok := handler.(func() interfaces.MessageHandlerInterface); ok {
		wrappedHandler := func(s *discordgo.Session, m *discordgo.MessageCreate) {
			h := messageHandler()
			h.SetSession(s)
			h.SetMessage(m)
			if m.Author.ID == s.State.User.ID || !h.ShouldRun() {
				return
			}
			if runError := h.Run(); runError != nil {
				log.Println(runError)
			}
		}
		d.session.AddHandler(wrappedHandler)
	} else {
		log.Println("Invalid handler type")
	}
}

func (d *Client) AttachService(service interfaces.BotService) {
	d.services = append(d.services, service)
}

func (d *Client) launchServices() {
	for _, service := range d.services {
		go func(service interfaces.BotService) {
			if startError := service.StartService(); startError != nil {
				d.errorChannel <- startError
			}
		}(service)
	}
}

func (d *Client) shutdownServicesHandler() {
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

func (d *Client) wrapHandler(handler interface{}) interface{} {
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

type MessageHandlerFactory func(session *discordgo.Session, message *discordgo.Message) interface{}
