package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"regexp"
	"testing"
	"time"

	everquest "github.com/Mortimus/goEverquest"
)

func TestBidOpen(t *testing.T) {
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Tt][Ee][Ll][Ll][Ss])?([Bb][Ii][Dd][Ss])?\sto\s.+,?.+(\d+).*`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	got := b.String()
	want := "> Bids open on Cloth Cap (x1) for 2 minutes.\n```Cloth Cap\nMAGIC LORE NO TRADE \nSlot: NONE \n\nEffect: Veeshan's Swarm \nWT: 0.5 Size: SMALL\nClass: ALL \nRace: ALL ```https://eq.magelo.com/item/42984"
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got, want)
	}
}

func TestBidChangeQuantity(t *testing.T) {
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	id, _ := itemDB.FindIDByName("Cloth Cap")
	msg.Msg = "Cloth Cap bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Tt][Ee][Ll][Ll][Ss])?([Bb][Ii][Dd][Ss])?\sto\s.+,?.+(\d+).*`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	oQuantity := plug.Bids[id].Quantity
	msgTwo := new(everquest.EqLog)
	msgTwo.Channel = "guild"
	msgTwo.Msg = "Cloth Capx3 bids to Bids, pst 2min"
	msgTwo.Source = "You"
	msgTwo.T = time.Now()
	plug.Handle(msgTwo, &b)
	nQuantity := plug.Bids[id].Quantity
	// got := b.String()
	// want := "Bids open on Cloth Cap(x1) for 2 minutes.\n"
	if nQuantity != 3 && oQuantity != nQuantity {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", oQuantity, nQuantity)
	}
}

func TestBidClose(t *testing.T) {
	plug := new(BidPlugin)
	plug.Output = STDOUT
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap bids to Bids, CLOSED"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Tt][Ee][Ll][Ll][Ss])?([Bb][Ii][Dd][Ss])?\sto\s.+,?.+(\d+).*`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	id, _ := itemDB.FindIDByName("Cloth Cap")
	item, _ := itemDB.GetItemByID(id)
	plug.Bids[id] = &OpenBid{
		Item:     item,
		Quantity: 1,
		Duration: 2 * time.Minute,
		Start:    time.Now(),
		End:      time.Now().Add(2 * time.Minute),
		Bidders:  []*Bidder{},
	}
	var b bytes.Buffer
	plug.Handle(msg, &b)
	got := b.String()
	want := "> Cloth Cap (x1) won for 0 DKP\n```1: Rot\n```[b1c9e6fb16cd0d49734d45ec331b5e008049e734ebe90a61284b835d39e19414]> Closed bids on Cloth Cap (x1)\n"
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got, want)
	}
}

func TestGetDKPRankInactive(t *testing.T) {
	member := &everquest.GuildMember{
		Rank: "Inactive",
	}
	got := getDKPRank(member)
	want := DKPRank(INACTIVE)
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", DKPRankToString(got), DKPRankToString(want))
	}
}

func TestGetDKPRankLeaderMain(t *testing.T) {
	member := &everquest.GuildMember{
		Rank: "<<< Guild Leader >>>",
		Alt:  false,
	}
	got := getDKPRank(member)
	want := DKPRank(MAIN)
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", DKPRankToString(got), DKPRankToString(want))
	}
}

func TestGetDKPRankLeaderSecondMain(t *testing.T) {
	member := &everquest.GuildMember{
		Rank:       "<<< Guild Leader >>>",
		Alt:        true,
		PublicNote: "Geban's Second Main",
	}
	got := getDKPRank(member)
	want := DKPRank(SECONDMAIN)
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", DKPRankToString(got), DKPRankToString(want))
	}
}
func TestGetDKPRankOfficerSecondMain(t *testing.T) {
	member := &everquest.GuildMember{
		Rank:       "<<< Officer >>>",
		Alt:        true,
		PublicNote: "Mortimus' 2nd main",
	}
	got := getDKPRank(member)
	want := DKPRank(SECONDMAIN)
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", DKPRankToString(got), DKPRankToString(want))
	}
}

func TestGetDKPRankRecruit(t *testing.T) {
	member := &everquest.GuildMember{
		Rank: "Recruit",
	}
	got := getDKPRank(member)
	want := DKPRank(RECRUIT)
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", DKPRankToString(got), DKPRankToString(want))
	}
}

func TestGetDKPRankSocial(t *testing.T) {
	member := &everquest.GuildMember{
		Rank: "Member",
	}
	got := getDKPRank(member)
	want := DKPRank(SOCIAL)
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", DKPRankToString(got), DKPRankToString(want))
	}
}

func TestGetDKPRankAlt(t *testing.T) {
	member := &everquest.GuildMember{
		Rank: "Member",
		Alt:  true,
	}
	got := getDKPRank(member)
	want := DKPRank(ALT)
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", DKPRankToString(got), DKPRankToString(want))
	}
}

func TestBidAdd(t *testing.T) {
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Tt][Ee][Ll][Ll][Ss])?([Bb][Ii][Dd][Ss])?\sto\s.+,?.+(\d+).*`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+)\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Cloth Cap 500"
	add.Source = "Mortimus"
	add.T = time.Now()
	plug.Handle(add, &b)
	id, _ := itemDB.FindIDByName("Cloth Cap")
	got := plug.Bids[id].FindBid("Mortimus")
	want := 0
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %d, want %d", got, want)
	}
}

func TestBidAddAmount(t *testing.T) {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	itemID := rand.Intn(8000-6000+1) + 6000 // TODO: update to have real range of itemDB
	randomItem, _ := itemDB.GetItemByID(itemID)
	msg.Msg = fmt.Sprintf("%s bids to Bids, pst 2min", randomItem.Name)
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Tt][Ee][Ll][Ll][Ss])?([Bb][Ii][Dd][Ss])?\sto\s.+,?.+(\d+).*`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+)\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	bidAmount := r1.Int()
	add.Source = "Mortimus"
	add.T = time.Now()

	add.Msg = fmt.Sprintf("%s %d", randomItem.Name, bidAmount)
	plug.Handle(add, &b)
	id, _ := itemDB.FindIDByName(randomItem.Name)
	bidder := plug.Bids[id].FindBid("Mortimus")
	if bidder < 0 {
		t.Errorf("ldplug.Handle(msg, &b) = %d, want %s", bidder, "positive number")
	}
	got := plug.Bids[id].Bidders[bidder].AttemptedBid
	want := bidAmount
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %d, want %d", got, want)
	}
}

func TestBidApply(t *testing.T) {
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Tt][Ee][Ll][Ll][Ss])?([Bb][Ii][Dd][Ss])?\sto\s.+,?.+(\d+).*`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+)\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Cloth Cap 500"
	add.Source = "Mortimus"
	add.T = time.Now()
	plug.Handle(add, &b)
	id, _ := itemDB.FindIDByName("Cloth Cap")
	got := plug.Bids[id].FindBid("Mortimus")
	plug.Bids[id].Bidders[got].Player.DKP = 2000
	plug.Bids[id].ApplyDKP()
	appliedBid := plug.Bids[id].Bidders[got].Bid
	want := 500
	if appliedBid != want {
		t.Errorf("ldplug.Handle(msg, &b) = %d, want %d", appliedBid, want)
	}
}

func TestBidApplyTooMuch(t *testing.T) {
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Tt][Ee][Ll][Ll][Ss])?([Bb][Ii][Dd][Ss])?\sto\s.+,?.+(\d+).*`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+)\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Cloth Cap 5000"
	add.Source = "Mortimus"
	add.T = time.Now()
	plug.Handle(add, &b)
	id, _ := itemDB.FindIDByName("Cloth Cap")
	got := plug.Bids[id].FindBid("Mortimus")
	plug.Bids[id].Bidders[got].Player.DKP = 2000
	plug.Bids[id].ApplyDKP()
	appliedBid := plug.Bids[id].Bidders[got].Bid
	want := 2000
	if appliedBid != want {
		t.Errorf("ldplug.Handle(msg, &b) = %d, want %d", appliedBid, want)
	}
}

func TestBidApplyBelowMin(t *testing.T) {
	configuration.Bids.MinimumBid = 10
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Tt][Ee][Ll][Ll][Ss])?([Bb][Ii][Dd][Ss])?\sto\s.+,?.+(\d+).*`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+)\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Cloth Cap 5"
	add.Source = "Mortimus"
	add.T = time.Now()
	plug.Handle(add, &b)
	id, _ := itemDB.FindIDByName("Cloth Cap")
	got := plug.Bids[id].FindBid("Mortimus")
	plug.Bids[id].Bidders[got].Player.DKP = 2000
	plug.Bids[id].ApplyDKP()
	appliedBid := plug.Bids[id].Bidders[got].Bid
	want := configuration.Bids.MinimumBid
	if appliedBid != want {
		t.Errorf("ldplug.Handle(msg, &b) = %d, want %d", appliedBid, want)
	}
}

func TestBidApplyNoIncrement(t *testing.T) {
	configuration.Bids.MinimumBid = 10
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Tt][Ee][Ll][Ll][Ss])?([Bb][Ii][Dd][Ss])?\sto\s.+,?.+(\d+).*`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+)\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Cloth Cap 13"
	add.Source = "Mortimus"
	add.T = time.Now()
	plug.Handle(add, &b)
	id, _ := itemDB.FindIDByName("Cloth Cap")
	got := plug.Bids[id].FindBid("Mortimus")
	plug.Bids[id].Bidders[got].Player.DKP = 2000
	plug.Bids[id].ApplyDKP()
	appliedBid := plug.Bids[id].Bidders[got].Bid
	want := configuration.Bids.MinimumBid
	if appliedBid != want {
		t.Errorf("ldplug.Handle(msg, &b) = %d, want %d", appliedBid, want)
	}
}

func TestBidApplyCancelledBid(t *testing.T) {
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Tt][Ee][Ll][Ll][Ss])?([Bb][Ii][Dd][Ss])?\sto\s.+,?.+(\d+).*`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+)\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Cloth Cap 0"
	add.Source = "Mortimus"
	add.T = time.Now()
	plug.Handle(add, &b)
	id, _ := itemDB.FindIDByName("Cloth Cap")
	got := plug.Bids[id].FindBid("Mortimus")
	plug.Bids[id].Bidders[got].Player.DKP = 2000
	plug.Bids[id].ApplyDKP()
	appliedBid := plug.Bids[id].Bidders[got].Bid
	want := 0
	if appliedBid != want {
		t.Errorf("ldplug.Handle(msg, &b) = %d, want %d", appliedBid, want)
	}
}

func TestBidApplyNerfedSecondMain(t *testing.T) {
	configuration.Bids.SecondMainsBidAsMains = true
	configuration.Bids.SecondMainAsMainMaxBid = 200
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Tt][Ee][Ll][Ll][Ss])?([Bb][Ii][Dd][Ss])?\sto\s.+,?.+(\d+).*`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+)\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Cloth Cap 300"
	add.Source = "Mortimus"
	add.T = time.Now()
	plug.Handle(add, &b)
	secondadd := new(everquest.EqLog)
	secondadd.Channel = "tell"
	secondadd.Msg = "Cloth Cap 3000"
	secondadd.Source = "Milliardo"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Cloth Cap")
	maingot := plug.Bids[id].FindBid("Mortimus")
	plug.Bids[id].Bidders[maingot].Player.DKP = 2000
	plug.Bids[id].Bidders[maingot].Player.DKPRank = MAIN
	secondgot := plug.Bids[id].FindBid("Milliardo")
	plug.Bids[id].Bidders[secondgot].Player.DKP = 2000
	plug.Bids[id].Bidders[secondgot].Player.DKPRank = SECONDMAIN
	plug.Bids[id].ApplyDKP()
	appliedBid := plug.Bids[id].Bidders[secondgot].Bid
	want := configuration.Bids.SecondMainAsMainMaxBid
	if appliedBid != want {
		t.Errorf("ldplug.Handle(msg, &b) = %d, want %d", appliedBid, want)
	}
}

func TestTiesSameRank(t *testing.T) {
	configuration.Bids.SecondMainsBidAsMains = true
	configuration.Bids.SecondMainAsMainMaxBid = 200
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Tt][Ee][Ll][Ll][Ss])?([Bb][Ii][Dd][Ss])?\sto\s.+,?.+(\d+).*`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+)\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Cloth Cap 300"
	add.Source = "Mortimus"
	add.T = time.Now()
	plug.Handle(add, &b)
	secondadd := new(everquest.EqLog)
	secondadd.Channel = "tell"
	secondadd.Msg = "Cloth Cap 300"
	secondadd.Source = "Milliardo"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Cloth Cap")
	maingot := plug.Bids[id].FindBid("Mortimus")
	plug.Bids[id].Bidders[maingot].Player.DKP = 2000
	plug.Bids[id].Bidders[maingot].Player.DKPRank = MAIN
	secondgot := plug.Bids[id].FindBid("Milliardo")
	plug.Bids[id].Bidders[secondgot].Player.DKP = 2000
	plug.Bids[id].Bidders[secondgot].Player.DKPRank = MAIN
	plug.Bids[id].ApplyDKP()
	plug.Bids[id].SortBids()
	ties := plug.Bids[id].CheckTiesAndApplyWinners()
	got := len(ties)
	want := 2
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %d, want %d", got, want)
	}
}

func TestTiesSameRankMulti(t *testing.T) {
	configuration.Bids.SecondMainsBidAsMains = true
	configuration.Bids.SecondMainAsMainMaxBid = 200
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Tt][Ee][Ll][Ll][Ss])?([Bb][Ii][Dd][Ss])?\sto\s.+,?.+(\d+).*`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+)\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Cloth Cap 300"
	add.Source = "Mortimus"
	add.T = time.Now()
	plug.Handle(add, &b)
	secondadd := new(everquest.EqLog)
	secondadd.Channel = "tell"
	secondadd.Msg = "Cloth Cap 300"
	secondadd.Source = "Milliardo"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	thirdadd := new(everquest.EqLog)
	thirdadd.Channel = "tell"
	thirdadd.Msg = "Cloth Cap 300"
	thirdadd.Source = "Penelo"
	thirdadd.T = time.Now()
	plug.Handle(thirdadd, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Cloth Cap")
	maingot := plug.Bids[id].FindBid("Mortimus")
	plug.Bids[id].Bidders[maingot].Player.DKP = 2000
	plug.Bids[id].Bidders[maingot].Player.DKPRank = MAIN
	secondgot := plug.Bids[id].FindBid("Milliardo")
	plug.Bids[id].Bidders[secondgot].Player.DKP = 2000
	plug.Bids[id].Bidders[secondgot].Player.DKPRank = MAIN
	thirdgot := plug.Bids[id].FindBid("Penelo")
	plug.Bids[id].Bidders[thirdgot].Player.DKP = 2000
	plug.Bids[id].Bidders[thirdgot].Player.DKPRank = MAIN
	plug.Bids[id].ApplyDKP()
	plug.Bids[id].SortBids()
	ties := plug.Bids[id].CheckTiesAndApplyWinners()
	got := len(ties)
	want := 3
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %d, want %d", got, want)
	}
}

func TestTiesSecondAndThirdTie(t *testing.T) {
	configuration.Bids.SecondMainsBidAsMains = true
	configuration.Bids.SecondMainAsMainMaxBid = 200
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Capx2 bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Tt][Ee][Ll][Ll][Ss])?([Bb][Ii][Dd][Ss])?\sto\s.+,?.+(\d+).*`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+)\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Cloth Cap 500"
	add.Source = "Mortimus"
	add.T = time.Now()
	plug.Handle(add, &b)
	secondadd := new(everquest.EqLog)
	secondadd.Channel = "tell"
	secondadd.Msg = "Cloth Cap 300"
	secondadd.Source = "Milliardo"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	thirdadd := new(everquest.EqLog)
	thirdadd.Channel = "tell"
	thirdadd.Msg = "Cloth Cap 300"
	thirdadd.Source = "Penelo"
	thirdadd.T = time.Now()
	plug.Handle(thirdadd, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Cloth Cap")
	maingot := plug.Bids[id].FindBid("Mortimus")
	plug.Bids[id].Bidders[maingot].Player.DKP = 2000
	plug.Bids[id].Bidders[maingot].Player.DKPRank = MAIN
	secondgot := plug.Bids[id].FindBid("Milliardo")
	plug.Bids[id].Bidders[secondgot].Player.DKP = 2000
	plug.Bids[id].Bidders[secondgot].Player.DKPRank = MAIN
	thirdgot := plug.Bids[id].FindBid("Penelo")
	plug.Bids[id].Bidders[thirdgot].Player.DKP = 2000
	plug.Bids[id].Bidders[thirdgot].Player.DKPRank = MAIN
	plug.Bids[id].ApplyDKP()
	plug.Bids[id].SortBids()
	ties := plug.Bids[id].CheckTiesAndApplyWinners()
	got := len(ties)
	want := 2
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %d, want %d", got, want)
	}
}

func TestTiesCancelledTies(t *testing.T) {
	configuration.Bids.SecondMainsBidAsMains = true
	configuration.Bids.SecondMainAsMainMaxBid = 200
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Capx2 bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Tt][Ee][Ll][Ll][Ss])?([Bb][Ii][Dd][Ss])?\sto\s.+,?.+(\d+).*`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+)\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Cloth Cap 0"
	add.Source = "Mortimus"
	add.T = time.Now()
	plug.Handle(add, &b)
	secondadd := new(everquest.EqLog)
	secondadd.Channel = "tell"
	secondadd.Msg = "Cloth Cap 0"
	secondadd.Source = "Milliardo"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	thirdadd := new(everquest.EqLog)
	thirdadd.Channel = "tell"
	thirdadd.Msg = "Cloth Cap 0"
	thirdadd.Source = "Penelo"
	thirdadd.T = time.Now()
	plug.Handle(thirdadd, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Cloth Cap")
	maingot := plug.Bids[id].FindBid("Mortimus")
	plug.Bids[id].Bidders[maingot].Player.DKP = 2000
	plug.Bids[id].Bidders[maingot].Player.DKPRank = MAIN
	secondgot := plug.Bids[id].FindBid("Milliardo")
	plug.Bids[id].Bidders[secondgot].Player.DKP = 2000
	plug.Bids[id].Bidders[secondgot].Player.DKPRank = MAIN
	thirdgot := plug.Bids[id].FindBid("Penelo")
	plug.Bids[id].Bidders[thirdgot].Player.DKP = 2000
	plug.Bids[id].Bidders[thirdgot].Player.DKPRank = MAIN
	plug.Bids[id].ApplyDKP()
	plug.Bids[id].SortBids()
	ties := plug.Bids[id].CheckTiesAndApplyWinners()
	got := len(ties)
	want := 0
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %d, want %d", got, want)
	}
}

func TestTiesDiffRank(t *testing.T) {
	configuration.Bids.SecondMainsBidAsMains = true
	configuration.Bids.SecondMainAsMainMaxBid = 200
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Tt][Ee][Ll][Ll][Ss])?([Bb][Ii][Dd][Ss])?\sto\s.+,?.+(\d+).*`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+)\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Cloth Cap 100"
	add.Source = "Mortimus"
	add.T = time.Now()
	plug.Handle(add, &b)
	secondadd := new(everquest.EqLog)
	secondadd.Channel = "tell"
	secondadd.Msg = "Cloth Cap 100"
	secondadd.Source = "Milliardo"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	thirdadd := new(everquest.EqLog)
	thirdadd.Channel = "tell"
	thirdadd.Msg = "Cloth Cap 100"
	thirdadd.Source = "Penelo"
	thirdadd.T = time.Now()
	plug.Handle(thirdadd, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Cloth Cap")
	maingot := plug.Bids[id].FindBid("Mortimus")
	plug.Bids[id].Bidders[maingot].Player.DKP = 2000
	plug.Bids[id].Bidders[maingot].Player.DKPRank = MAIN
	secondgot := plug.Bids[id].FindBid("Milliardo")
	plug.Bids[id].Bidders[secondgot].Player.DKP = 2000
	plug.Bids[id].Bidders[secondgot].Player.DKPRank = ALT
	thirdgot := plug.Bids[id].FindBid("Penelo")
	plug.Bids[id].Bidders[thirdgot].Player.DKP = 2000
	plug.Bids[id].Bidders[thirdgot].Player.DKPRank = MAIN
	plug.Bids[id].ApplyDKP()
	plug.Bids[id].SortBids()
	ties := plug.Bids[id].CheckTiesAndApplyWinners()
	got := len(ties)
	want := 2
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %d, want %d", got, want)
	}
}

func TestTiesDiffRankCancelledBidSecondMain(t *testing.T) {
	configuration.Bids.SecondMainsBidAsMains = true
	configuration.Bids.SecondMainAsMainMaxBid = 200
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Tt][Ee][Ll][Ll][Ss])?([Bb][Ii][Dd][Ss])?\sto\s.+,?.+(\d+).*`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+)\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Cloth Cap 100"
	add.Source = "Mortimus"
	add.T = time.Now()
	plug.Handle(add, &b)
	secondadd := new(everquest.EqLog)
	secondadd.Channel = "tell"
	secondadd.Msg = "Cloth Cap 100"
	secondadd.Source = "Milliardo"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	thirdadd := new(everquest.EqLog)
	thirdadd.Channel = "tell"
	thirdadd.Msg = "Cloth Cap 0"
	thirdadd.Source = "Penelo"
	thirdadd.T = time.Now()
	plug.Handle(thirdadd, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Cloth Cap")
	maingot := plug.Bids[id].FindBid("Mortimus")
	plug.Bids[id].Bidders[maingot].Player.DKP = 2000
	plug.Bids[id].Bidders[maingot].Player.DKPRank = MAIN
	secondgot := plug.Bids[id].FindBid("Milliardo")
	plug.Bids[id].Bidders[secondgot].Player.DKP = 2000
	plug.Bids[id].Bidders[secondgot].Player.DKPRank = SECONDMAIN
	thirdgot := plug.Bids[id].FindBid("Penelo")
	plug.Bids[id].Bidders[thirdgot].Player.DKP = 2000
	plug.Bids[id].Bidders[thirdgot].Player.DKPRank = MAIN
	plug.Bids[id].ApplyDKP()
	plug.Bids[id].SortBids()
	ties := plug.Bids[id].CheckTiesAndApplyWinners()
	got := len(ties)
	want := 2
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %d, want %d", got, want)
	}
}

func TestTiesDiffRankCancelledBid(t *testing.T) {
	configuration.Bids.SecondMainsBidAsMains = true
	configuration.Bids.SecondMainAsMainMaxBid = 200
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Tt][Ee][Ll][Ll][Ss])?([Bb][Ii][Dd][Ss])?\sto\s.+,?.+(\d+).*`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+)\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Cloth Cap 0"
	add.Source = "Mortimus"
	add.T = time.Now()
	plug.Handle(add, &b)
	secondadd := new(everquest.EqLog)
	secondadd.Channel = "tell"
	secondadd.Msg = "Cloth Cap 100"
	secondadd.Source = "Milliardo"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	thirdadd := new(everquest.EqLog)
	thirdadd.Channel = "tell"
	thirdadd.Msg = "Cloth Cap 100"
	thirdadd.Source = "Penelo"
	thirdadd.T = time.Now()
	plug.Handle(thirdadd, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Cloth Cap")
	maingot := plug.Bids[id].FindBid("Mortimus")
	plug.Bids[id].Bidders[maingot].Player.DKP = 2000
	plug.Bids[id].Bidders[maingot].Player.DKPRank = MAIN
	secondgot := plug.Bids[id].FindBid("Milliardo")
	plug.Bids[id].Bidders[secondgot].Player.DKP = 2000
	plug.Bids[id].Bidders[secondgot].Player.DKPRank = ALT
	thirdgot := plug.Bids[id].FindBid("Penelo")
	plug.Bids[id].Bidders[thirdgot].Player.DKP = 2000
	plug.Bids[id].Bidders[thirdgot].Player.DKPRank = ALT
	plug.Bids[id].ApplyDKP()
	plug.Bids[id].SortBids()
	ties := plug.Bids[id].CheckTiesAndApplyWinners()
	got := len(ties)
	want := 2
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %d, want %d", got, want)
	}
}

func TestTiesMoreItemsThanTies(t *testing.T) {
	configuration.Bids.SecondMainsBidAsMains = true
	configuration.Bids.SecondMainAsMainMaxBid = 200
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Capx6 bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Tt][Ee][Ll][Ll][Ss])?([Bb][Ii][Dd][Ss])?\sto\s.+,?.+(\d+).*`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+)\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Cloth Cap 100"
	add.Source = "Mortimus"
	add.T = time.Now()
	plug.Handle(add, &b)
	secondadd := new(everquest.EqLog)
	secondadd.Channel = "tell"
	secondadd.Msg = "Cloth Cap 100"
	secondadd.Source = "Milliardo"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	thirdadd := new(everquest.EqLog)
	thirdadd.Channel = "tell"
	thirdadd.Msg = "Cloth Cap 100"
	thirdadd.Source = "Penelo"
	thirdadd.T = time.Now()
	plug.Handle(thirdadd, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Cloth Cap")
	maingot := plug.Bids[id].FindBid("Mortimus")
	plug.Bids[id].Bidders[maingot].Player.DKP = 2000
	plug.Bids[id].Bidders[maingot].Player.DKPRank = MAIN
	secondgot := plug.Bids[id].FindBid("Milliardo")
	plug.Bids[id].Bidders[secondgot].Player.DKP = 2000
	plug.Bids[id].Bidders[secondgot].Player.DKPRank = MAIN
	thirdgot := plug.Bids[id].FindBid("Penelo")
	plug.Bids[id].Bidders[thirdgot].Player.DKP = 2000
	plug.Bids[id].Bidders[thirdgot].Player.DKPRank = MAIN
	plug.Bids[id].ApplyDKP()
	plug.Bids[id].SortBids()
	ties := plug.Bids[id].CheckTiesAndApplyWinners()
	got := len(ties)
	want := 0
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %d, want %d", got, want)
	}
}

func TestBidApplyNoDKP(t *testing.T) {
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Tt][Ee][Ll][Ll][Ss])?([Bb][Ii][Dd][Ss])?\sto\s.+,?.+(\d+).*`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+)\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Cloth Cap 100"
	add.Source = "Mortimus"
	add.T = time.Now()
	plug.Handle(add, &b)
	id, _ := itemDB.FindIDByName("Cloth Cap")
	got := plug.Bids[id].FindBid("Mortimus")
	plug.Bids[id].Bidders[got].Player.DKP = 0
	plug.Bids[id].ApplyDKP()
	appliedBid := plug.Bids[id].Bidders[got].Bid
	want := 10
	if appliedBid != want {
		t.Errorf("ldplug.Handle(msg, &b) = %d, want %d", appliedBid, want)
	}
}
