package incrementgroupkarma

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
	matches []string
}

func (h *Handler) ShouldRun() bool {
	h.matches = regexp.MustCompile(`<@&(\d+)> ((--)|(\+\+))`).FindAllString(h.Message.Content, -1)
	return len(h.matches) > 0
}

func (h *Handler) Run() error {
	karmaService := new(karma.Service)

	guildMembers, guildMembersError := h.getGuildMembers("")
	if guildMembersError != nil {
		return guildMembersError
	}

	roleIdSet := &lib.StringSet{}
	for _, match := range h.matches {
		roleIdSet.Add(regexp.MustCompile(`\d+`).FindString(match))
	}
	roleIds := roleIdSet.GetItems()

	var wg sync.WaitGroup
	var lock sync.Mutex
	userUpdates := make(map[string]int32)
	rewardUser := func(member *discordgo.Member, num int32) {
		for _, roleId := range member.Roles {
			if _, ok := roleIds[roleId]; ok && member.User.ID != h.Message.Author.ID {
				lock.Lock()
				if _, ok := userUpdates[member.User.ID]; ok {
					userUpdates[member.User.ID] += num
				} else {
					userUpdates[member.User.ID] = num
				}
				lock.Unlock()
			}
		}
		wg.Done()
	}

	for _, match := range h.matches {
		if strings.HasSuffix(match, "++") {
			for _, guildMember := range guildMembers {
				wg.Add(1)
				go rewardUser(guildMember, 1)
			}
		} else if strings.HasSuffix(match, "--") {
			for _, guildMember := range guildMembers {
				wg.Add(1)
				go rewardUser(guildMember, -1)
			}
		}
	}

	wg.Wait()

	if len(userUpdates) > 0 {
		if updateError := karmaService.UpdateUsersKarma(userUpdates); updateError != nil {
			return updateError
		}

		userIdSet := &lib.Set{}
		for userId, _ := range userUpdates {
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
	return nil
}

func (h *Handler) getGuildMembers(after string) ([]*discordgo.Member, error) {
	members, guildMembersError := h.Session.GuildMembers(h.Message.GuildID, "", 1000)
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
