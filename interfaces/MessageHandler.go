package interfaces

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
)

type MessageHandler struct {
	MessageHandlerInterface
	Session *discordgo.Session
	Message *discordgo.MessageCreate
}

type ChannelName string
type ChannelID string
type ChannelDictionary map[ChannelName]ChannelID

func (h *MessageHandler) ShouldRun() bool {
	return true
}

func (h *MessageHandler) Run() error {
	return nil
}

func (h *MessageHandler) GetPrefixedCommand(prefix rune) string {
	runes := []rune(h.Message.Content)
	isCommand := false
	var command []rune

	for _, r := range runes {
		willBeCommand := !isCommand && r == prefix
		if r == ' ' {
			break
		}

		if isCommand {
			command = append(command, r)
		}

		if willBeCommand {
			isCommand = true
		}
	}

	return string(command)
}

func (h *MessageHandler) GetCommand() string {
	return h.GetPrefixedCommand('!')
}

func (h *MessageHandler) CommandHasPrefix(prefix rune) bool {
	runes := []rune(h.Message.Content)
	return runes[0] == prefix
}

func (h *MessageHandler) IsCommand(command string) bool {
	return command == h.GetCommand()
}

func (h *MessageHandler) ChannelMessageSend(channelId string, message string) (*discordgo.Message, error) {
	if os.Getenv("ENV") == "development" {
		message = message + "\n^=== DEV ===^"
	}
	return h.Session.ChannelMessageSend(channelId, message)
}

func (h *MessageHandler) Reply(message string) error {
	if _, sendError := h.ChannelMessageSend(h.Message.ChannelID, message); sendError != nil {
		return sendError
	}
	return nil
}

func (h *MessageHandler) ReplyWithMention(message string) error {
	mentionMessage := fmt.Sprintf("%s, %s", h.Message.Author.Mention(), message)
	if _, sendError := h.ChannelMessageSend(h.Message.ChannelID, mentionMessage); sendError != nil {
		return sendError
	}
	return nil
}

func (h *MessageHandler) IsChannel(channelId string) bool {
	return h.Message.ChannelID == channelId
}

func (h *MessageHandler) GetChannelByName(channelName string) (*discordgo.Channel, error) {
	guild, guildError := h.Session.Guild(h.Message.GuildID)
	if guildError != nil {
		return nil, guildError
	}
	for _, channel := range guild.Channels {
		if channel.Name == channelName {
			return channel, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("channel does not exist by name %s", channelName))
}

func (h *MessageHandler) SetSession(s *discordgo.Session) {
	h.Session = s
}

func (h *MessageHandler) SetMessage(m *discordgo.MessageCreate) {
	h.Message = m
}
