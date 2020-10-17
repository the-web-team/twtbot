package interfaces

import "github.com/bwmarrin/discordgo"

type HandlerInterface interface {
	ShouldRun() bool
	Run() error
	SetSession(*discordgo.Session)
}

type MessageHandlerInterface interface {
	HandlerInterface
	SetMessage(*discordgo.MessageCreate)
}
