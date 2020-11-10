package incrementkarma

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"math/rand"
	"regexp"
	"strings"
	"sync"
	"twtbot/interfaces"
	"twtbot/lib"
	"twtbot/services/karma"
)

type Handler struct {
	interfaces.MessageHandler
	matches     []string
	userMatches []string
	roleMatches []string
}

func (h *Handler) ShouldRun() bool {
	h.userMatches = h.getMentionedUsers()
	h.roleMatches = h.getMentionedRoles()
	h.matches = regexp.MustCompile(`<@&(\d+)> ((--)|(\+\+))`).FindAllString(h.Message.Content, -1)
	return len(h.matches) > 0
}

func (h *Handler) Run() error {
	if typingError := h.Session.ChannelTyping(h.Message.ChannelID); typingError != nil {
		return typingError
	}

	karmaService := new(karma.Service)
	triedSelf := false
	deltas := make(map[string]int32)

	if len(h.userMatches) > 0 {
		var userDeltas map[string]int32
		userDeltas, triedSelf = h.getUserKarmaDeltas(h.userMatches)
		for userId, delta := range userDeltas {
			deltas[userId] += delta
		}
	}

	if len(h.roleMatches) > 0 {
		userRoleDeltas, userRoleDeltaError := h.getRoleKarmaDeltas(h.roleMatches)
		if userRoleDeltaError != nil {
			return userRoleDeltaError
		}
		for userId, delta := range userRoleDeltas {
			deltas[userId] += delta
		}
	}

	if len(deltas) > 0 {
		if updateError := karmaService.UpdateUsersKarma(deltas); updateError != nil {
			return updateError
		}

		userIdSet := &lib.Set{}
		for userId, _ := range deltas {
			userIdSet.Add(userId)
		}
		var userIds []string
		for userId := range userIdSet.GetItems() {
			userIds = append(userIds, userId.(string))
		}

		karmaRecords, getError := karmaService.GetUsersKarma(userIds)
		if getError != nil {
			return getError
		}

		var replies []string
		for userId, newKarma := range karmaRecords {
			karmaDelta := deltas[userId]
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

		if triedSelf {
			message := fmt.Sprintf("%s... You cannot give or take karma from yourself!", h.Message.Author.Mention())
			if _, sendError := h.Session.ChannelMessageSend(h.Message.ChannelID, message); sendError != nil {
				return sendError
			}
		}
	}
	return nil
}

func (h *Handler) getMentionedUsers() []string {
	return regexp.MustCompile(`<@!?(\d+)> ((--)|(\+\+))`).FindAllString(h.Message.Content, -1)
}

func (h *Handler) getMentionedRoles() []string {
	return regexp.MustCompile(`<@&(\d+)> ((--)|(\+\+))`).FindAllString(h.Message.Content, -1)
}

func (h *Handler) getUserKarmaDeltas(matches []string) (deltas map[string]int32, triedSelf bool) {
	deltas = make(map[string]int32)

	// User Karmas
	for _, match := range matches {
		userId := regexp.MustCompile(`\d+`).FindString(match)
		if userId == h.Message.Author.ID {
			triedSelf = true
			continue
		}
		if _, ok := deltas[userId]; !ok {
			deltas[userId] = 0
		}
		if strings.HasSuffix(match, "++") {
			deltas[userId]++
		} else if strings.HasSuffix(match, "--") {
			deltas[userId]--
		}
	}

	return deltas, triedSelf
}

func (h *Handler) getRoleKarmaDeltas(matches []string) (deltas map[string]int32, err error) {
	guildMembers, guildMembersError := h.getGuildMembers("")
	if guildMembersError != nil {
		return nil, guildMembersError
	}

	// Get Mentioned Role ID and their respective members
	incrementRoles := make(map[string]int32)
	decrementRoles := make(map[string]int32)
	for _, match := range matches {
		roleId := regexp.MustCompile(`\d+`).FindString(match)
		if strings.HasSuffix(match, "++") {
			if _, ok := incrementRoles[roleId]; !ok {
				incrementRoles[roleId] = 1
			} else {
				incrementRoles[roleId]++
			}
		} else if strings.HasSuffix(match, "--") {
			if _, ok := decrementRoles[roleId]; !ok {
				decrementRoles[roleId] = 1
			} else {
				decrementRoles[roleId]++
			}
		}
	}

	var wg sync.WaitGroup
	deltas = make(map[string]int32)
	updateUserDelta := func(member *discordgo.Member) {
		wg.Add(1)
		if member.User.ID != h.Message.Author.ID {
			var delta int32
			for _, roleId := range member.Roles {
				if _, ok := incrementRoles[roleId]; ok {
					delta++
				}

				if _, ok := decrementRoles[roleId]; ok {
					delta--
				}
			}
			deltas[member.User.ID] = delta
		}
		wg.Done()
	}

	for _, guildMember := range guildMembers {
		go updateUserDelta(guildMember)
	}

	wg.Wait()

	return deltas, nil
}

func (h *Handler) getGuildMembers(after string) ([]*discordgo.Member, error) {
	members, guildMembersError := h.Session.GuildMembers(h.Message.GuildID, after, 1000)
	if guildMembersError != nil {
		return nil, guildMembersError
	}
	if len(members) == 1000 {
		additionalMembers, additionalError := h.getGuildMembers(members[len(members)-1].User.ID)
		if additionalError != nil {
			return nil, additionalError
		}
		members = append(members, additionalMembers...)
	}
	return members, nil
}

var compliments = []string{
	"you are a gift to those around you!",
	"is awesome!",
	"is a smart cookie!",
	"has got swag!",
	"you are appreciated!",
	"is strong!",
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
