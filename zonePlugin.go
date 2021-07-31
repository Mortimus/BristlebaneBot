package main

import (
	"fmt"
	"io"
	"strings"

	everquest "github.com/Mortimus/goEverquest"
)

var currentZone string

type ZonePlugin Plugin

func init() {
	ldplug := new(ZonePlugin)
	ldplug.Name = "Zone detection"
	ldplug.Author = "Mortimus"
	ldplug.Version = "1.0.0"
	ldplug.Output = STDOUT
	Handlers = append(Handlers, ldplug)
}

// Handle for LinkdeadPlugin sends a message if it detects a player has gone linkdead.
func (p *ZonePlugin) Handle(msg *everquest.EqLog, out io.Writer) {
	if msg.Channel == "system" && strings.Contains(msg.Msg, "You have entered ") && !strings.Contains(msg.Msg, "function.") && !strings.Contains(msg.Msg, "Bind Affinity") { // You have entered Vex Thal. NOT You have entered an area where levitation effects do not function.
		currentZone = msg.Msg[17 : len(msg.Msg)-1]
		fmt.Fprintf(out, "Changing zone to %s\n", currentZone)

	}
}

func (p *ZonePlugin) Info(out io.Writer) {
	fmt.Fprintf(out, "---------------\n")
	fmt.Fprintf(out, "Name: %s\n", p.Name)
	fmt.Fprintf(out, "Author: %s\n", p.Author)
	fmt.Fprintf(out, "Version: %s\n", p.Version)
	fmt.Fprintf(out, "---------------\n")
}

func (p *ZonePlugin) OutputChannel() int {
	return p.Output
}
