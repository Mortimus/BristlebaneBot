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
	plug.LootMatch, _ = regexp.Compile(`--(\w+) has looted a[n]? (.+) from (.+)['s corpse]?[ ]?\.--`)
	roster["Mortimus"] = &Player{Name: "Mortimus", Class: "Necromancer"}
	var b bytes.Buffer
	plug.Handle(msg, &b)
	got := b.String()
	want := "> Mortimus (Necromancer) looted Spell: Form of the Great Bear from a glimmer drake's corpse\n```Spell: Form of the Great Bear\nMAGIC \nSlot: NONE \n\nWT: 0.1 Size: SMALL\nClass: SHM  \nRace: ALL ```\n"
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
	plug.LootMatch, _ = regexp.Compile(`--(\w+) has looted a[n]? (.+) from (.+)['s corpse]?[ ]?\.--`)
	roster["Mortimus"] = &Player{Name: "Mortimus", Class: "Necromancer"}
	var b bytes.Buffer
	plug.Handle(msg, &b)
	got := b.String()
	want := "> Mortimus (Necromancer) looted Ancient: Master of Death from a glimmer drake's corpse\n```Ancient: Master of Death\nMAGIC NO TRADE \nSlot: NONE \n\nWT: 0.1 Size: SMALL\nClass: NEC  \nRace: NONE ```\n"
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
	plug.LootMatch, _ = regexp.Compile(`--(\w+) has looted a[n]? (.+) from (.+)['s corpse]?[ ]?\.--`)
	roster["Mortimus"] = &Player{Name: "Mortimus", Class: "Necromancer"}
	var b bytes.Buffer
	plug.Handle(msg, &b)
	got := b.String()
	want := "> Mortimus (Necromancer) looted Glyphed Rune Word from a glimmer drake's corpse\n```Glyphed Rune Word\nMAGIC NO TRADE \nSlot: NONE \n\nWT: 0.1 Size: TINY\nClass: NONE \nRace: NONE ```\n"
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
	plug.LootMatch, _ = regexp.Compile(`--(\w+) has looted a[n]? (.+) from (.+)['s corpse]?[ ]?\.--`)
	roster["Mortimus"] = &Player{Name: "Mortimus", Class: "Necromancer"}
	needsLooted = []string{"Cloth Cap"}
	var b bytes.Buffer
	plug.Handle(msg, &b)
	got := b.String()
	want := "> Mortimus (Necromancer) looted Cloth Cap from a glimmer drake's corpse\n```Cloth Cap\nMAGIC LORE NO TRADE \nSlot: NONE \n\nEffect: Veeshan's Swarm \nWT: 0.5 Size: SMALL\nClass: ALL \nRace: ALL ```\n"
	if got != want {
		t.Errorf("plug.Handle(msg, &b) = %q, want %q", got, want)
	}
}

func TestItemDesc(t *testing.T) {
	id := 11621
	item, _ := itemDB.GetItemByID(id)
	want := "Cloak of Flames\nMAGIC \nSlot: BACK  \nAC: 10\nDEX: +9 AGI: +9 HP: +50 \nSV FIRE: +15 \nHaste: +36% \nWT: 0.1 Size: MEDIUM\nClass: ALL \nRace: ALL \nSlot 1, Type 7 (General: Group)"
	got := getItemDesc(item)
	// fmt.Printf("--%d--\n%s\n", id, got)
	if got != want {
		t.Errorf("plug.Handle(msg, &b) = %q, want %q", got, want)
	}
}

func TestAwardedSelfLoot(t *testing.T) {
	plug := new(LootPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	msg.Msg = "--You have looted a Cloth Cap from a glimmer drake's corpse.--"
	msg.Source = "You"
	msg.T = time.Now()
	plug.LootMatch, _ = regexp.Compile(`--(\w+) ha\w{1,2} looted a[n]? (.+) from (.+)['s corpse]?[ ]?\.--`)
	roster["Mortimus"] = &Player{Name: "Mortimus", Class: "Necromancer"}
	needsLooted = []string{"Cloth Cap"}
	var b bytes.Buffer
	plug.Handle(msg, &b)
	got := b.String()
	want := "> Mortimus (Necromancer) looted Cloth Cap from a glimmer drake's corpse\n```Cloth Cap\nMAGIC LORE NO TRADE \nSlot: NONE \n\nEffect: Veeshan's Swarm \nWT: 0.5 Size: SMALL\nClass: ALL \nRace: ALL ```\n"
	if got != want {
		t.Errorf("plug.Handle(msg, &b) = %q, want %q", got, want)
	}
}

func TestInferredLoot(t *testing.T) {
	plug := new(LootPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	msg.Msg = "--Mortimus has looted a Timeless Silk Robe Pattern from a Quarm's corpse.--"
	msg.Source = "Mortimus"
	msg.T = time.Now()
	plug.LootMatch, _ = regexp.Compile(`--(\w+) has looted a[n]? (.+) from (.+)['s corpse]?[ ]?\.--`)
	roster["Mortimus"] = &Player{Name: "Mortimus", Class: "Necromancer"}
	needsLooted = []string{"Timeless Silk Robe Pattern"}
	var b bytes.Buffer
	plug.Handle(msg, &b)
	got := b.String()
	want := "> Mortimus (Necromancer) looted Miragul's Shroud of Risen Souls from a Quarm's corpse\n```Miragul's Shroud of Risen Souls\nMAGIC LORE NO TRADE \nSlot: CHEST  \nAC: 35\nSkill Mod: Specialize Alteration +8% (21 Max)\nDEX: +25 STA: +30 CHA: +15 WIS: +0+5 INT: +30+5 AGI: +25 HP: +185 MANA: +200 \nSV FIRE: +20 SV DISEASE: +20 SV COLD: +20 SV MAGIC: +20 SV POISON: +20 \nRequired level of 65 \nFocus: Vengeance of Eternity \nWT: 1.0 Size: LARGE\nClass: NEC  \nRace: ALL \nSlot 1, Type 8 (General: Raid)\nSlot 2, Type 21 (Special Ornamentation)```\n"
	if got != want {
		t.Errorf("plug.Handle(msg, &b) = %q, want %q", got, want)
	}
}
