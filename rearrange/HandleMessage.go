package rearrange

import "github.com/bwmarrin/discordgo"

func HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) error {
	channel, getChannelError := s.Channel(m.ChannelID)
	if getChannelError != nil {
		return getChannelError
	}

	for _, categoryId := range validCategories {
		if channel.ParentID == categoryId {
			_, editError := s.ChannelEditComplex(m.ChannelID, &discordgo.ChannelEdit{
				Position: 0,
			})
			if editError != nil {
				return editError
			}
			break
		}
	}

	return nil
}

var validCategories = []string{
	"597918822296584203",
	"563929079620042774",
}
