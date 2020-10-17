package message_handlers

import (
	"log"
	"twtbot/interfaces"
	"twtbot/rearrange"
)

type RearrangerHandler struct {
	interfaces.MessageHandler
}

func (h *RearrangerHandler) ShouldRun() bool {
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

func (h *RearrangerHandler) Run() error {
	r := rearrange.Service{Session: h.Session}
	return r.Rearrange(h.Message.GuildID, h.Message.ChannelID)
}

var validParentCategories = []string{
	"597918822296584203",
	"563929079620042774",
	"766764321152827392", // Test Area
}
