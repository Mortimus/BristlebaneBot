package main

import (
	"bytes"
	"testing"
	"time"

	everquest "github.com/Mortimus/goEverquest"
)

func TestParse(t *testing.T) {
	plug := new(ParsePlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "Von_parses"
	msg.Msg = "Mortimus tells Von_parses:5, 'Vulak`Aerr in 1344s, 1909k @1420sdps | Bramil + pets 98016@(76dps in 1277s) | Ravnor 97647@(75dps in 1297s) | Gnoro 93471@(72dps in 1289s) | Glooping 89802@(69dps in 1293s) | Wallen 83648@(64dps in 1294s) | Helbinor + pets 82973@(62dps in 1324s) {X} | Blepper 73549@(56dps in 1291s) | Person 70902@(55dps in 1289s) | Boogabooga 67002@(52dps in 1270s) {H} | Yzzy 66324@(51dps in 1290s) | Penelo 63993@(49dps in 1295s) | Baconlegs 63915@(49dps in 1292s) | Abram 61066@(47dps in 1292s) | Ryze 54609@(42dps ...'"
	msg.Source = "Mortimus"
	msg.T = time.Now()
	var b bytes.Buffer
	plug.Handle(msg, &b)
	got := b.String()
	want := "> Mortimus provided a parse\n```Vulak`Aerr in 1344s, 1909k @1420sdps | Bramil + pets 98016@(76dps in 1277s) | Ravnor 97647@(75dps in 1297s) | Gnoro 93471@(72dps in 1289s) | Glooping 89802@(69dps in 1293s) | Wallen 83648@(64dps in 1294s) | Helbinor + pets 82973@(62dps in 1324s) {X} | Blepper 73549@(56dps in 1291s) | Person 70902@(55dps in 1289s) | Boogabooga 67002@(52dps in 1270s) {H} | Yzzy 66324@(51dps in 1290s) | Penelo 63993@(49dps in 1295s) | Baconlegs 63915@(49dps in 1292s) | Abram 61066@(47dps in 1292s) | Ryze 54609@(42dps ...```\n"
	if got != want {
		t.Errorf("plug.Handle(msg, &b) = %q, want %q", got, want)
	}
}
