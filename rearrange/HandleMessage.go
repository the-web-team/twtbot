package rearrange

import (
	"github.com/bwmarrin/discordgo"
	"log"
	"sort"
)

func HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) error {
	channel, getChannelError := s.Channel(m.ChannelID)
	if getChannelError != nil {
		return getChannelError
	}

	isValidCategory := false
	for _, validCategory := range validCategories {
		if validCategory == channel.ParentID {
			isValidCategory = true
			break
		}
	}
	if isValidCategory {
		guild, getGuildError := s.Guild(m.GuildID)
		if getGuildError != nil {
			return getGuildError
		}

		var channelsInSameCategory []*discordgo.Channel
		for _, guildChannel := range guild.Channels {
			if guildChannel.ParentID == channel.ParentID && channel.ID != guildChannel.ID {
				channelsInSameCategory = append(channelsInSameCategory, guildChannel)
			}
		}

		sort.Slice(channelsInSameCategory, func(i, j int) bool {
			return channelsInSameCategory[i].Position < channelsInSameCategory[j].Position
		})
		channels := append([]*discordgo.Channel{channel}, channelsInSameCategory...)
		for i, ch := range channels {
			ch.Position = i
		}
		if reorderError := s.GuildChannelsReorder(m.GuildID, channels); reorderError != nil {
			log.Fatal(reorderError)
		}
	}

	return nil
}

var validCategories = []string{
	"597918822296584203",
	"563929079620042774",
}
