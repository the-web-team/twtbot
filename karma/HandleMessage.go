package karma

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"regexp"
	"strings"
)

func HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) error {
	karmaService := &Service{}

	matches := regexp.MustCompile(`<@!?(\d+)> ((--)|(\+\+))`).FindAllString(m.Content, -1)
	triedSelf := false

	if len(matches) > 0 {
		var userIds []string
		var updates []Operation
		for _, match := range matches {
			userId := regexp.MustCompile(`\d+`).FindString(match)
			if userId == m.Author.ID {
				// Cannot give/remove karma to yourself
				triedSelf = true
				continue
			}
			userIds = append(userIds, userId)
			if strings.HasSuffix(match, "++") {
				updates = append(updates, Operation{
					UserId:     userId,
					KarmaDelta: 1,
				})
			} else if strings.HasSuffix(match, "--") {
				updates = append(updates, Operation{
					UserId:     userId,
					KarmaDelta: -1,
				})
			}
		}

		if len(updates) > 0 {
			if updateError := karmaService.updateUsersKarma(updates); updateError != nil {
				return updateError
			}

			karmaRecords, getError := karmaService.getUsersKarma(userIds)
			if getError != nil {
				return getError
			}

			var replies []string
			for _, update := range updates {
				user, userError := s.User(update.UserId)
				if userError != nil {
					return userError
				}

				newKarma := karmaRecords[update.UserId]
				if update.KarmaDelta == 1 {
					compliment := compliments[rand.Intn(len(compliments))]
					replies = append(replies, fmt.Sprintf("%s %s (Karma: %d)", user.Mention(), compliment, newKarma))
				} else if update.KarmaDelta == -1 {
					negativeComment := negativeComments[rand.Intn(len(negativeComments))]
					replies = append(replies, fmt.Sprintf("%s %s (Karma: %d)", user.Mention(), negativeComment, newKarma))
				}
			}

			if len(replies) > 0 {
				_, sendError := s.ChannelMessageSend(m.ChannelID, strings.Join(replies, "\n"))
				if sendError != nil {
					return sendError
				}
			}
		}

		if triedSelf {
			if typingError := s.ChannelTyping(m.ChannelID); typingError != nil {
				return typingError
			}
			if _, sendError := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s... You cannot give or take karma from yourself!", m.Author.Mention())); sendError != nil {
				return sendError
			}
		}
	}

	return nil
}
