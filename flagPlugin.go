package main

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	everquest "github.com/Mortimus/goEverquest"
)

var flagPiece map[string]interface{}

type FlagPlugin struct {
	Plugin
	LootMatch *regexp.Regexp
}

func init() {
	plug := new(FlagPlugin)
	plug.Name = "Flag hailing"
	plug.Author = "Mortimus"
	plug.Version = "1.0.0"
	plug.Output = FLAGOUT
	Handlers = append(Handlers, plug)

	plug.LootMatch, _ = regexp.Compile(configuration.Everquest.RegexLoot)
	seedFlagPieces()
}

// Handle for ParsePlugin sends a message if a parse was pasted to the parse channel
func (p *FlagPlugin) Handle(msg *everquest.EqLog, out io.Writer) {
	if msg.Channel == "say" && strings.Contains(msg.Msg, "Hail, ") {
		for _, flaggiver := range configuration.Everquest.FlagGiver {
			if strings.Contains(msg.Msg, flaggiver) {
				fmt.Fprintf(out, "%s got the flag from %s\n", msg.Source, currentZone)
			}
		}
	}
	if msg.Channel == "system" {
		match := p.LootMatch.FindStringSubmatch(msg.Msg)
		if len(match) > 0 {
			player := match[1]
			if player == "You" {
				player = getPlayerName(configuration.Everquest.LogPath)
			}
			loot := match[2]
			if loot != "" && isFlagPiece(loot) {
				fmt.Fprintf(out, "%s got the %s flag from %s\n", player, loot, currentZone)
			}
		}
	}
}

func (p *FlagPlugin) Info(out io.Writer) {
	fmt.Fprintf(out, "---------------\n")
	fmt.Fprintf(out, "Name: %s\n", p.Name)
	fmt.Fprintf(out, "Author: %s\n", p.Author)
	fmt.Fprintf(out, "Version: %s\n", p.Version)
	fmt.Fprintf(out, "---------------\n")
}

func (p *FlagPlugin) OutputChannel() int {
	return p.Output
}

func isFlagPiece(item string) bool {
	if _, ok := flagPiece[item]; ok {
		return true
	}
	return false
}

func seedFlagPieces() {
	flagPiece = make(map[string]interface{})
	flagPiece["Artifact of Righteousness"] = nil
	flagPiece["Artifact of Glorification"] = nil
	flagPiece["Artifact of Transcendence"] = nil
	flagPiece["Sliver of the High Temple"] = nil
	flagPiece["Zun'Muram's Signet of Command"] = nil
}
