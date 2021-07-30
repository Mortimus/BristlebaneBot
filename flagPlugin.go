package main

import (
	"fmt"
	"io"
	"strings"

	everquest "github.com/Mortimus/goEverquest"
)

type FlagPlugin Plugin

func init() {
	plug := new(FlagPlugin)
	plug.Name = "Flag hailing"
	plug.Author = "Mortimus"
	plug.Version = "1.0.0"
	plug.Output = FLAGOUT
	Handlers = append(Handlers, plug)
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
