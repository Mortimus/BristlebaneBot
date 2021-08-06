package main

import (
	"bytes"
	"regexp"
	"testing"
	"time"

	everquest "github.com/Mortimus/goEverquest"
)

func TestSpellLoot(t *testing.T) {
	plug := new(LootPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	msg.Msg = "--Mortimus has looted a Spell: Form of the Great Bear from a glimmer drake's corpse.--"
	msg.Source = "Mortimus"
	msg.T = time.Now()
	plug.LootMatch, _ = regexp.Compile(`--(\w+) has looted a[n]? (.+) from (.+)'s corpse.--`)
	roster["Mortimus"] = &Player{Name: "Mortimus", Class: "Necromancer"}
	var b bytes.Buffer
	plug.Handle(msg, &b)
	got := b.String()
	want := "Mortimus (Necromancer) looted Spell: Form of the Great Bear from a glimmer drake\n"
	if got != want {
		t.Errorf("plug.Handle(msg, &b) = %q, want %q", got, want)
	}
}

func TestAncientLoot(t *testing.T) {
	plug := new(LootPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	msg.Msg = "--Mortimus has looted an Ancient: Master of Death from a glimmer drake's corpse.--"
	msg.Source = "Mortimus"
	msg.T = time.Now()
	plug.LootMatch, _ = regexp.Compile(`--(\w+) has looted a[n]? (.+) from (.+)'s corpse.--`)
	roster["Mortimus"] = &Player{Name: "Mortimus", Class: "Necromancer"}
	var b bytes.Buffer
	plug.Handle(msg, &b)
	got := b.String()
	want := "Mortimus (Necromancer) looted Ancient: Master of Death from a glimmer drake\n"
	if got != want {
		t.Errorf("plug.Handle(msg, &b) = %q, want %q", got, want)
	}
}

func TestLootProvider(t *testing.T) {
	plug := new(LootPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	msg.Msg = "--Mortimus has looted a Glyphed Rune Word from a glimmer drake's corpse.--"
	msg.Source = "Mortimus"
	msg.T = time.Now()
	plug.LootMatch, _ = regexp.Compile(`--(\w+) has looted a[n]? (.+) from (.+)'s corpse.--`)
	roster["Mortimus"] = &Player{Name: "Mortimus", Class: "Necromancer"}
	var b bytes.Buffer
	plug.Handle(msg, &b)
	got := b.String()
	want := "Mortimus (Necromancer) looted Glyphed Rune Word from a glimmer drake\n"
	if got != want {
		t.Errorf("plug.Handle(msg, &b) = %q, want %q", got, want)
	}
}

func TestAwardedLoot(t *testing.T) {
	plug := new(LootPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	msg.Msg = "--Mortimus has looted a Cloth Cap from a glimmer drake's corpse.--"
	msg.Source = "Mortimus"
	msg.T = time.Now()
	plug.LootMatch, _ = regexp.Compile(`--(\w+) has looted a[n]? (.+) from (.+)'s corpse.--`)
	roster["Mortimus"] = &Player{Name: "Mortimus", Class: "Necromancer"}
	needsLooted = []string{"Cloth Cap"}
	var b bytes.Buffer
	plug.Handle(msg, &b)
	got := b.String()
	want := "Mortimus (Necromancer) looted Cloth Cap from a glimmer drake\n"
	if got != want {
		t.Errorf("plug.Handle(msg, &b) = %q, want %q", got, want)
	}
}

// func TestItemDesc(t *testing.T) {
// 	id := 40999
// 	item, _ := itemDB.GetItemByID(id)
// 	want := "Mortimus (Necromancer) looted Cloth Cap from a glimmer drake\n"
// 	got := getItemDesc(item)
// 	fmt.Printf("--%d--\n%s\n", id, got)
// 	if got != want {
// 		t.Errorf("plug.Handle(msg, &b) = %q, want %q", got, want)
// 	}
// }
