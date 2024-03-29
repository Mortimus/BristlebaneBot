package main

import (
	"bytes"
	"testing"
	"time"

	everquest "github.com/Mortimus/goEverquest"
)

func TestZone(t *testing.T) {
	ldplug := new(ZonePlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	msg.Msg = "You have entered Vex Thal."
	msg.Source = "Mortimus"
	msg.T = time.Now()
	currentZone = ""
	var b bytes.Buffer
	ldplug.Handle(msg, &b)
	got := currentZone
	want := "Vex Thal"
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got, want)
	}
}

func TestZoneLevitate(t *testing.T) {
	ldplug := new(ZonePlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	msg.Msg = "You have entered an area where levitation effects do not function."
	msg.Source = "Mortimus"
	msg.T = time.Now()
	currentZone = ""
	var b bytes.Buffer
	ldplug.Handle(msg, &b)
	got := b.String()
	want := ""
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got, want)
	}
}

func TestZoneBind(t *testing.T) {
	ldplug := new(ZonePlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	msg.Msg = "You have entered an area where Bind Affinity is allowed."
	msg.Source = "Mortimus"
	msg.T = time.Now()
	currentZone = ""
	var b bytes.Buffer
	ldplug.Handle(msg, &b)
	got := b.String()
	want := ""
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got, want)
	}
}
