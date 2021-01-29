package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var discord *discordgo.Session

func reactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	l := LogInit("reactionAdd-main.go")
	defer l.End()
	// fmt.Printf("Reaction Added! Emoji: %#+v MessageID: %s\n", m.Emoji, m.MessageID)
	// fmt.Printf("isEmoji: %t isArchive: %t\n", isEmoji, isArchive(m.MessageID))
	if m.Emoji.Name == configuration.InvestigationStartEmoji && isPriviledged(s, m.UserID) && getPrivReactions(s, m.MessageID, configuration.InvestigationStartEmoji) == configuration.InvestigationStartMinReq+1 && isArchive(m.MessageID) {
		// fmt.Printf("Investigating!\n")
		// TODO: Make this check if they are officers
		uploadArchive(m.MessageID)
	}
}

func getPrivReactions(s *discordgo.Session, messageID string, emoji string) int {
	l := LogInit("getReactions-main.go")
	defer l.End()
	msg, err := discord.ChannelMessage(configuration.LootChannelID, messageID)
	if err != nil {
		l.ErrorF("Error getting message: %s", err.Error())
		return -1
	}
	var privEmoji int
	for _, react := range msg.Reactions {
		if react.Emoji.Name == emoji {
			return react.Count
		}
		// TODO: This returns nil sometimes?
		// if react.Emoji.Name == emoji && isPriviledged(s, react.Emoji.User.ID) {
		// 	privEmoji++
		// }
	}
	return privEmoji
	// l.ErrorF("Cannot find emoji")
	// return -1 // Emoji not found
}

func isPriviledged(s *discordgo.Session, userID string) bool {
	// TODO: Fix this
	return true
	l := LogInit("isPriviledged-main.go")
	defer l.End()
	guildID := configuration.DiscordGuildID
	l.InfoF("GuildID: %+v\nUserID: %+v", guildID, userID)
	member, err := s.State.Member(guildID, userID)
	if err != nil {
		// if member, err = s.GuildMember(guildID, userID); err != nil {
		// 	return false, err
		// }
		l.ErrorF("Error: %s", err.Error())
	}
	l.InfoF("Member: %+v", member)
	for _, roleID := range member.Roles {
		role, err := s.State.Role(guildID, roleID)
		if err != nil {
			l.ErrorF("Error: %s", err.Error())
			return false
		}
		for _, cRole := range configuration.DiscordPrivRoles {
			l.InfoF("Crole: %v vs role.Name: %v", cRole, role.Name)
			if cRole == role.Name {
				l.InfoF("Role found, authorizing: %s == %s", cRole, role.Name)
				return true
			}
		}
	}
	return false
}

// DiscordF provides a printf to a discord channel
func DiscordF(format string, v ...interface{}) {
	l := LogInit("DiscordF-commands.go")
	defer l.End()
	msg := fmt.Sprintf(format, v...)
	_, err := discord.ChannelMessageSend(configuration.InvestigationChannelID, msg)
	if err != nil {
		l.ErrorF("Failed to send to discord: %s", err.Error())
	}
}
