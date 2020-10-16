package rearrange

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

func HandleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	rearranger := &Service{session: s}
	if rearrangeError := rearranger.Rearrange(m.GuildID, m.ChannelID); rearrangeError != nil {
		log.Fatal(rearrangeError)
	}
}
