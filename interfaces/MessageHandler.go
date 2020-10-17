package interfaces

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
)

type MessageHandler struct {
	MessageHandlerInterface
	Session *discordgo.Session
	Message *discordgo.MessageCreate
}

func (h *MessageHandler) ShouldRun() bool {
	return true
}

func (h *MessageHandler) Run() error {
	return nil
}

func (h *MessageHandler) GetCommand() string {
	runes := []rune(h.Message.Content)
	isCommand := false
	var command []rune

	for _, r := range runes {
		willBeCommand := !isCommand && r == '!'
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

func (h *MessageHandler) IsCommand(command string) bool {
	return command == h.GetCommand()
}

func (h *MessageHandler) Reply(message string) error {
	if _, sendError := h.Session.ChannelMessageSend(h.Message.ChannelID, message); sendError != nil {
		return sendError
	}
	return nil
}

func (h *MessageHandler) ReplyWithMention(message string) error {
	mentionMessage := fmt.Sprintf("%s, %s", h.Message.Author.Mention(), message)
	if _, sendError := h.Session.ChannelMessageSend(h.Message.ChannelID, mentionMessage); sendError != nil {
		return sendError
	}
	return nil
}

func (h *MessageHandler) SetSession(s *discordgo.Session) {
	h.Session = s
}

func (h *MessageHandler) SetMessage(m *discordgo.MessageCreate) {
	h.Message = m
}

func CreateMessageHandler(h MessageHandlerInterface) func(s *discordgo.Session, m *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {
		h.SetSession(s)
		h.SetMessage(m)
		if m.Author.ID == s.State.User.ID || !h.ShouldRun() {
			return
		}
		if runError := h.Run(); runError != nil {
			log.Println(runError)
		}
	}
}
