package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var discord *discordgo.Session

func reactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	if m.Emoji.Name == configuration.Discord.InvestigationStartEmoji && getPrivReactions(s, m.MessageID, configuration.Discord.InvestigationStartEmoji) == configuration.Discord.InvestigationMinRequired && isArchive(m.MessageID) {
		Info.Printf("Investigation message: %s", m.MessageID)
		uploadArchive(m.MessageID)
	}
}

func getPrivReactions(s *discordgo.Session, messageID string, emoji string) int {
	var pReactions int
	users, err := discord.MessageReactions(configuration.Discord.LootChannelID, messageID, configuration.Discord.InvestigationStartEmoji, 100, "", "")
	if err != nil {
		Err.Printf("Error getting message reactions: %s", err.Error())
		return -1
	}
	for _, user := range users {
		if isPriviledged(s, user.ID) {
			Info.Printf("User: %s signed off on an investigation for %s", user.Username, messageID)
			pReactions++
		}
	}
	Info.Printf("%d priveledged reactions for message id: %s", pReactions, messageID)
	return pReactions
}

func isPriviledged(s *discordgo.Session, userID string) bool {
	// TODO: Fix this
	// return true
	guildID := configuration.Discord.GuildID
	Info.Printf("UserID: %s SessionUser: %s", userID, s.State.User.ID)
	Info.Printf("GuildID: %+v\nUserID: %+v", guildID, userID)
	member, err := s.State.Member(guildID, userID)
	if err != nil {
		// if member, err = s.GuildMember(guildID, userID); err != nil {
		// 	return false, err
		// }
		Err.Printf("Error: %s", err.Error())
	}
	Info.Printf("Member: %+v", member)
	for _, roleID := range member.Roles {
		role, err := s.State.Role(guildID, roleID)
		if err != nil {
			Err.Printf("Error: %s", err.Error())
			return false
		}
		for _, cRole := range configuration.Discord.PrivRoles {
			Info.Printf("Crole: %v vs role.Name: %v", cRole, role.Name)
			if cRole == role.Name {
				Info.Printf("Role found, authorizing: %s == %s", cRole, role.Name)
				return true
			}
		}
	}
	return false
}

// DiscordF provides a printf to a discord channel
func DiscordF(channel string, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	_, err := discord.ChannelMessageSend(channel, msg)
	if err != nil {
		Err.Printf("Failed to send message to %s: %s", channel, err.Error())
	}
}
