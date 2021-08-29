package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	everquest "github.com/Mortimus/goEverquest"
)

// type BidPlugin Plugin
type BidPlugin struct {
	Plugin
	BidOpenMatch  *regexp.Regexp
	BidCloseMatch *regexp.Regexp
	BidAddMatch   *regexp.Regexp
	Bids          map[int]*OpenBid
}

type OpenBid struct {
	Item                 everquest.Item
	Quantity             int
	Duration             time.Duration
	Start                time.Time
	End                  time.Time
	Tells                []everquest.EqLog
	Bidders              []*Bidder
	Zone                 string
	SecondMainBidsAsMain bool
	SecondMainMaxBid     int
}

type Bidder struct {
	Player       *DKPHolder
	AttemptedBid int
	Bid          int
	Message      everquest.EqLog
}

// DKP Ranks
type DKPRank int

const (
	INACTIVE = iota
	SOCIAL
	ALT
	RECRUIT
	SECONDMAIN
	MAIN
)

type DKPHolder struct {
	everquest.GuildMember
	DKP     int
	DKPRank DKPRank
	Thirty  float64
	Sixty   float64
	Ninety  float64
	AllTime float64
}

var Roster map[string]*DKPHolder

func init() {
	plug := new(BidPlugin)
	plug.Name = "Bidding detection"
	plug.Author = "Mortimus"
	plug.Version = "1.0.0"
	plug.Output = BIDOUT
	plug.BidOpenMatch, _ = regexp.Compile(configuration.Bids.RegexOpenBid)
	plug.BidCloseMatch, _ = regexp.Compile(configuration.Bids.RegexClosedBid)
	plug.BidAddMatch, _ = regexp.Compile(configuration.Bids.RegexTellBid)
	plug.Bids = make(map[int]*OpenBid)
	Roster = make(map[string]*DKPHolder)
	path, err := everquest.GetRecentRosterDump(configuration.Everquest.BaseFolder, configuration.Everquest.GuildName)
	if err != nil {
		fmt.Printf("Error finding roster dump: %s", err.Error())
	} else {
		// var guild *everquest.Guild
		guild := new(everquest.Guild)
		fileLog := log.New(os.Stdout, "[WARN] ", log.Lshortfile|log.Ldate|log.Ltime|log.LUTC|log.Lmsgprefix)
		fullpath := configuration.Everquest.BaseFolder + "/" + path
		// fmt.Printf("FullPath: %s\n", fullpath)
		err := guild.LoadFromPath(fullpath, fileLog)
		if err != nil {
			fmt.Printf("Error loading roster dump: %s", err.Error())
		} else {
			loadGuildRoster(guild)
		}
	}

	Handlers = append(Handlers, plug)
}

func (h *DKPHolder) AddDKPAttendance(dkp int, attendance float64, date time.Time) {

}

func loadGuildRoster(guild *everquest.Guild) {
	for _, member := range guild.Members {
		Roster[member.Name] = &DKPHolder{
			GuildMember: member,
			DKPRank:     getDKPRank(&member),
		}
	}
	// updateRosterDKP() // We don't need to load dkp when the app starts, only when we need to accept bids
}

func updateGuildRoster(guild *everquest.Guild) {
	for _, member := range guild.Members {
		if _, ok := Roster[member.Name]; ok {
			Roster[member.Name].GuildMember = member
			Roster[member.Name].DKPRank = getDKPRank(&member)
		} else {
			Roster[member.Name] = &DKPHolder{
				GuildMember: member,
				DKPRank:     getDKPRank(&member),
			}
		}
	}
	updateRosterDKP()
}

func updateRosterDKP() {
	// TODO: Update DKP and Attendance
	// Info.Printf("Getting Attendance from Google Sheets\n")
	spreadsheetID := configuration.Sheets.RawSheetURL
	readRange := configuration.Sheets.RawSheetName
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		Err.Printf("Unable to retrieve data from sheet: %v", err)
		DiscordF(configuration.Discord.InvestigationChannelID, "Unable to read data from the DKP sheet, cannot calculate winners! - %s\n", err)
		// return errors.New("Unable to retrieve data from sheet: " + err.Error())
	}

	if len(resp.Values) == 0 {
		Err.Printf("Cannot read dkp sheet: %v", resp)
	} else {
		for i, row := range resp.Values {
			if i == 0 {
				continue // skip the header
			}
			name := fmt.Sprintf("%s", row[configuration.Sheets.RawSheetPlayerCol])
			name = strings.TrimSpace(name)
			name = strings.ToTitle(name)
			if name != "" {
				// 06/27/20
				dateString := fmt.Sprintf("%s", row[configuration.Sheets.RawSheetDateCol])
				if dateString == "" { // just to lower the logging
					dateString = "01/01/06" // set to some old date so it's not counted towards current attendance
				}
				date, err := time.Parse("01/02/06", dateString)
				if err != nil {
					Err.Printf("Error converting attendance time to time.Time: %s", err.Error())
					// continue
					date = time.Date(2006, 1, 1, 0, 0, 0, 0, time.Local) // set to some old date so it's not counted towards current attendance
				}
				dkpString := fmt.Sprintf("%s", row[configuration.Sheets.RawSheetDKPCol])
				if dkpString == "" { // just to lower the logging
					dkpString = "0"
				}
				dkpPoints, err := strconv.Atoi(dkpString)
				if err != nil {
					Err.Printf("Error converting attendance points to float: %s", err.Error())
					// continue
					dkpPoints = 0
				}
				attString := fmt.Sprintf("%s", row[configuration.Sheets.RawSheetAttendanceCol])
				if attString == "" { // just to lower the logging
					attString = "0.0"
				}
				attPoints, err := strconv.ParseFloat(attString, 64)
				if err != nil {
					Err.Printf("Error converting attendance points to float: %s", err.Error())
					// continue
					attPoints = 0.0
				}
				addDKPAttendance(name, date, dkpPoints, attPoints)
			}
		}
	}
}

func addDKPAttendance(name string, date time.Time, dkp int, attendance float64) {
	// TODO: Add attendance to the Roster
	if _, ok := Roster[name]; ok {
		//do something here
		Roster[name].DKP += dkp
		Roster[name].AllTime += attendance
		if date.After(time.Now().AddDate(0, 0, -30)) {
			Roster[name].Thirty += attendance
		}
		if date.After(time.Now().AddDate(0, 0, -60)) {
			Roster[name].Sixty += attendance
		}
		if date.After(time.Now().AddDate(0, 0, -90)) {
			Roster[name].Ninety += attendance
		}
	}
}

func getDKPRank(member *everquest.GuildMember) DKPRank {
	if member.Rank == "Inactive" {
		return INACTIVE
	}
	if !member.Alt && member.HasRank([]string{"<<< Guild Leader >>>", "<<< Raid/Class Lead/Recruitment >>>", "<<< Officer >>>", "Raider"}) {
		return MAIN
	}
	if (member.Alt && strings.Contains(member.PublicNote, "nd Main") || member.Alt && strings.Contains(member.PublicNote, "nd main")) && member.HasRank([]string{"<<< Guild Leader >>>", "<<< Raid/Class Lead/Recruitment >>>", "<<< Officer >>>", "Raider"}) { // need to check for spelling mistakes
		return SECONDMAIN
	}
	if member.Rank == "Recruit" {
		return RECRUIT
	}
	if member.Alt {
		return ALT
	}
	if member.Rank == "Member" { // TODO: fix for alts - this actually is fine
		return SOCIAL
	}

	return INACTIVE
}

func DKPRankToString(rank DKPRank) string {
	switch rank {
	case INACTIVE:
		return "Inactive"
	case MAIN:
		return "Main"
	case SECONDMAIN:
		return "Second Main"
	case RECRUIT:
		return "Recruit"
	case SOCIAL:
		return "Social"
	case ALT:
		return "Alt"
	}
	return "Unknown"
}

// Handle for BidPlugin sends a message if it detects a player has gone linkdead.
func (p *BidPlugin) Handle(msg *everquest.EqLog, out io.Writer) {
	if msg.Channel == "guild" && msg.Source == "You" {
		{ // Check for open bid
			result := p.BidOpenMatch.FindStringSubmatch(msg.Msg)
			if len(result) > 0 {
				// result[1] == Item
				itemName := result[1]
				// result[2] == count
				count := 1
				if result[2] != "" {
					count, _ = strconv.Atoi(result[2][1:])
				}
				// result[6] == Open timer
				openTimer := 2
				if result[5] != "" {
					openTimer, _ = strconv.Atoi(result[5])
				}
				id, _ := itemDB.FindIDByName(itemName)
				p.OpenBid(id, count, openTimer, out)
			}
		}
		{ // Check for closed bid
			result := p.BidCloseMatch.FindStringSubmatch(msg.Msg)
			if len(result) > 0 {
				itemName := result[1]
				count := 1
				if result[2] != "" {
					count, _ = strconv.Atoi(result[2][1:])
				}
				id, _ := itemDB.FindIDByName(itemName)
				if id != -1 {
					if _, ok := p.Bids[id]; ok { // Only close bids if item is in the map
						item, _ := itemDB.GetItemByID(id)
						// Dump investigation info

						// Remove item from map
						delete(p.Bids, id)
						fmt.Fprintf(out, "Closed bids on %s(x%d)\n", item.Name, count)
					} else {
						fmt.Fprintf(out, "Bids already closed for %s(x%d)\n", itemName, count)
					}
				}
			}
		}
	}
	if msg.Channel == "tell" {
		// TODO: Add tell to investigation log
		// fmt.Printf("Got tell!: %#+v\n", msg)
		result := p.BidAddMatch.FindStringSubmatch(msg.Msg)
		if len(result) >= 2 {
			itemName := result[1]
			itemID, _ := itemDB.FindIDByName(itemName)
			bid, _ := strconv.Atoi(result[2])
			// fmt.Printf("Result: %#+v itemName: %s itemID: %d bid: %d\n", result, itemName, itemID, bid)
			if _, ok := p.Bids[itemID]; ok {
				if _, ok := Roster[msg.Source]; ok {
					p.Bids[itemID].AddBid(*Roster[msg.Source], bid, *msg)
				}
			}
		}
	}
}

func (p *BidPlugin) Info(out io.Writer) {
	fmt.Fprintf(out, "---------------\n")
	fmt.Fprintf(out, "Name: %s\n", p.Name)
	fmt.Fprintf(out, "Author: %s\n", p.Author)
	fmt.Fprintf(out, "Version: %s\n", p.Version)
	fmt.Fprintf(out, "---------------\n")
}

func (p *BidPlugin) OutputChannel() int {
	return p.Output
}

func (p *BidPlugin) OpenBid(itemID int, quantity int, minutes int, out io.Writer) error {
	if itemID < 0 {
		return errors.New("invalid item ID: " + strconv.Itoa(itemID))
	}
	if quantity < 1 {
		return errors.New("invalid quantity: " + strconv.Itoa(quantity))
	}
	if minutes < 1 {
		return errors.New("invalid duration: " + strconv.Itoa(minutes))
	}
	if _, ok := p.Bids[itemID]; !ok { // Only open bids if item is not already in the map
		item, _ := itemDB.GetItemByID(itemID)
		bidders := make([]*Bidder, 1)
		// for i := range bidders {
		// 	bidders[i] = new(Bidder)
		// }
		p.Bids[itemID] = &OpenBid{
			Item:                 item,
			Quantity:             quantity,
			Duration:             time.Duration(minutes) * time.Minute,
			Start:                time.Now(),
			End:                  time.Now().Add(time.Duration(minutes) * time.Minute),
			Tells:                []everquest.EqLog{},
			Bidders:              bidders,
			Zone:                 currentZone,
			SecondMainBidsAsMain: configuration.Bids.SecondMainsBidAsMains,
			SecondMainMaxBid:     configuration.Bids.SecondMainAsMainMaxBid,
		}
		fmt.Fprintf(out, "Bids open on %s(x%d) for %d minutes.\n", item.Name, quantity, minutes)
		return nil
	}
	return errors.New("bids already open for item: " + strconv.Itoa(itemID))
}

func (b *OpenBid) AddBid(player DKPHolder, amount int, msg everquest.EqLog) {
	pos := b.FindBid(player.Name)
	if pos >= 0 {
		b.Bidders[pos].AttemptedBid = amount
		b.Bidders[pos].Message = msg
		return
	} else {
		bidder := &Bidder{
			Player:       &player,
			AttemptedBid: amount,
			Message:      msg,
		}
		// fmt.Printf("Bidder: %#+v\n", bidder)
		b.Bidders = append(b.Bidders, bidder)
	}
}

func (b *OpenBid) CloseBids() {
	// Refresh DKP

	// Remove cancelled bids
	b.RemoveCancelledBids()
	// Update max dkp based on attempted amount
	b.ApplyDKP()
	// Sort bidders by highest accounting for rank

	// Check for ties

	// Fill in Rot as needed

	// Announce winner

	// Write closed bid investigation file

}

func (b *OpenBid) RemoveCancelledBids() {
	for i := range b.Bidders { // TODO: This won't work cause we modify the slice in the loop
		if b.Bidders[i].AttemptedBid <= 0 {
			b.Bidders = append(b.Bidders[:i], b.Bidders[i+1:]...) // Todo: does this work if bid is final bid in slice
		}
	}
}

func (b *OpenBid) ApplyDKP() { // TODO: Limit 2nd mains to 200 if mains bid
	for i := range b.Bidders {
		if b.Bidders[i].AttemptedBid > b.Bidders[i].Player.DKP {
			b.Bidders[i].Bid = b.Bidders[i].Player.DKP
		} else {
			b.Bidders[i].Bid = b.Bidders[i].AttemptedBid
		}
		if b.Bidders[i].Bid <= 0 {
			b.Bidders[i].Bid = configuration.Bids.MinimumBid
		}
	}
}

func (b *OpenBid) FindBid(name string) int {
	// fmt.Printf("Finding %s in len %d\nBidders: %#+v\n", name, len(b.Bidders), &b.Bidders)
	for pos, bidder := range b.Bidders {
		if bidder == nil {
			continue
		}
		if bidder.Player.Name == name {
			return pos
		}
	}
	return -1
}
