package rearrange

import (
	"github.com/bwmarrin/discordgo"
	"sort"
)

var validParentCategories = []string{
	"597918822296584203",
	"563929079620042774",
}

type Service struct {
	session *discordgo.Session
}

func (r *Service) Rearrange(guildId string, channelId string) error {
	channel, getChannelError := r.session.Channel(channelId)
	if getChannelError != nil {
		return getChannelError
	}

	if r.isValidChannel(channel) {
		guild, getGuildError := r.session.Guild(guildId)
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
		return r.session.GuildChannelsReorder(guildId, channels)
	}

	return nil
}

func (r *Service) isValidChannel(channel *discordgo.Channel) bool {
	for _, validCategory := range validParentCategories {
		if validCategory == channel.ParentID {
			return true
		}
	}
	return false
}