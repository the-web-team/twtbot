package incrementuserkarma

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
	fmt.Println(h.Message.Content)
	h.matches = regexp.MustCompile(`<@!?(\d+)> ((--)|(\+\+))`).FindAllString(h.Message.Content, -1)
	return len(h.matches) > 0
}

func (h *Handler) Run() error {
	karmaService := new(karma.Service)

	var userIds []string
	userUpdates := make(map[string]int32)

	triedSelf := false
	for _, match := range h.matches {
		userId := regexp.MustCompile(`\d+`).FindString(match)
		if userId == h.Message.Author.ID {
			// Cannot give/remove karma to yourself
			triedSelf = true
			continue
		}
		userIds = append(userIds, userId)
		if _, ok := userUpdates[userId]; !ok {
			userUpdates[userId] = 0
		}
		if strings.HasSuffix(match, "++") {
			userUpdates[userId]++
		} else if strings.HasSuffix(match, "--") {
			userUpdates[userId]--
		}
	}

	if len(userUpdates) > 0 {
		if updateError := karmaService.UpdateUsersKarma(userUpdates); updateError != nil {
			return updateError
		}

		karmaRecords, getError := karmaService.GetUsersKarma(userIds)
		if getError != nil {
			return getError
		}

		var replies []string
		for userId, newKarma := range karmaRecords {
			karmaDelta := userUpdates[userId]
			user, userError := h.Session.User(userId)
			if userError != nil {
				return userError
			}

			if karmaDelta > 0 {
				compliment := compliments[rand.Intn(len(compliments))]
				message := fmt.Sprintf("%s %s (Karma: %d)", user.Mention(), compliment, newKarma)
				replies = append(replies, message)
			} else if karmaDelta < 0 {
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
