package main

import (
	"fmt"
	"io"
	"strings"

	everquest "github.com/Mortimus/goEverquest"
)

type LinkdeadPlugin Plugin

func init() {
	ldplug := new(LinkdeadPlugin)
	ldplug.Name = "Linkdead detection"
	ldplug.Author = "Mortimus"
	ldplug.Version = "1.0.0"
	ldplug.Output = RAIDOUT
	Handlers = append(Handlers, ldplug)
}

// Handle for LinkdeadPlugin sends a message if it detects a player has gone linkdead.
func (p *LinkdeadPlugin) Handle(msg *everquest.EqLog, out io.Writer) {
	if msg.Channel == "system" && strings.Contains(msg.Msg, "has gone Linkdead.") {
		fmt.Fprintf(out, "%s\n", msg.Msg)
	}
}

func (p *LinkdeadPlugin) Info(out io.Writer) {
	fmt.Fprintf(out, "---------------\n")
	fmt.Fprintf(out, "Name: %s\n", p.Name)
	fmt.Fprintf(out, "Author: %s\n", p.Author)
	fmt.Fprintf(out, "Version: %s\n", p.Version)
	fmt.Fprintf(out, "---------------\n")
}

func (p *LinkdeadPlugin) OutputChannel() int {
	return p.Output
}
