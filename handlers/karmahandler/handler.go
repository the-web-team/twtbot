package karmahandler

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"twtbot/interfaces"
	"twtbot/services/karma"
)

type Handler struct {
	interfaces.MessageHandler
	matches []string
}

func (h *Handler) ShouldRun() bool {
	h.matches = regexp.MustCompile(`<@!?(\d+)> ((--)|(\+\+))`).FindAllString(h.Message.Content, -1)
	return len(h.matches) > 0
}

func (h *Handler) Run() error {
	karmaService := new(karma.Service)

	var userIds []string
	var updates []karma.Operation

	triedSelf := false
	for _, match := range h.matches {
		userId := regexp.MustCompile(`\d+`).FindString(match)
		if userId == h.Message.Author.ID {
			// Cannot give/remove karma to yourself
			triedSelf = true
			continue
		}
		userIds = append(userIds, userId)
		if strings.HasSuffix(match, "++") {
			updates = append(updates, karma.Operation{
				UserId:     userId,
				KarmaDelta: 1,
			})
		} else if strings.HasSuffix(match, "--") {
			updates = append(updates, karma.Operation{
				UserId:     userId,
				KarmaDelta: -1,
			})
		}
	}

	if len(updates) > 0 {
		if updateError := karmaService.UpdateUsersKarma(updates); updateError != nil {
			return updateError
		}

		karmaRecords, getError := karmaService.GetUsersKarma(userIds)
		if getError != nil {
			return getError
		}

		var replies []string
		for _, update := range updates {
			user, userError := h.Session.User(update.UserId)
			if userError != nil {
				return userError
			}

			newKarma := karmaRecords[update.UserId]
			if update.KarmaDelta == 1 {
				compliment := compliments[rand.Intn(len(compliments))]
				message := fmt.Sprintf("%s %s (Karma: %d)", user.Mention(), compliment, newKarma)
				replies = append(replies, message)
			} else if update.KarmaDelta == -1 {
				negativeComment := negativeComments[rand.Intn(len(negativeComments))]
				message := fmt.Sprintf("%s %s (Karma: %d)", user.Mention(), negativeComment, newKarma)
				replies = append(replies, message)
			}
		}

		if len(replies) > 0 {
			if _, sendError := h.Session.ChannelMessageSend(h.Message.ChannelID, strings.Join(replies, "\n")); sendError != nil {
				return sendError
			}
		}
	}

	if triedSelf {
		if typingError := h.Session.ChannelTyping(h.Message.ChannelID); typingError != nil {
			return typingError
		}
		message := fmt.Sprintf("%s... You cannot give or take karma from yourself!", h.Message.Author.Mention())
		if _, sendError := h.Session.ChannelMessageSend(h.Message.ChannelID, message); sendError != nil {
			return sendError
		}
	}

	return nil
}

var compliments = []string{
	"you are a gift to those around you!",
	"is awesome!",
	"is a smart cookie!",
	"has got swag!",
	"you are appreciated!",
	"is string!",
	"is inspiring!",
	"on a scale from 1 to 10, you're an 11!",
	"is a ray of sunshine!",
	"is making a difference!",
	"brings out the best in us!",
	"thank you!",
}

var negativeComments = []string{
	"my middle finger salutes you!",
	"is disappointing!",
	"rip...",
	"nope...",
	"fail!",
	"gtfo!",
}
