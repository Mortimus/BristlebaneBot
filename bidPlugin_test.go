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
	want := "Bids open on Cloth Cap(x1) for 2 minutes.\n"
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got, want)
	}
}

func TestBidClose(t *testing.T) {
	plug := new(BidPlugin)
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
		Tells:    []everquest.EqLog{},
		Bidders:  []*Bidder{},
	}
	var b bytes.Buffer
	plug.Handle(msg, &b)
	got := b.String()
	want := "Closed bids on Cloth Cap(x1)\n"
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
	want := 1
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
