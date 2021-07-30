package main

import (
	"bytes"
	"testing"
	"time"

	everquest "github.com/Mortimus/goEverquest"
)

func TestLinkdead(t *testing.T) {
	ldplug := new(LinkdeadPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	msg.Msg = "Mortimus has gone Linkdead."
	msg.Source = "Mortimus"
	msg.T = time.Now()
	var b bytes.Buffer
	ldplug.Handle(msg, &b)
	got := b.String()
	want := "Mortimus has gone Linkdead.\n"
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got, want)
	}
}
