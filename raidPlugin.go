package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
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
		if !strings.EqualFold(Boss, p.LastBoss) {
			lowerBoss := strings.ToLower(Boss)
			if _, ok := bosses[lowerBoss]; ok {
				if bosses[lowerBoss].IsFTK {
					fmt.Fprintf(out, "%s was slain by %s awarding the raid %d+%d=%d DKP due to FTK\n", Boss, Slayer, bosses[lowerBoss].DKP, bosses[lowerBoss].FTK, bosses[lowerBoss].DKP+bosses[lowerBoss].FTK)
				} else {
					fmt.Fprintf(out, "%s was slain by %s awarding the raid %d DKP\n", Boss, Slayer, bosses[lowerBoss].DKP)
				}

				p.LastBoss = Boss
			}
		}
	}
	if msg.Channel == "system" && strings.Contains(msg.Msg, "Outputfile") && strings.Contains(msg.Msg, "RaidRoster") {
		// Upload the Raid Dump
		outputName := msg.Msg[21:] // Filename Outputfile sent data to
		file, err := os.Open(configuration.Everquest.BaseFolder + "/" + outputName)
		stamp := msg.T.Format("20060102")
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
			p.NextDump = msg.T.Add(1 * time.Hour)
			p.Started = true
		}
		if p.NeedsDump && p.Hours > 0 {
			fileName = fmt.Sprintf("%s_hour_%d.txt", stamp, p.Hours)
			p.NeedsDump = false
			p.Hours++
			p.NextDump = msg.T.Add(1 * time.Hour)
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

type BossDKP struct {
	Zone  string
	Note  string
	Boss  string
	DKP   int
	FTK   int
	IsFTK bool
}

var bosses map[string]*BossDKP

func seedBosses() {
	bosses = make(map[string]*BossDKP)
	spreadsheetID := configuration.Sheets.RawSheetURL
	readRange := configuration.Sheets.BossesSheetName
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		Err.Printf("Unable to retrieve data from sheet: %v", err)
		DiscordF(configuration.Discord.InvestigationChannelID, "Unable to read data from the Bosses sheet, cannot determine kills! - %s\n", err)
		// return errors.New("Unable to retrieve data from sheet: " + err.Error())
	}

	if len(resp.Values) == 0 {
		Err.Printf("Cannot read bosses sheet: %v", resp)
	} else {
		for i, row := range resp.Values {
			if i == 1 {
				continue // skip the header
			}
			boss := fmt.Sprintf("%s", row[configuration.Sheets.BossSheetBossCol])
			i := strings.Index(boss, ":")
			if i > -1 {
				boss = boss[i+1:]
			}
			boss = strings.TrimSpace(boss)
			boss = strings.ToLower(boss)
			if boss != "" {
				var newBoss BossDKP
				newBoss.Boss = boss

				zone := fmt.Sprintf("%s", row[configuration.Sheets.BossSheetZoneCol])
				zone = strings.TrimSpace(zone)
				newBoss.Zone = zone

				note := fmt.Sprintf("%s", row[configuration.Sheets.BossSheetNoteCol])
				note = strings.TrimSpace(note)
				newBoss.Note = note

				dkpString := fmt.Sprintf("%s", row[configuration.Sheets.BossSheetDKPCol])
				dkpPoints, err := strconv.Atoi(dkpString)
				if err != nil {
					Err.Printf("Error converting dkp points to float at row %d: %s", i+1, err.Error())
					// continue
					dkpPoints = 0
				}
				newBoss.DKP = dkpPoints

				ftkString := fmt.Sprintf("%s", row[configuration.Sheets.BossSheetFTKCol])
				ftkPoints, err := strconv.Atoi(ftkString)
				if err != nil {
					Err.Printf("Error converting ftk points to float at row %d: %s", i+1, err.Error())
					// continue
					dkpPoints = 0
				}
				newBoss.FTK = ftkPoints

				isFTK := true
				if len(row) > configuration.Sheets.BossSheetisFTKCol {
					isFTKString := fmt.Sprintf("%s", row[configuration.Sheets.BossSheetisFTKCol])
					isFTKString = strings.TrimSpace(isFTKString)
					if isFTKString == "Yes" {
						isFTK = false
					}
					newBoss.IsFTK = isFTK
				}

				bosses[boss] = &newBoss
			}
		}
	}
}

func printBosses() {
	for _, boss := range bosses {
		fmt.Printf("%#+v\n", boss)
	}
}
