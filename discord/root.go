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

type CommandHandler = func(s *discordgo.Session, i *discordgo.InteractionCreate)

type Client struct {
	session         *discordgo.Session
	handlers        []interface{}
	commands        []*discordgo.ApplicationCommand
	commandHandlers map[string]CommandHandler
	services        []interfaces.BotService
	errorChannel    chan error
}

func NewDiscordClient(authToken string) (*Client, error) {
	discordClient := new(Client)
	discord, sessionError := discordgo.New("Bot " + authToken)
	if sessionError != nil {
		return nil, sessionError
	}
	discordClient.session = discord
	discordClient.commandHandlers = make(map[string]CommandHandler)
	return discordClient, nil
}

func (d *Client) GetSession() *discordgo.Session {
	return d.session
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
	go d.attachCommands()

	// Register Command Handlers
	d.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if handler, ok := d.commandHandlers[i.ApplicationCommandData().Name]; ok {
			handler(s, i)
		}
	})

	d.session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Println("Discord Bot is running...")
	})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-stop

	// Shutdown Services
	fmt.Println("Shutting down services...")
	d.shutdownServicesHandler()

	fmt.Println("Shutting down bot...")
	return d.session.Close()
}

func (d *Client) AttachCommand(command *discordgo.ApplicationCommand, handler CommandHandler) {
	d.commandHandlers[command.Name] = handler
	d.commands = append(d.commands, command)
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

func (d *Client) attachCommands() {
	fmt.Println(d.session.State.Guilds)
	for _, command := range d.commands {
		for _, guild := range d.session.State.Guilds {
			fmt.Printf("Adding %s slash command to %s guild\n", command.Name, guild.Name)
			userId := d.session.State.User.ID
			_, createCommandError := d.session.ApplicationCommandCreate(userId, guild.ID, command)
			if createCommandError != nil {
				log.Fatal(createCommandError)
			}
		}
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
