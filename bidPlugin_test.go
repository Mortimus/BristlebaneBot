package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"regexp"
	"testing"
	"time"

	everquest "github.com/Mortimus/goEverquest"
)

func init() {
	configuration.Discord.UseDiscord = false
}

func TestBidOpen(t *testing.T) {
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	id, _ := itemDB.FindIDByName("Cloth Cap")
	// fmt.Printf("ID: %d\n", id)
	got := plug.Bids[id].Quantity
	want := 1
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got, want)
	}
}

func TestBidOpenTellsTo(t *testing.T) {
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap tells to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	id, _ := itemDB.FindIDByName("Cloth Cap")
	// fmt.Printf("ID: %d\n", id)
	got := plug.Bids[id].Quantity
	want := 1
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got, want)
	}
}

func TestBidTimeWithSeconds(t *testing.T) {
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Gloves of the Unseen bids to Mortimus, pst 2min30s"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	id, _ := itemDB.FindIDByName("Gloves of the Unseen")
	// fmt.Printf("ID: %d\n", id)
	got := plug.Bids[id].Duration.Seconds()
	want := 150.0
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %f, want %f", got, want)
	}
}

func TestBidTimeWithoutSeconds(t *testing.T) {
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	id, _ := itemDB.FindIDByName("Cloth Cap")
	// fmt.Printf("ID: %d\n", id)
	got := plug.Bids[id].Duration.Seconds()
	want := 120.0
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %f, want %f", got, want)
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
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
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
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
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
	var bidClosed bool
	if _, ok := plug.Bids[id]; !ok {
		bidClosed = true
	}
	got := bidClosed
	want := true
	if got != want {
		t.Errorf("ldplug.Handle(msg, &t) = %t, want %t", got, want)
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

func TestGuildHotReload(t *testing.T) {
	guild := new(everquest.Guild)
	guild.LoadFromPath(configuration.Everquest.BaseFolder+"/"+"Vets of Norrath_aradune-20210911-205830.txt", Err)
	updateGuildRoster(guild)
	got := getDKPRank(&Roster["Struummin"].GuildMember)
	want := DKPRank(SECONDMAIN)
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", DKPRankToString(got), DKPRankToString(want))
	}
}

func TestGetDKPRankAltSecondMain(t *testing.T) {
	member := &everquest.GuildMember{
		Rank:       "Alt",
		Alt:        true,
		PublicNote: "Ruperts' 2nd main",
	}
	got := getDKPRank(member)
	want := DKPRank(SECONDMAIN)
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", DKPRankToString(got), DKPRankToString(want))
	}
}

func TestGetDKPRankSecondMainHigherRankThanMain(t *testing.T) { // TODO: Fix this with fake members
	member := everquest.GuildMember{
		Name:       "Fakesecond",
		Rank:       "Alt",
		Alt:        true,
		PublicNote: "Fakemain 2nd main",
	}
	memberMain := everquest.GuildMember{
		Name: "Fakemain",
		Rank: "Recruit",
		Alt:  true,
	}
	Roster["Fakemain"] = &DKPHolder{
		GuildMember: memberMain,
		DKPRank:     getDKPRank(&memberMain),
	}
	Roster["Fakesecond"] = &DKPHolder{
		GuildMember: member,
		DKPRank:     getDKPRank(&member),
	}
	fixOutrankingSecondMains()
	// Roster["Blepper"].Rank = "Recruit"
	got := getDKPRank(&Roster["Fakesecond"].GuildMember)
	want := DKPRank(RECRUIT)
	if got != want {
		t.Errorf("Got %q, want %q", DKPRankToString(got), DKPRankToString(want))
	}
	got2 := getDKPRank(&Roster["Fakemain"].GuildMember)
	want2 := DKPRank(RECRUIT)
	if got2 != want2 {
		t.Errorf("Got %q, want %q", DKPRankToString(got2), DKPRankToString(want2))
	}
	got3 := getDKPRank(&Roster["Fakesecond"].GuildMember)
	want3 := getDKPRank(&Roster["Fakemain"].GuildMember)
	if got3 != want3 {
		t.Errorf("Got %q, want %q", DKPRankToString(got3), DKPRankToString(want3))
	}
}

func TestGetDKPRankSecondMainNoApostraphe(t *testing.T) {
	member := &everquest.GuildMember{
		Rank:       "<<< Officer >>>",
		Alt:        true,
		PublicNote: "Mortimus 2nd main",
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
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
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
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
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
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
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
	Roster["Mortimus"].DKP = 2000
	Roster["Mortimus"].DKPRank = MAIN
	// plug.Bids[id].Bidders[got].Player.DKP = 2000
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
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
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
	Roster["Mortimus"].DKP = 2000
	Roster["Mortimus"].DKPRank = MAIN
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
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
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
	Roster["Mortimus"].DKP = 2000
	Roster["Mortimus"].DKPRank = MAIN
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
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
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
	Roster["Mortimus"].DKP = 2000
	Roster["Mortimus"].DKPRank = MAIN
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
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
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
	Roster["Mortimus"].DKP = 2000
	Roster["Mortimus"].DKPRank = MAIN
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
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
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
	// maingot := plug.Bids[id].FindBid("Mortimus")
	Roster["Mortimus"].DKP = 2000
	Roster["Mortimus"].DKPRank = MAIN
	Roster["Milliardo"].DKP = 2000
	Roster["Mortimus"].DKPRank = SECONDMAIN
	secondgot := plug.Bids[id].FindBid("Milliardo")
	plug.Bids[id].ApplyDKP()
	appliedBid := plug.Bids[id].Bidders[secondgot].Bid
	want := configuration.Bids.SecondMainAsMainMaxBid
	if appliedBid != want {
		t.Errorf("ldplug.Handle(msg, &b) = %d, want %d", appliedBid, want)
	}
}

func TestTiesSameRank(t *testing.T) {
	Roster["Mortimus"].DKP = 2000
	Roster["Mortimus"].DKPRank = MAIN
	Roster["Penelo"].DKP = 2000
	Roster["Penelo"].DKPRank = MAIN
	configuration.Bids.SecondMainsBidAsMains = true
	configuration.Bids.SecondMainAsMainMaxBid = 200
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
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
	secondadd.Source = "Penelo"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Cloth Cap")
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
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
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
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
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
	secondadd.Source = "Zortax"
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
	Roster["Mortimus"].DKP = 2000
	Roster["Mortimus"].DKPRank = MAIN
	Roster["Zortax"].DKP = 2000
	Roster["Zortax"].DKPRank = MAIN
	Roster["Penelo"].DKP = 2000
	Roster["Penelo"].DKPRank = MAIN
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
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
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
	secondadd.Source = "Zortax"
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
	Roster["Mortimus"].DKP = 2000
	Roster["Mortimus"].DKPRank = MAIN
	Roster["Zortax"].DKP = 2000
	Roster["Zortax"].DKPRank = MAIN
	Roster["Penelo"].DKP = 2000
	Roster["Penelo"].DKPRank = MAIN
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
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
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
	secondadd.Source = "Rokem"
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
	Roster["Mortimus"].DKP = 2000
	Roster["Mortimus"].DKPRank = MAIN
	Roster["Rokem"].DKP = 2000
	Roster["Rokem"].DKPRank = ALT
	Roster["Penelo"].DKP = 2000
	Roster["Penelo"].DKPRank = MAIN
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
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
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
	Roster["Mortimus"].DKP = 2000
	Roster["Mortimus"].DKPRank = MAIN
	Roster["Milliardo"].DKP = 2000
	Roster["Milliardo"].DKPRank = SECONDMAIN
	Roster["Penelo"].DKP = 2000
	Roster["Penelo"].DKPRank = MAIN
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
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
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
	secondadd.Source = "Rokem"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	thirdadd := new(everquest.EqLog)
	thirdadd.Channel = "tell"
	thirdadd.Msg = "Cloth Cap 100"
	thirdadd.Source = "Glavin"
	thirdadd.T = time.Now()
	plug.Handle(thirdadd, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Cloth Cap")
	Roster["Mortimus"].DKP = 2000
	Roster["Mortimus"].DKPRank = MAIN
	Roster["Rokem"].DKP = 2000
	Roster["Rokem"].DKPRank = ALT
	Roster["Glavin"].DKP = 2000
	Roster["Glavin"].DKPRank = ALT
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
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
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
	secondadd.Source = "Zortax"
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
	Roster["Mortimus"].DKP = 2000
	Roster["Mortimus"].DKPRank = MAIN
	Roster["Zortax"].DKP = 2000
	Roster["Zortax"].DKPRank = MAIN
	Roster["Penelo"].DKP = 2000
	Roster["Penelo"].DKPRank = MAIN
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
	mem := everquest.GuildMember{Name: "NotReal", Class: "Warrior", Level: 1, Rank: "Raider"}
	Roster["NotReal"] = &DKPHolder{DKP: 0, DKPRank: MAIN, GuildMember: mem}
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Cloth Cap 100"
	add.Source = "NotReal"
	add.T = time.Now()
	plug.Handle(add, &b)
	id, _ := itemDB.FindIDByName("Cloth Cap")
	got := plug.Bids[id].FindBid("NotReal")

	plug.Bids[id].ApplyDKP()
	appliedBid := plug.Bids[id].Bidders[got].Bid
	want := 10
	if appliedBid != want {
		t.Errorf("ldplug.Handle(msg, &b) = %d, want %d", appliedBid, want)
	}
}

func TestBidGitHubIssue34(t *testing.T) {
	Roster["Rabidtiger"].DKP = 2000
	Roster["Rabidtiger"].DKPRank = MAIN
	Roster["Yilumi"].DKP = 2000
	Roster["Yilumi"].DKPRank = MAIN
	Roster["Nistalkin"].DKP = 2000
	Roster["Nistalkin"].DKPRank = MAIN
	Roster["Boseth"].DKP = 2000
	Roster["Boseth"].DKPRank = MAIN
	Roster["Bremen"].DKP = 2000
	Roster["Bremen"].DKPRank = MAIN
	configuration.Bids.SecondMainsBidAsMains = true
	configuration.Bids.SecondMainAsMainMaxBid = 200
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Capx2 bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Cloth Cap 75"
	add.Source = "Rabidtiger"
	add.T = time.Now()
	plug.Handle(add, &b)
	secondadd := new(everquest.EqLog)
	secondadd.Channel = "tell"
	secondadd.Msg = "Cloth Cap 20"
	secondadd.Source = "Yilumi"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	thirdadd := new(everquest.EqLog)
	thirdadd.Channel = "tell"
	thirdadd.Msg = "Cloth Cap 20"
	thirdadd.Source = "Nistalkin"
	thirdadd.T = time.Now()
	plug.Handle(thirdadd, &b)
	fourthadd := new(everquest.EqLog)
	fourthadd.Channel = "tell"
	fourthadd.Msg = "Cloth Cap 20"
	fourthadd.Source = "Boseth"
	fourthadd.T = time.Now()
	plug.Handle(fourthadd, &b)
	fifthadd := new(everquest.EqLog)
	fifthadd.Channel = "tell"
	fifthadd.Msg = "Cloth Cap 20"
	fifthadd.Source = "Bremen"
	fifthadd.T = time.Now()
	plug.Handle(fifthadd, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Cloth Cap")
	plug.Bids[id].CloseBids(io.Discard)
	got := plug.Bids[id].WinningBid
	want := 20
	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestBidWinningPlusFive(t *testing.T) {
	Roster["Rabidtiger"].DKP = 2000
	Roster["Rabidtiger"].DKPRank = MAIN
	Roster["Yilumi"].DKP = 2000
	Roster["Yilumi"].DKPRank = MAIN
	Roster["Nistalkin"].DKP = 2000
	Roster["Nistalkin"].DKPRank = MAIN
	Roster["Boseth"].DKP = 2000
	Roster["Boseth"].DKPRank = MAIN
	Roster["Bremen"].DKP = 2000
	Roster["Bremen"].DKPRank = MAIN
	configuration.Bids.SecondMainsBidAsMains = true
	configuration.Bids.SecondMainAsMainMaxBid = 200
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Cloth Cap 75"
	add.Source = "Rabidtiger"
	add.T = time.Now()
	plug.Handle(add, &b)
	secondadd := new(everquest.EqLog)
	secondadd.Channel = "tell"
	secondadd.Msg = "Cloth Cap 65"
	secondadd.Source = "Yilumi"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	thirdadd := new(everquest.EqLog)
	thirdadd.Channel = "tell"
	thirdadd.Msg = "Cloth Cap 60"
	thirdadd.Source = "Nistalkin"
	thirdadd.T = time.Now()
	plug.Handle(thirdadd, &b)
	fourthadd := new(everquest.EqLog)
	fourthadd.Channel = "tell"
	fourthadd.Msg = "Cloth Cap 20"
	fourthadd.Source = "Boseth"
	fourthadd.T = time.Now()
	plug.Handle(fourthadd, &b)
	fifthadd := new(everquest.EqLog)
	fifthadd.Channel = "tell"
	fifthadd.Msg = "Cloth Cap 10"
	fifthadd.Source = "Bremen"
	fifthadd.T = time.Now()
	plug.Handle(fifthadd, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Cloth Cap")
	// plug.Bids[id].ApplyDKP()
	// plug.Bids[id].SortBids()
	plug.Bids[id].CloseBids(io.Discard)
	got := plug.Bids[id].WinningBid
	want := 70
	if got != want {
		t.Errorf("got %d, want %d", got, want)
	}
}

func TestTiesSameRankTiedWinningBid(t *testing.T) {
	Roster["Mortimus"].DKP = 2000
	Roster["Mortimus"].DKPRank = MAIN
	Roster["Penelo"].DKP = 2000
	Roster["Penelo"].DKPRank = MAIN
	configuration.Bids.SecondMainsBidAsMains = true
	configuration.Bids.SecondMainAsMainMaxBid = 200
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
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
	secondadd.Source = "Penelo"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Cloth Cap")
	// plug.Bids[id].ApplyDKP()
	// plug.Bids[id].SortBids()
	plug.Bids[id].CloseBids(io.Discard)
	got := plug.Bids[id].WinningBid
	want := 300
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %d, want %d", got, want)
	}
}

func TestBidGitHubIssue40(t *testing.T) {
	Roster["Greyvvolf"].DKP = 2000
	Roster["Greyvvolf"].DKPRank = RECRUIT
	Roster["Canniblepper"].DKP = 2000
	Roster["Canniblepper"].DKPRank = SECONDMAIN
	configuration.Bids.SecondMainsBidAsMains = true
	configuration.Bids.SecondMainAsMainMaxBid = 200
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Cloth Cap 10"
	add.Source = "Greyvvolf"
	add.T = time.Now()
	plug.Handle(add, &b)
	secondadd := new(everquest.EqLog)
	secondadd.Channel = "tell"
	secondadd.Msg = "Cloth Cap 125"
	secondadd.Source = "Canniblepper"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Cloth Cap")
	// plug.Bids[id].ApplyDKP()
	// plug.Bids[id].SortBids()
	plug.Bids[id].CloseBids(io.Discard)
	got := plug.Bids[id].WinningBid
	want := 10
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %d, want %d", got, want)
	}
	got2 := plug.Bids[id].Bidders[0].Player.Name
	want2 := "Canniblepper"
	if got2 != want2 {
		t.Errorf("ldplug.Handle(msg, &b) = %s, want %s", got2, want2)
	}
}

func TestBidSingleMinBidWinner(t *testing.T) {
	Roster["Guzz"].DKP = 2000
	Roster["Guzz"].DKPRank = RECRUIT
	Roster["Flappyhands"].DKP = 2000
	Roster["Flappyhands"].DKPRank = SECONDMAIN
	Roster["Boogabooga"].DKP = 2000
	Roster["Boogabooga"].DKPRank = MAIN
	configuration.Bids.SecondMainsBidAsMains = true
	configuration.Bids.SecondMainAsMainMaxBid = 200
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Cloth Cap 700"
	add.Source = "Guzz"
	add.T = time.Now()
	plug.Handle(add, &b)
	secondadd := new(everquest.EqLog)
	secondadd.Channel = "tell"
	secondadd.Msg = "Cloth Cap 60"
	secondadd.Source = "Flappyhands"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	thirdadd := new(everquest.EqLog)
	thirdadd.Channel = "tell"
	thirdadd.Msg = "Cloth Cap 0"
	thirdadd.Source = "Boogabooga"
	thirdadd.T = time.Now()
	plug.Handle(thirdadd, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Cloth Cap")
	// plug.Bids[id].ApplyDKP()
	// plug.Bids[id].SortBids()
	// fmt.Printf("GuzzRank: %d FlappyRank: %d\n", GetEffectiveDKPRank(Roster["Guzz"].DKPRank), GetEffectiveDKPRank(Roster["Flappyhands"].DKPRank))
	plug.Bids[id].CloseBids(io.Discard)
	// fmt.Printf("GuzzRank: %d FlappyRank: %d\n", GetEffectiveDKPRank(Roster["Guzz"].DKPRank), GetEffectiveDKPRank(Roster["Flappyhands"].DKPRank))
	got := plug.Bids[id].WinningBid
	want := 10
	if got != want {
		t.Errorf("WinningBid %d, want %d", got, want)
	}
	got2 := plug.Bids[id].Bidders[0].Player.Name
	want2 := "Flappyhands"
	if got2 != want2 {
		t.Errorf("Winner %s, want %s", got2, want2)
	}
}

func TestBidMultiMinBidWinner(t *testing.T) {
	Roster["Guzz"].DKP = 2000
	Roster["Guzz"].DKPRank = RECRUIT
	Roster["Flappyhands"].DKP = 2000
	Roster["Flappyhands"].DKPRank = SECONDMAIN
	Roster["Boogabooga"].DKP = 2000
	Roster["Boogabooga"].DKPRank = MAIN
	configuration.Bids.MinimumBid = 10
	configuration.Bids.SecondMainsBidAsMains = true
	configuration.Bids.SecondMainAsMainMaxBid = 200
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Greaves of Furious Mightx2 bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Greaves of Furious Might 700"
	add.Source = "Guzz"
	add.T = time.Now()
	plug.Handle(add, &b)
	secondadd := new(everquest.EqLog)
	secondadd.Channel = "tell"
	secondadd.Msg = "Greaves of Furious Might 60 "
	secondadd.Source = "Flappyhands"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	thirdadd := new(everquest.EqLog)
	thirdadd.Channel = "tell"
	thirdadd.Msg = "Greaves of Furious Might 0"
	thirdadd.Source = "Boogabooga"
	thirdadd.T = time.Now()
	plug.Handle(thirdadd, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Greaves of Furious Might")
	// plug.Bids[id].ApplyDKP()
	// plug.Bids[id].SortBids()
	plug.Bids[id].CloseBids(io.Discard)
	got := plug.Bids[id].WinningBid
	want := 10
	if got != want {
		t.Errorf("Got %d, want %d", got, want)
	}
	winners := plug.Bids[id].GetWinnerNames()
	got2 := winners[0]
	want2 := "Flappyhands"

	if got2 != want2 {
		t.Errorf("1st place = %s, want %s", got2, want2)
	}
	got3 := winners[1]
	want3 := "Guzz"
	if got3 != want3 {
		t.Errorf("2nd place = %s, want %s", got3, want3)
	}
}

func TestBidSingleMinBidWinner2(t *testing.T) {
	Roster["Glooping"].DKP = 2000
	Roster["Glooping"].DKPRank = MAIN
	Roster["Yilumi"].DKP = 2000
	Roster["Yilumi"].DKPRank = MAIN
	configuration.Bids.SecondMainsBidAsMains = true
	configuration.Bids.SecondMainAsMainMaxBid = 200
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Cap bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Cloth Cap 325"
	add.Source = "Glooping"
	add.T = time.Now()
	plug.Handle(add, &b)
	secondadd := new(everquest.EqLog)
	secondadd.Channel = "tell"
	secondadd.Msg = "Cloth Cap 0"
	secondadd.Source = "Yilumi"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Cloth Cap")
	// plug.Bids[id].ApplyDKP()
	// plug.Bids[id].SortBids()
	// fmt.Printf("GuzzRank: %d FlappyRank: %d\n", GetEffectiveDKPRank(Roster["Guzz"].DKPRank), GetEffectiveDKPRank(Roster["Flappyhands"].DKPRank))
	plug.Bids[id].CloseBids(io.Discard)
	// fmt.Printf("GuzzRank: %d FlappyRank: %d\n", GetEffectiveDKPRank(Roster["Guzz"].DKPRank), GetEffectiveDKPRank(Roster["Flappyhands"].DKPRank))
	got := plug.Bids[id].WinningBid
	want := 10
	if got != want {
		t.Errorf("WinningBid %d, want %d", got, want)
	}
	got2 := plug.Bids[id].Bidders[0].Player.Name
	want2 := "Glooping"
	if got2 != want2 {
		t.Errorf("Winner %s, want %s", got2, want2)
	}
}

func TestBidMultiDiffBids(t *testing.T) {
	Roster["Drae"].DKP = 2000
	Roster["Drae"].DKPRank = MAIN
	Roster["Penelo"].DKP = 2000
	Roster["Penelo"].DKPRank = MAIN
	configuration.Bids.SecondMainsBidAsMains = true
	configuration.Bids.SecondMainAsMainMaxBid = 200
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Cloth Capx2 bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Cloth Cap 200"
	add.Source = "Drae"
	add.T = time.Now()
	plug.Handle(add, &b)
	secondadd := new(everquest.EqLog)
	secondadd.Channel = "tell"
	secondadd.Msg = "Cloth Cap 40"
	secondadd.Source = "Penelo"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Cloth Cap")
	// plug.Bids[id].ApplyDKP()
	// plug.Bids[id].SortBids()
	// fmt.Printf("GuzzRank: %d FlappyRank: %d\n", GetEffectiveDKPRank(Roster["Guzz"].DKPRank), GetEffectiveDKPRank(Roster["Flappyhands"].DKPRank))
	plug.Bids[id].CloseBids(io.Discard)
	// fmt.Printf("GuzzRank: %d FlappyRank: %d\n", GetEffectiveDKPRank(Roster["Guzz"].DKPRank), GetEffectiveDKPRank(Roster["Flappyhands"].DKPRank))
	got := plug.Bids[id].WinningBid
	want := 10
	if got != want {
		t.Errorf("WinningBid %d, want %d", got, want)
	}
	got2 := plug.Bids[id].Bidders[0].Player.Name
	want2 := "Drae"
	if got2 != want2 {
		t.Errorf("Winner %s, want %s", got2, want2)
	}
}

func TestBidTripleBidWinner(t *testing.T) {
	Roster["Voltha"].DKP = 2000
	Roster["Voltha"].DKPRank = MAIN
	Roster["Mayfair"].DKP = 2000
	Roster["Mayfair"].DKPRank = MAIN
	Roster["Boogabooga"].DKP = 2000
	Roster["Boogabooga"].DKPRank = MAIN
	Roster["Guzz"].DKP = 2000
	Roster["Guzz"].DKPRank = MAIN
	Roster["Sitoknight"].DKP = 2000
	Roster["Sitoknight"].DKPRank = RECRUIT
	configuration.Bids.MinimumBid = 10
	configuration.Bids.SecondMainsBidAsMains = true
	configuration.Bids.SecondMainAsMainMaxBid = 200
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Mossy Enchanted Stonex2 bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Mossy Enchanted Stone 200"
	add.Source = "Voltha"
	add.T = time.Now()
	plug.Handle(add, &b)
	secondadd := new(everquest.EqLog)
	secondadd.Channel = "tell"
	secondadd.Msg = "Mossy Enchanted Stone 110"
	secondadd.Source = "Mayfair"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	thirdadd := new(everquest.EqLog)
	thirdadd.Channel = "tell"
	thirdadd.Msg = "Mossy Enchanted Stone 45"
	thirdadd.Source = "Boogabooga"
	thirdadd.T = time.Now()
	plug.Handle(thirdadd, &b)
	fouradd := new(everquest.EqLog)
	fouradd.Channel = "tell"
	fouradd.Msg = "Mossy Enchanted Stone 15"
	fouradd.Source = "Guzz"
	fouradd.T = time.Now()
	plug.Handle(fouradd, &b)
	fiveadd := new(everquest.EqLog)
	fiveadd.Channel = "tell"
	fiveadd.Msg = "Mossy Enchanted Stone 75"
	fiveadd.Source = "Sitoknight"
	fiveadd.T = time.Now()
	plug.Handle(fiveadd, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Mossy Enchanted Stone")
	// plug.Bids[id].ApplyDKP()
	// plug.Bids[id].SortBids()
	plug.Bids[id].CloseBids(io.Discard)
	got := plug.Bids[id].WinningBid
	want := 50
	if got != want {
		t.Errorf("Got %d, want %d", got, want)
	}
	got2 := plug.Bids[id].Bidders[0].Player.Name
	want2 := "Voltha"
	if got2 != want2 {
		t.Errorf("ldplug.Handle(msg, &b) = %s, want %s", got2, want2)
	}
}

func TestBidMultiItemBidIssue45(t *testing.T) {
	updateDKP = false
	Roster["Blepper"].DKP = 2000
	Roster["Blepper"].DKPRank = MAIN
	Roster["Renab"].DKP = 2000
	Roster["Renab"].DKPRank = MAIN
	Roster["Yilumi"].DKP = 2000
	Roster["Yilumi"].DKPRank = MAIN
	Roster["Mortimus"].DKP = 2000
	Roster["Mortimus"].DKPRank = MAIN
	Roster["Ravnor"].DKP = 2000
	Roster["Ravnor"].DKPRank = MAIN
	Roster["Bipp"].DKP = 2000
	Roster["Bipp"].DKPRank = MAIN
	Roster["Yzzy"].DKP = 2000
	Roster["Yzzy"].DKPRank = MAIN
	Roster["Glert"].DKP = 2000
	Roster["Glert"].DKPRank = MAIN
	Roster["Raage"].DKP = 2000
	Roster["Raage"].DKPRank = MAIN
	Roster["Ryder"].DKP = 2000
	Roster["Ryder"].DKPRank = MAIN
	Roster["Gausbert"].DKP = 2000
	Roster["Gausbert"].DKPRank = MAIN
	Roster["Liqqy"].DKP = 2000
	Roster["Liqqy"].DKPRank = RECRUIT
	configuration.Bids.MinimumBid = 10
	configuration.Bids.SecondMainsBidAsMains = true
	configuration.Bids.SecondMainAsMainMaxBid = 200
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Mossy Enchanted Stonex2 bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`(.+[\w\d])\s+(\d+).*`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "Mossy Enchanted Stone 450"
	add.Source = "Blepper"
	add.T = time.Now()
	plug.Handle(add, &b)
	secondadd := new(everquest.EqLog)
	secondadd.Channel = "tell"
	secondadd.Msg = "Mossy Enchanted Stone 25"
	secondadd.Source = "Renab"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	thirdadd := new(everquest.EqLog)
	thirdadd.Channel = "tell"
	thirdadd.Msg = "Mossy Enchanted Stone 305"
	thirdadd.Source = "Yilumi"
	thirdadd.T = time.Now()
	plug.Handle(thirdadd, &b)
	fouradd := new(everquest.EqLog)
	fouradd.Channel = "tell"
	fouradd.Msg = "Mossy Enchanted Stone 300"
	fouradd.Source = "Mortimus"
	fouradd.T = time.Now()
	plug.Handle(fouradd, &b)
	fiveadd := new(everquest.EqLog)
	fiveadd.Channel = "tell"
	fiveadd.Msg = "Mossy Enchanted Stone 210"
	fiveadd.Source = "Ravnor"
	fiveadd.T = time.Now()
	plug.Handle(fiveadd, &b)
	sixadd := new(everquest.EqLog)
	sixadd.Channel = "tell"
	sixadd.Msg = "Mossy Enchanted Stone 150"
	sixadd.Source = "Bipp"
	sixadd.T = time.Now()
	plug.Handle(sixadd, &b)
	sevnadd := new(everquest.EqLog)
	sevnadd.Channel = "tell"
	sevnadd.Msg = "Mossy Enchanted Stone 110"
	sevnadd.Source = "Yzzy"
	sevnadd.T = time.Now()
	plug.Handle(sevnadd, &b)
	eightadd := new(everquest.EqLog)
	eightadd.Channel = "tell"
	eightadd.Msg = "Mossy Enchanted Stone 25"
	eightadd.Source = "Glert"
	eightadd.T = time.Now()
	plug.Handle(eightadd, &b)
	nineadd := new(everquest.EqLog)
	nineadd.Channel = "tell"
	nineadd.Msg = "Mossy Enchanted Stone 20"
	nineadd.Source = "Raage"
	nineadd.T = time.Now()
	plug.Handle(nineadd, &b)
	tenadd := new(everquest.EqLog)
	tenadd.Channel = "tell"
	tenadd.Msg = "Mossy Enchanted Stone 15"
	tenadd.Source = "Ryder"
	tenadd.T = time.Now()
	plug.Handle(tenadd, &b)
	eleadd := new(everquest.EqLog)
	eleadd.Channel = "tell"
	eleadd.Msg = "Mossy Enchanted Stone 10"
	eleadd.Source = "Gausbert"
	eleadd.T = time.Now()
	plug.Handle(eleadd, &b)
	twelveadd := new(everquest.EqLog)
	twelveadd.Channel = "tell"
	twelveadd.Msg = "Mossy Enchanted Stone 400"
	twelveadd.Source = "Liqqy"
	twelveadd.T = time.Now()
	plug.Handle(twelveadd, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Mossy Enchanted Stone")
	// plug.Bids[id].ApplyDKP()
	// plug.Bids[id].SortBids()
	plug.Bids[id].CloseBids(io.Discard)
	// plug.Bids[id].printBidders()
	got := plug.Bids[id].WinningBid
	want := 305
	if got != want {
		t.Errorf("Got %d, want %d", got, want)
	}
	got2 := plug.Bids[id].Bidders[0].Player.Name
	want2 := "Blepper"
	if got2 != want2 {
		t.Errorf("ldplug.Handle(msg, &b) = %s, want %s", got2, want2)
	}
	got3 := plug.Bids[id].Bidders[1].Player.Name
	want3 := "Yilumi"
	if got3 != want3 {
		t.Errorf("ldplug.Handle(msg, &b) = %s, want %s", got3, want3)
	}
}

func TestCanEquipNone(t *testing.T) { // TODO: Fix this with fake members
	war := everquest.GuildMember{
		Name:  "MrWarrior",
		Class: "Warrior",
	}
	nec := everquest.GuildMember{
		Name:  "MrNecro",
		Class: "Necromancer",
	}
	id, _ := itemDB.FindIDByName("Kreljnok's Sword of Eternal Power")
	item, _ := itemDB.GetItemByID(id)
	got := canEquip(item, war)
	want := true
	if got != want {
		t.Errorf("%s canEquip %s: %t, want %t", war.Name, item.Name, got, want)
	}
	// id2, _ := itemDB.FindIDByName("Mossy Enchanted Stone")
	// item2, _ := itemDB.GetItemByID(id2)
	got2 := canEquip(item, nec)
	want2 := false
	if got2 != want2 {
		t.Errorf("%s canEquip %s: %t, want %t", nec.Name, item.Name, got2, want2)
	}
	id3, _ := itemDB.FindIDByName("Shard of Dark Matter")
	item3, _ := itemDB.GetItemByID(id3)
	got3 := canEquip(item3, nec)
	want3 := true
	if got3 != want3 {
		t.Errorf("%s canEquip %s: %t, want %t", nec.Name, item3.Name, got3, want3)
	}
}

func TestMissingBidsIssue47(t *testing.T) {
	updateDKP = false
	Roster["Draeadin"].DKP = 1420
	Roster["Draeadin"].DKPRank = MAIN
	Roster["Bremen"].DKP = 540
	Roster["Bremen"].DKPRank = MAIN
	Roster["Zortax"].DKP = 2695
	Roster["Zortax"].DKPRank = MAIN
	Roster["Raage"].DKP = 1470
	Roster["Raage"].DKPRank = MAIN
	configuration.Bids.MinimumBid = 10
	configuration.Bids.SecondMainsBidAsMains = true
	configuration.Bids.SecondMainAsMainMaxBid = 200
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Bulwark of Living Stone bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`'(.+[\w\d])\s+(\d+).*'`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	secondadd := new(everquest.EqLog)
	secondadd.Channel = "tell"
	secondadd.Msg = "'Bulwark of Living Stone 999999999999999999999999999'"
	secondadd.Source = "Bremen"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	thirdadd := new(everquest.EqLog)
	thirdadd.Channel = "tell"
	thirdadd.Msg = "'Bulwark of Living Stone  800 '"
	thirdadd.Source = "Zortax"
	thirdadd.T = time.Now()
	plug.Handle(thirdadd, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "'Bulwark of Living Stone 600'"
	add.Source = "Draeadin"
	add.T = time.Now()
	plug.Handle(add, &b)
	fouradd := new(everquest.EqLog)
	fouradd.Channel = "tell"
	fouradd.Msg = "'Bulwark of Living Stone  805'"
	fouradd.Source = "Raage"
	fouradd.T = time.Now()
	plug.Handle(fouradd, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Bulwark of Living Stone")
	// plug.Bids[id].ApplyDKP()
	// plug.Bids[id].SortBids()
	plug.Bids[id].CloseBids(io.Discard)
	// plug.Bids[id].printBidders()
	got := plug.Bids[id].WinningBid
	want := 805
	if got != want {
		t.Errorf("Got %d, want %d", got, want)
	}
	got2 := plug.Bids[id].Bidders[0].Player.Name
	want2 := "Raage"
	if got2 != want2 {
		t.Errorf("ldplug.Handle(msg, &b) = %s, want %s", got2, want2)
	}
}

func TestRoundDownIssue46(t *testing.T) {
	updateDKP = false
	Roster["Draeadin"].DKP = 1420
	Roster["Draeadin"].DKPRank = MAIN
	Roster["Bremen"].DKP = 540
	Roster["Bremen"].DKPRank = MAIN
	configuration.Bids.MinimumBid = 10
	configuration.Bids.SecondMainsBidAsMains = true
	configuration.Bids.SecondMainAsMainMaxBid = 200
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Bulwark of Living Stone bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`'(.+[\w\d])\s+(\d+).*'`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	secondadd := new(everquest.EqLog)
	secondadd.Channel = "tell"
	secondadd.Msg = "'Bulwark of Living Stone 999999999999999999999999999'"
	secondadd.Source = "Bremen"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "'Bulwark of Living Stone 200'"
	add.Source = "Draeadin"
	add.T = time.Now()
	plug.Handle(add, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Bulwark of Living Stone")
	// plug.Bids[id].ApplyDKP()
	// plug.Bids[id].SortBids()
	plug.Bids[id].CloseBids(io.Discard)
	// plug.Bids[id].printBidders()
	got := plug.Bids[id].WinningBid
	want := 205
	if got != want {
		t.Errorf("Got %d, want %d", got, want)
	}
	got2 := plug.Bids[id].Bidders[0].Player.Name
	want2 := "Bremen"
	if got2 != want2 {
		t.Errorf("ldplug.Handle(msg, &b) = %s, want %s", got2, want2)
	}
}

func TestItemBidNoSpace(t *testing.T) {
	updateDKP = false
	Roster["Draeadin"].DKP = 1420
	Roster["Draeadin"].DKPRank = MAIN
	Roster["Bremen"].DKP = 540
	Roster["Bremen"].DKPRank = MAIN
	configuration.Bids.MinimumBid = 10
	configuration.Bids.SecondMainsBidAsMains = true
	configuration.Bids.SecondMainAsMainMaxBid = 200
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Bulwark of Living Stone bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`'(.+[\w\d])\s+(\d+).*'`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	secondadd := new(everquest.EqLog)
	secondadd.Channel = "tell"
	secondadd.Msg = "'Bulwark of Living Stone 999999999999999999999999999'"
	secondadd.Source = "Bremen"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "'Bulwark of Living Stone200'"
	add.Source = "Draeadin"
	add.T = time.Now()
	plug.Handle(add, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Bulwark of Living Stone")
	// plug.Bids[id].ApplyDKP()
	// plug.Bids[id].SortBids()
	plug.Bids[id].CloseBids(io.Discard)
	// plug.Bids[id].printBidders()
	got := plug.Bids[id].WinningBid
	want := 205
	if got != want {
		t.Errorf("Got %d, want %d", got, want)
	}
	got2 := plug.Bids[id].Bidders[0].Player.Name
	want2 := "Bremen"
	if got2 != want2 {
		t.Errorf("ldplug.Handle(msg, &b) = %s, want %s", got2, want2)
	}
}

func TestRoundDownIssue51(t *testing.T) {
	updateDKP = false
	// for k, _ := range Roster {
	// 	fmt.Printf("%s\n", k)
	// }
	Roster["Silvae"].DKP = 170
	Roster["Silvae"].DKPRank = MAIN
	Roster["Geban"].DKP = 855
	Roster["Geban"].DKPRank = MAIN
	Roster["Karalaine"].DKP = 1800
	Roster["Karalaine"].DKPRank = MAIN
	configuration.Bids.MinimumBid = 10
	configuration.Bids.SecondMainsBidAsMains = true
	configuration.Bids.SecondMainAsMainMaxBid = 200
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "Bulwark of Living Stone bids to Bids, pst 2min"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.BidAddMatch, _ = regexp.Compile(`'(.+[\w\d])\s+(\d+).*'`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	secondadd := new(everquest.EqLog)
	secondadd.Channel = "tell"
	secondadd.Msg = "'Bulwark of Living Stone 999'"
	secondadd.Source = "Silvae"
	secondadd.T = time.Now()
	plug.Handle(secondadd, &b)
	add := new(everquest.EqLog)
	add.Channel = "tell"
	add.Msg = "'Bulwark of Living Stone 855'"
	add.Source = "Geban"
	add.T = time.Now()
	plug.Handle(add, &b)
	thirdadd := new(everquest.EqLog)
	thirdadd.Channel = "tell"
	thirdadd.Msg = "'Bulwark of Living Stone 800'"
	thirdadd.Source = "Karalaine"
	thirdadd.T = time.Now()
	plug.Handle(thirdadd, &b)
	//----------------
	id, _ := itemDB.FindIDByName("Bulwark of Living Stone")
	// plug.Bids[id].ApplyDKP()
	// plug.Bids[id].SortBids()
	plug.Bids[id].CloseBids(io.Discard)
	// plug.Bids[id].printBidders()
	got := plug.Bids[id].WinningBid
	want := 805
	if got != want {
		t.Errorf("Got %d, want %d", got, want)
	}
	got2 := plug.Bids[id].Bidders[0].Player.Name
	want2 := "Geban"
	if got2 != want2 {
		t.Errorf("ldplug.Handle(msg, &b) = %s, want %s", got2, want2)
	}
}

func TestBidMultiOpen(t *testing.T) {
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "You say to your guild, 'Scales of the Cragbeast Queen | Phosphorescent Bile | Misshapen Cragbeast Flesh bids to Mortimus, pst 2min'"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	id1, _ := itemDB.FindIDByName("Scales of the Cragbeast Queen")
	// fmt.Printf("ID: %d\n", id)
	got := plug.Bids[id1].Quantity
	want := 1
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got, want)
	}
	id2, _ := itemDB.FindIDByName("Phosphorescent Bile")
	// fmt.Printf("ID: %d\n", id)
	got2 := plug.Bids[id2].Quantity
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got2, want)
	}
	id3, _ := itemDB.FindIDByName("Misshapen Cragbeast Flesh")
	// fmt.Printf("ID: %d\n", id)
	got3 := plug.Bids[id3].Quantity
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got3, want)
	}
}

func TestBidMultiOpenMultiQuantity(t *testing.T) {
	plug := new(BidPlugin)
	msg := new(everquest.EqLog)
	msg.Channel = "guild"
	msg.Msg = "You say to your guild, 'Scales of the Cragbeast Queen | Scales of the Cragbeast Queen | Phosphorescent Bile | Cloth Cap bids to Mortimus, pst 2min'"
	msg.Source = "You"
	msg.T = time.Now()
	plug.Bids = make(map[int]*OpenBid)
	plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(`(.+?)(x\d)?\s+([Bb][Ii][Dd][Ss])?([Tt][Ee][Ll][Ll][Ss])?\sto\s.+,?.+([Cc][Ll][Oo][Ss][Ee][Dd]).*`)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	var b bytes.Buffer
	plug.Handle(msg, &b)
	id1, _ := itemDB.FindIDByName("Scales of the Cragbeast Queen")
	// fmt.Printf("ID: %d\n", id)
	got := plug.Bids[id1].Quantity
	want := 2
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got, want)
	}
	id2, _ := itemDB.FindIDByName("Phosphorescent Bile")
	// fmt.Printf("ID: %d\n", id)
	got2 := plug.Bids[id2].Quantity
	want2 := 1
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got2, want2)
	}
	id3, _ := itemDB.FindIDByName("Cloth Cap")
	// fmt.Printf("ID: %d\n", id)
	got3 := plug.Bids[id3].Quantity
	if got != want {
		t.Errorf("ldplug.Handle(msg, &b) = %q, want %q", got3, want2)
	}
}
