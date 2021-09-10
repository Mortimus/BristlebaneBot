package main

import (
	"bytes"
	"regexp"
	"testing"
	"time"

	everquest "github.com/Mortimus/goEverquest"
)

func TestRoll(t *testing.T) {
	plug := new(RollPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	msg.Msg = "**A Magic Die is rolled by Mortimus. It could have been any number from 0 to 1000, but this time it turned up a 894."
	msg.Source = "You"
	msg.T = time.Now()
	var b bytes.Buffer
	needsRolled = append(needsRolled, "Mortimus")
	plug.RollMatch, _ = regexp.Compile(`\*\*A Magic Die is rolled by (\w+). It could have been any number from (\d+) to (\d+), but this time it turned up a (\d+).`)
	plug.Handle(msg, &b)
	got := b.String()
	want := "```ini\n[Mortimus rolled a 894]\n```"
	if got != want {
		t.Errorf("plug.Handle(msg, &b) = %q, want %q", got, want)
	}
	needsRolled = []string{}
}

func TestWrongLowRoll(t *testing.T) {
	plug := new(RollPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	msg.Msg = "**A Magic Die is rolled by Mortimus. It could have been any number from 2 to 1000, but this time it turned up a 894."
	msg.Source = "You"
	msg.T = time.Now()
	var b bytes.Buffer
	needsRolled = append(needsRolled, "Mortimus")
	plug.RollMatch, _ = regexp.Compile(`\*\*A Magic Die is rolled by (\w+). It could have been any number from (\d+) to (\d+), but this time it turned up a (\d+).`)
	plug.Handle(msg, &b)
	got := b.String()
	want := ""
	if got != want {
		t.Errorf("plug.Handle(msg, &b) = %q, want %q", got, want)
	}
	needsRolled = []string{}
}

func TestWrongHighRoll(t *testing.T) {
	plug := new(RollPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	msg.Msg = "**A Magic Die is rolled by Mortimus. It could have been any number from 0 to 999, but this time it turned up a 894."
	msg.Source = "You"
	msg.T = time.Now()
	var b bytes.Buffer
	needsRolled = append(needsRolled, "Mortimus")
	plug.RollMatch, _ = regexp.Compile(`\*\*A Magic Die is rolled by (\w+). It could have been any number from (\d+) to (\d+), but this time it turned up a (\d+).`)
	plug.Handle(msg, &b)
	got := b.String()
	want := ""
	if got != want {
		t.Errorf("plug.Handle(msg, &b) = %q, want %q", got, want)
	}
	needsRolled = []string{}
}

func TestWrongLowHighRoll(t *testing.T) {
	plug := new(RollPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	msg.Msg = "**A Magic Die is rolled by Mortimus. It could have been any number from 2 to 999, but this time it turned up a 894."
	msg.Source = "You"
	msg.T = time.Now()
	var b bytes.Buffer
	needsRolled = append(needsRolled, "Mortimus")
	plug.RollMatch, _ = regexp.Compile(`\*\*A Magic Die is rolled by (\w+). It could have been any number from (\d+) to (\d+), but this time it turned up a (\d+).`)
	plug.Handle(msg, &b)
	got := b.String()
	want := ""
	if got != want {
		t.Errorf("plug.Handle(msg, &b) = %q, want %q", got, want)
	}
	needsRolled = []string{}
}

func TestMysteryRoll(t *testing.T) {
	plug := new(RollPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	msg.Msg = "**A Magic Die is rolled by Mortimus. It could have been any number from 0 to 1000, but this time it turned up a 894."
	msg.Source = "You"
	msg.T = time.Now()
	var b bytes.Buffer
	needsRolled = append(needsRolled, "Penelo")
	plug.RollMatch, _ = regexp.Compile(`\*\*A Magic Die is rolled by (\w+). It could have been any number from (\d+) to (\d+), but this time it turned up a (\d+).`)
	plug.Handle(msg, &b)
	got := b.String()
	want := ""
	if got != want {
		t.Errorf("plug.Handle(msg, &b) = %q, want %q", got, want)
	}
	needsRolled = []string{}
}
