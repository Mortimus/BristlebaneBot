package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	everquest "github.com/Mortimus/goEverquest"
)

// var nextDump time.Time
type GuildPlugin Plugin

func init() {
	plug := new(GuildPlugin)
	plug.Name = "Guild Dump Detector"
	plug.Author = "Mortimus"
	plug.Version = "1.0.0"
	plug.Output = INVESTIGATEOUT
	Handlers = append(Handlers, plug)
}

// Handle for ParsePlugin sends a message if a parse was pasted to the parse channel
func (p *GuildPlugin) Handle(msg *everquest.EqLog, out io.Writer) {
	if msg.Channel == "system" && strings.Contains(msg.Msg, "Outputfile") && strings.Contains(msg.Msg, configuration.Everquest.GuildName) {
		outputName := msg.Msg[21:] // Filename Outputfile sent data to

		guild := new(everquest.Guild)
		err := guild.LoadFromPath(configuration.Everquest.BaseFolder+"/"+outputName, Err)
		if err != nil {
			fmt.Printf("Error loading roster dump: %s", err.Error())
		} else {
			fmt.Fprintf(out, "Updating Guild Roster: %s\n", outputName)
			apiUploadGuildRoster(configuration.Everquest.BaseFolder + "/" + outputName)
			guildFile, err := os.Open(configuration.Everquest.BaseFolder + "/" + outputName)
			if err != nil {
				fmt.Fprintf(out, "Error finding Guild Dump: %s\n", outputName)
			} else {
				discord.ChannelFileSend(configuration.Discord.RaidDumpChannelID, outputName, guildFile)
			}
			updateGuildRoster(guild) // Fix github issue?
			if _, ok := Roster[getPlayerName(configuration.Everquest.LogPath)]; ok {
				currentZone = Roster[getPlayerName(configuration.Everquest.LogPath)].Zone
				// fmt.Printf("Changing zone to %s\n", currentZone)
			}
		}
	}
}

// func exportGuild(guild *everquest.Guild) {

// }

func (p *GuildPlugin) Info(out io.Writer) {
	fmt.Fprintf(out, "---------------\n")
	fmt.Fprintf(out, "Name: %s\n", p.Name)
	fmt.Fprintf(out, "Author: %s\n", p.Author)
	fmt.Fprintf(out, "Version: %s\n", p.Version)
	fmt.Fprintf(out, "---------------\n")
}

func (p *GuildPlugin) OutputChannel() int {
	return p.Output
}
