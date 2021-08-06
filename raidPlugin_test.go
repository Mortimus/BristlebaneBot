package main

import (
	"bytes"
	"regexp"
	"testing"
	"time"

	everquest "github.com/Mortimus/goEverquest"
)

func TestRaidStart(t *testing.T) {
	ldplug := new(RaidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	msg.Msg = "Outputfile Complete: RaidRoster_aradune-20210417-205952.txt'"
	msg.Source = "Mortimus"
	msg.T = time.Date(2021, time.April, 17, 20, 59, 52, 0, time.Local)
	ldplug.Output = TESTOUT // anything but raid dump channel
	ldplug.NeedsDump = true
	var b bytes.Buffer
	ldplug.Handle(msg, &b)
	got := b.String()
	want := "Uploading Raid Dump: 20210417_raid_start.txt\n"
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got, want)
	}
}

func TestRaidBossKill(t *testing.T) {
	ldplug := new(RaidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	msg.Msg = "Kraksmaal Fir'Dethsin has been slain by Mortimus!'"
	msg.Source = "Mortimus"
	msg.T = time.Date(2021, time.April, 17, 20, 59, 52, 0, time.Local)
	ldplug.Output = TESTOUT // anything but raid dump channel
	ldplug.NeedsDump = false
	ldplug.NextDump = msg.T.Add(time.Hour * 5)
	ldplug.SlayMatch, _ = regexp.Compile(`(.+) has been slain by (\w+)!`)
	ldplug.Started = true
	var b bytes.Buffer
	ldplug.Handle(msg, &b)
	got := b.String()
	want := "Kraksmaal Fir'Dethsin was slain by Mortimus awarding the raid DKP\n"
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got, want)
	}
}

func TestRaidBossUpload(t *testing.T) {
	ldplug := new(RaidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	msg.Msg = "Outputfile Complete: RaidRoster_aradune-20210417-205952.txt'"
	msg.Source = "Mortimus"
	msg.T = time.Date(2021, time.April, 17, 20, 59, 52, 0, time.Local)
	ldplug.Output = TESTOUT // anything but raid dump channel
	ldplug.NeedsDump = false
	ldplug.NextDump = msg.T.Add(time.Hour * 5)
	ldplug.LastBoss = "TestBoss"
	ldplug.Bosses += 2
	var b bytes.Buffer
	ldplug.Handle(msg, &b)
	got := b.String()
	want := "Uploading Raid Dump: 20210417_TestBoss_2.txt\n"
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got, want)
	}
}

func TestRaidHourly(t *testing.T) {
	ldplug := new(RaidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	msg.Msg = "Outputfile Complete: RaidRoster_aradune-20210417-205952.txt'"
	msg.Source = "Mortimus"
	msg.T = time.Date(2021, time.April, 17, 20, 59, 52, 0, time.Local)
	ldplug.Output = TESTOUT // anything but raid dump channel
	ldplug.NeedsDump = true
	ldplug.Hours++
	var b bytes.Buffer
	ldplug.Handle(msg, &b)
	got := b.String()
	want := "Uploading Raid Dump: 20210417_hour_1.txt\n"
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got, want)
	}
}

func TestRaidDumpReminder(t *testing.T) {
	ldplug := new(RaidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "say"
	msg.Msg = "Mortimus says, 'This Log doesn't matter'"
	msg.Source = "Mortimus"
	msg.T = time.Date(2021, time.April, 17, 20, 59, 52, 0, time.Local)
	ldplug.Output = TESTOUT // anything but raid dump channel
	ldplug.NeedsDump = false
	ldplug.NextDump = msg.T
	ldplug.Started = true
	currentTime = msg.T
	ldplug.Hours++
	var b bytes.Buffer
	ldplug.Handle(msg, &b)
	got := b.String()
	want := "Time for another hourly raid dump!\n"
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got, want)
	}
}

func TestRaidNoDumpReminder(t *testing.T) {
	ldplug := new(RaidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "say"
	msg.Msg = "Mortimus says, 'This Log doesn't matter'"
	msg.Source = "Mortimus"
	msg.T = time.Date(2021, time.April, 17, 20, 59, 52, 0, time.Local)
	ldplug.Output = TESTOUT // anything but raid dump channel
	ldplug.NeedsDump = false
	ldplug.NextDump = msg.T
	ldplug.Started = true
	currentTime = msg.T.Add(time.Minute * 30)
	ldplug.Hours++
	var b bytes.Buffer
	ldplug.Handle(msg, &b)
	got := b.String()
	want := ""
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got, want)
	}
}
