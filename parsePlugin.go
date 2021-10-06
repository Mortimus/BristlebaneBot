package main

import (
	"fmt"
	"io"
	"strings"

	everquest "github.com/Mortimus/goEverquest"
)

type ParsePlugin Plugin

func init() {
	plug := new(ParsePlugin)
	plug.Name = "Parse posting"
	plug.Author = "Mortimus"
	plug.Version = "1.0.0"
	plug.Output = PARSEOUT
	Handlers = append(Handlers, plug)
}

// Handle for ParsePlugin sends a message if a parse was pasted to the parse channel
func (p *ParsePlugin) Handle(msg *everquest.EqLog, out io.Writer) {
	if strings.Contains(strings.ToLower(msg.Channel), "von_parses") && strings.Contains(msg.Msg, "s, ") {
		if msg.Source == "You" {
			msg.Source = getPlayerName(configuration.Everquest.LogPath)
		}
		i := strings.Index(msg.Msg, "'")
		parse := strings.ReplaceAll(msg.Msg[i:], "'", "")
		fmt.Fprintf(out, "> %s provided a parse\n```%s```\n", msg.Source, parse)
	}
}

func (p *ParsePlugin) Info(out io.Writer) {
	fmt.Fprintf(out, "---------------\n")
	fmt.Fprintf(out, "Name: %s\n", p.Name)
	fmt.Fprintf(out, "Author: %s\n", p.Author)
	fmt.Fprintf(out, "Version: %s\n", p.Version)
	fmt.Fprintf(out, "---------------\n")
}

func (p *ParsePlugin) OutputChannel() int {
	return p.Output
}
