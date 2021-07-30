package main

import (
	"bytes"
	"testing"
	"time"

	everquest "github.com/Mortimus/goEverquest"
)

func TestFlag(t *testing.T) {
	ldplug := new(FlagPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "say"
	msg.Msg = "Mortimus says, 'Hail, a planar projection'"
	msg.Source = "Mortimus"
	msg.T = time.Now()
	currentZone = "TEST"
	var b bytes.Buffer
	ldplug.Handle(msg, &b)
	got := b.String()
	want := "Mortimus got the flag from TEST\n"
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got, want)
	}
}
