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
			// exportGuild(guild)
			if _, ok := Roster[getPlayerName(configuration.Everquest.LogPath)]; ok {
				currentZone = Roster[getPlayerName(configuration.Everquest.LogPath)].Zone
				// fmt.Printf("Changing zone to %s\n", currentZone)
			}
		}
	}
}

// exportGuild takes a guild dump, replaces alts with mains, removes offline, and dedupes
// func exportGuild(guild *everquest.Guild) {
// 	// var uniqueMembers map[string]*DKPHolder
// 	uniqueMembers := make(map[string][]string)
// 	for _, member := range guild.Members {
// 		timeDiff := time.Since(member.LastOnline)
// 		if timeDiff.Hours() > 24 {
// 			// loc, _ := time.LoadLocation("America/Chicago")
// 			fmt.Printf("Removing %s from guild roster not online %s\n", member.Name, timeDiff.String())
// 			continue
// 		}
// 		main := member.Name
// 		if member.Alt { // If alt change to the main
// 			main := getMain(&member)
// 			if _, ok := Roster[main]; !ok { // Verify the member is in the map
// 				continue
// 			}
// 		}
// 		level := fmt.Sprintf("%d", member.Level)
// 		uniqueMembers[main] = []string{"0", member.Name, level, member.Class, "", "", "", "", "", ""}
// 	}
// 	// Make fake raid dump
// 	file, err := ioutil.TempFile("./", "raiddump")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	// defer os.Remove(file.Name())

// 	// fmt.Println(file.Name()) // For example "dir/prefix054003078"
// 	write := csv.NewWriter(file)
// 	write.Comma = '\t'
// 	fmt.Printf("Exporting guild roster to %s len: %d\n", file.Name(), len(uniqueMembers))
// 	for _, member := range uniqueMembers {
// 		write.Write(member)
// 		// fmt.Printf("Writing: %s\n", member)
// 		write.Flush()
// 	}
// 	// write.Flush()
// 	time.Sleep(2 * time.Second)
// 	_, err = discord.ChannelFileSend(configuration.Discord.RaidDumpChannelID, "Converted_GuildRaid_"+TimeStamp()+".txt", file)
// 	if err != nil {
// 		fmt.Printf("Error sending guild dump: %s\n", err.Error())
// 	}
// 	os.Remove(file.Name())
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
