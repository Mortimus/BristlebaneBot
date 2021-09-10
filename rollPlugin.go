package main

import (
	"fmt"
	"io"
	"regexp"
	"strconv"

	everquest "github.com/Mortimus/goEverquest"
)

var needsRolled []string

type RollPlugin struct {
	Plugin
	RollMatch *regexp.Regexp
}

func init() {
	plug := new(RollPlugin)
	plug.Name = "Roll detection"
	plug.Author = "Mortimus"
	plug.Version = "1.0.0"
	plug.Output = BIDOUT
	plug.RollMatch, _ = regexp.Compile(configuration.Everquest.RegexRoll)
	Handlers = append(Handlers, plug)
}

// Handle for RollPlugin sends a message if a parse was pasted to the parse channel
func (p *RollPlugin) Handle(msg *everquest.EqLog, out io.Writer) {
	if msg.Channel == "system" {
		match := p.RollMatch.FindStringSubmatch(msg.Msg)
		if len(match) > 0 {
			player := match[1]
			if player == "You" {
				player = getPlayerName(configuration.Everquest.LogPath)
			}
			low, _ := strconv.Atoi(match[2])
			high, _ := strconv.Atoi(match[3])
			result, _ := strconv.Atoi(match[4])
			if low != 0 || high != 1000 {
				return
			}
			for _, rollers := range needsRolled {
				if rollers == player {
					fmt.Fprintf(out, "```ini\n[%s rolled a %d]\n```", player, result)
					removeRollerFromRoll(player)
					return
				}
			}
		}
	}
}

func removeRollerFromRoll(player string) {
	var PlayerPos int
	for pos, name := range needsRolled {
		if name == player {
			PlayerPos = pos
		}
	}
	needsRolled = append(needsRolled[:PlayerPos], needsRolled[PlayerPos+1:]...)
}

func (p *RollPlugin) Info(out io.Writer) {
	fmt.Fprintf(out, "---------------\n")
	fmt.Fprintf(out, "Name: %s\n", p.Name)
	fmt.Fprintf(out, "Author: %s\n", p.Author)
	fmt.Fprintf(out, "Version: %s\n", p.Version)
	fmt.Fprintf(out, "---------------\n")
}

func (p *RollPlugin) OutputChannel() int {
	return p.Output
}
