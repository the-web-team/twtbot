package rearrange

import (
	"github.com/bwmarrin/discordgo"
)

func HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) error {
	if m.Author.ID == s.State.User.ID {
		return nil
	}

	if rearrangeError := rearrange(s, m.GuildID, m.ChannelID); rearrangeError != nil {
		return rearrangeError
	}

	return nil
}

var validCategories = []string{
	"597918822296584203",
	"563929079620042774",
}
