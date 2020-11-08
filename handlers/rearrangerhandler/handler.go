package rearrangerhandler

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"twtbot/interfaces"
	"twtbot/services/rearrange"
)

type Handler struct {
	interfaces.MessageHandler
}

func New(session *discordgo.Session, message *discordgo.MessageCreate) *Handler {
	handler := &Handler{}
	handler.SetSession(session)
	handler.SetMessage(message)
	return handler
}

func (h *Handler) ShouldRun() bool {
	channel, getChannelError := h.Session.Channel(h.Message.ChannelID)
	if getChannelError != nil {
		log.Println(getChannelError)
		return false
	}

	for _, validCategory := range validParentCategories {
		if validCategory == channel.ParentID {
			return true
		}
	}
	return false
}

func (h *Handler) Run() error {
	r := rearrange.Service{Session: h.Session}
	return r.Rearrange(h.Message.GuildID, h.Message.ChannelID)
}

var validParentCategories = []string{
	"597918822296584203",
	"563929079620042774",
	"766764321152827392", // Test Area
}
