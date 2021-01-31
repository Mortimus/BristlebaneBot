package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var discord *discordgo.Session

func reactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	l := LogInit("reactionAdd-main.go")
	defer l.End()
	if m.Emoji.Name == configuration.InvestigationStartEmoji && getPrivReactions(s, m.MessageID, configuration.InvestigationStartEmoji) == configuration.InvestigationStartMinReq && isArchive(m.MessageID) {
		l.InfoF("Investigation message: %s", m.MessageID)
		uploadArchive(m.MessageID)
	}
}

func getPrivReactions(s *discordgo.Session, messageID string, emoji string) int {
	l := LogInit("getReactions-main.go")
	defer l.End()
	var pReactions int
	users, err := discord.MessageReactions(configuration.LootChannelID, messageID, configuration.InvestigationStartEmoji, 100, "", "")
	if err != nil {
		l.ErrorF("Error getting message reactions: %s", err.Error())
		return -1
	}
	for _, user := range users {
		if isPriviledged(s, user.ID) {
			l.InfoF("User: %s signed off on an investigation for %s", user.Username, messageID)
			pReactions++
		}
	}
	l.InfoF("%d priveledged reactions for message id: %s", pReactions, messageID)
	return pReactions
}

func isPriviledged(s *discordgo.Session, userID string) bool {
	// TODO: Fix this
	// return true
	l := LogInit("isPriviledged-main.go")
	defer l.End()
	guildID := configuration.DiscordGuildID
	l.InfoF("UserID: %s SessionUser: %s", userID, s.State.User.ID)
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
