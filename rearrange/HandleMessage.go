package rearrange

import (
	"github.com/bwmarrin/discordgo"
)

func HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) error {
	if m.Author.ID == s.State.User.ID {
		return nil
	}

	rearranger := &Service{session: s}

	if rearrangeError := rearranger.Rearrange(m.GuildID, m.ChannelID); rearrangeError != nil {
		return rearrangeError
	}

	return nil
}
