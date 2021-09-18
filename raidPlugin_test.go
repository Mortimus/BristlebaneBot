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
	msg.Msg = "Outputfile Complete: RaidRoster_aradune-20210417-205952.txt"
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

func TestRaidBossKillGithub43(t *testing.T) {
	ldplug := new(RaidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	// &main.BossDKP{Zone:"Takish-Hiz: Fading Temple", Note:"LDoN", Boss:"Quintessence Of Sand", DKP:30, FTK:10, IsFTK:false}
	//   LDoN: Quintessence of Sand
	msg.Msg = "Quintessence of Sand has been slain by Mortimus!"
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
	want := "Quintessence of Sand was slain by Mortimus awarding the raid 30 DKP\n"
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got, want)
	}
}

func TestRaidBossKill(t *testing.T) {
	ldplug := new(RaidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	msg.Msg = "Kraksmaal Fir`Dethsin has been slain by Mortimus!"
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
	want := "Kraksmaal Fir`Dethsin was slain by Mortimus awarding the raid 10 DKP\n"
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got, want)
	}
}

func TestRaidBossKillFTK(t *testing.T) {
	ldplug := new(RaidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	//Tacvi: Tunat`Muram Cuu Vauax
	msg.Msg = "Tunat`Muram Cuu Vauax has been slain by Mortimus!"
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
	want := "Tunat`Muram Cuu Vauax was slain by Mortimus awarding the raid 10+10=20 DKP due to FTK\n"
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got, want)
		// printBosses()
	}
}

func TestRaidBossUpload(t *testing.T) {
	ldplug := new(RaidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	msg.Msg = "Outputfile Complete: RaidRoster_aradune-20210417-205952.txt"
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

func TestRaidBossUploadDiff(t *testing.T) {
	ldplug := new(RaidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	msg.Msg = "Outputfile Complete: RaidRoster_aradune-20210417-205952.txt"
	msg.Source = "Mortimus"
	msg.T = time.Date(2021, time.April, 17, 20, 59, 52, 0, time.Local)
	ldplug.Output = TESTOUT // anything but raid dump channel
	ldplug.NeedsDump = false
	ldplug.NextDump = msg.T.Add(time.Hour * 5)
	ldplug.LastBoss = "TestBoss"
	ldplug.LastRaid = everquest.Raid{}
	ldplug.LastRaid.LoadFromPath(configuration.Everquest.BaseFolder+"/"+"RaidRoster_aradune-20201122-180031.txt", Err)
	ldplug.Started = true
	ldplug.Bosses += 2
	var b bytes.Buffer
	ldplug.Handle(msg, &b)
	got := b.String()
	want := "```diff\n+ Bids\n``````diff\n+ Amanar\n``````diff\n+ Guzz\n``````diff\n+ Daangerzone\n``````diff\n+ Beandip\n``````diff\n+ Canniblepper\n``````diff\n+ Blepper\n``````diff\n+ Kickfu\n``````diff\n+ Cronos\n``````diff\n+ Perc\n``````diff\n+ Mortimus\n``````diff\n+ Advenia\n``````diff\n+ Rabidtiger\n``````diff\n+ Kejek\n``````diff\n+ Kirynn\n``````diff\n+ Xarielx\n``````diff\n+ Tomdar\n``````diff\n+ Doctorbear\n``````diff\n+ Tators\n``````diff\n+ Rinon\n``````diff\n+ Wonders\n``````diff\n+ Helbinor\n``````diff\n+ Tyrannikal\n``````diff\n+ Ticklez\n``````diff\n+ Joule\n``````diff\n+ Valcis\n``````diff\n+ Thasumr\n``````diff\n+ Banis\n``````diff\n+ Nistalkin\n``````diff\n+ Boomerbear\n``````diff\n- Person\n``````diff\n- Crasis\n``````diff\n- Iovelost\n``````diff\n- Tigermancer\n``````diff\n- Galdo\n``````diff\n- Bunzz\n``````diff\n- Ayamnivay\n``````diff\n- Ravnor\n``````diff\n- Torsey\n``````diff\n- Cadenza\n``````diff\n- Coltaine\n``````diff\n- Mysfit\n``````diff\n- Dromi\n``````diff\n- Nosirrah\n``````diff\n- Talen\n``````diff\n- Utair\n``````diff\n- Whidon\n``````diff\n- Mollwin\n``````diff\n- Iilenye\n``````diff\n- Haldemir\n``````diff\n- Glooping\n``````diff\n- Wallen\n``````diff\n- Rokem\n``````diff\n- Fluffer\n``````diff\n- Milliardo\n``````diff\n- Ryze\n``````diff\n- Rost\n```Uploading Raid Dump: 20210417_TestBoss_2.txt\n"
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got, want)
	}
}

func TestRaidHourly(t *testing.T) {
	ldplug := new(RaidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "system"
	msg.Msg = "Outputfile Complete: RaidRoster_aradune-20210417-205952.txt"
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
