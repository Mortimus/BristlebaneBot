package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	everquest "github.com/Mortimus/goEverquest"
)

// var nextDump time.Time

type RaidPlugin struct {
	Plugin
	Hours     int
	Bosses    int
	NeedsDump bool
	LastBoss  string
	SlayMatch *regexp.Regexp
	Start     time.Time
	NextDump  time.Time
	Started   bool
}

func init() {
	plug := new(RaidPlugin)
	plug.Name = "Raid Dump Detector"
	plug.Author = "Mortimus"
	plug.Version = "1.0.0"
	plug.Output = RAIDOUT
	Handlers = append(Handlers, plug)
	plug.NeedsDump = true
	plug.LastBoss = "Unknown"

	plug.SlayMatch, _ = regexp.Compile(configuration.Everquest.RegexSlay)
}

// Handle for ParsePlugin sends a message if a parse was pasted to the parse channel
func (p *RaidPlugin) Handle(msg *everquest.EqLog, out io.Writer) {
	if p.Started && !p.NeedsDump && getTime().Round(5*time.Minute) == p.NextDump.Round(5*time.Minute) {
		fmt.Fprintf(out, "Time for another hourly raid dump!\n")
		p.NeedsDump = true
	}
	if p.Started && msg.Channel == "system" && strings.Contains(msg.Msg, "has been slain by ") { // A spectre has been slain by Mortimus!
		// fmt.Fprintf(out, "TEST: %s\n", msg.Msg)
		Boss := p.SlayMatch.FindStringSubmatch(msg.Msg)[1]
		Slayer := p.SlayMatch.FindStringSubmatch(msg.Msg)[2]
		for _, dkpgiver := range configuration.Everquest.DKPGiver {
			if !strings.EqualFold(Boss, p.LastBoss) && strings.EqualFold(Boss, dkpgiver) {
				fmt.Fprintf(out, "%s was slain by %s awarding the raid DKP\n", Boss, Slayer)
				p.LastBoss = Boss
			}
		}
	}
	if msg.Channel == "system" && strings.Contains(msg.Msg, "Outputfile") && strings.Contains(msg.Msg, "RaidRoster") {
		// Upload the Raid Dump
		outputName := msg.Msg[21:] // Filename Outputfile sent data to
		file, err := os.Open(configuration.Everquest.BaseFolder + "/" + outputName)
		stamp := time.Now().Format("20060102")
		var fileName string
		if !p.NeedsDump { // Boss Kill
			formattedBoss := strings.Replace(p.LastBoss, " ", "_", -1)  // Remove Spaces
			formattedBoss = strings.Replace(formattedBoss, "`", "", -1) // Remove `
			formattedBoss = strings.Replace(formattedBoss, "'", "", -1) // Remove '
			fileName = fmt.Sprintf("%s_%s_%d.txt", stamp, formattedBoss, p.Bosses)
			p.LastBoss = "Unknown"
			p.Bosses++
		}
		if p.NeedsDump && p.Hours == 0 {
			fileName = stamp + "_raid_start.txt"
			p.NeedsDump = false
			p.Hours++
			p.Start = getTime().Round(1 * time.Hour)
			p.NextDump = time.Now().Add(1 * time.Hour)
			p.Started = true
		}
		if p.NeedsDump && p.Hours > 0 {
			fileName = fmt.Sprintf("%s_hour_%d.txt", stamp, p.Hours)
			p.NeedsDump = false
			p.Hours++
			p.NextDump = time.Now().Add(1 * time.Hour)
		}
		if p.Output != TESTOUT && err != nil {
			fmt.Fprintf(out, "Error finding Raid Dump: %s\n", outputName)
			return
		}
		if p.Output == RAIDOUT { // Send to discord as an upload
			discord.ChannelFileSend(configuration.Discord.RaidDumpChannelID, fileName, file)
			// uploadRaidDump(outputName)
		} else {
			fmt.Fprintf(out, "Uploading Raid Dump: %s\n", fileName)
		}
	}
}

func (p *RaidPlugin) Info(out io.Writer) {
	fmt.Fprintf(out, "---------------\n")
	fmt.Fprintf(out, "Name: %s\n", p.Name)
	fmt.Fprintf(out, "Author: %s\n", p.Author)
	fmt.Fprintf(out, "Version: %s\n", p.Version)
	fmt.Fprintf(out, "---------------\n")
}

func (p *RaidPlugin) OutputChannel() int {
	return p.Output
}
