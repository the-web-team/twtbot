package rearrange

import (
	"github.com/bwmarrin/discordgo"
	"sort"
)

type Service struct {
	Session *discordgo.Session
}

func (r *Service) Rearrange(guildId string, channelId string) error {
	channel, getChannelError := r.Session.Channel(channelId)
	if getChannelError != nil {
		return getChannelError
	}

	guildChannels, getChannelsError := r.Session.GuildChannels(guildId)
	if getChannelsError != nil {
		return getChannelsError
	}

	var channelsInSameCategory []*discordgo.Channel
	for _, guildChannel := range guildChannels {
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

	return r.Session.GuildChannelsReorder(guildId, channels)
}
