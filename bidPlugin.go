package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	everquest "github.com/Mortimus/goEverquest"
	"github.com/fatih/color"
)

// type BidPlugin Plugin
type BidPlugin struct {
	Plugin
	BidOpenMatch  *regexp.Regexp
	BidCloseMatch *regexp.Regexp
	BidAddMatch   *regexp.Regexp
	BidNumber     *regexp.Regexp
	Bids          map[int]*OpenBid
}

type OpenBid struct {
	Item                 everquest.Item
	Quantity             int
	Duration             time.Duration
	Start                time.Time
	End                  time.Time
	Bidders              []*Bidder
	Zone                 string
	MessageID            string
	SecondMainBidsAsMain bool
	SecondMainMaxBid     int
	WinningBid           int
}

type Bidder struct {
	Player       *DKPHolder
	Message      everquest.EqLog
	AttemptedBid int
	Bid          int
	WonOrTied    bool
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
var updateDKP bool

func init() {
	plug := new(BidPlugin)
	plug.Name = "Bidding detection"
	plug.Author = "Mortimus"
	plug.Version = "1.0.0"
	plug.Output = BIDOUT
	plug.BidOpenMatch, _ = regexp.Compile(configuration.Bids.RegexOpenBid)
	// match1 := `(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`
	// match2 := "'(.+?)(x\\d)*\\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\\sto\\s.+,?\\s?(?:pst)?\\s(\\d+)(?:min|m)(\\d+)?'"
	// plug.BidOpenMatch, _ = regexp.Compile(`(.+?)(x\d)*\s+(?:[Tt][Ee][Ll][Ll][Ss]|[Bb][Ii][Dd][Ss])?\sto\s.+,?\s?(?:pst)?\s(\d+)(?:min|m)(\d+)?`)
	plug.BidCloseMatch, _ = regexp.Compile(configuration.Bids.RegexClosedBid)
	plug.BidAddMatch, _ = regexp.Compile(configuration.Bids.RegexTellBid)
	plug.BidNumber, _ = regexp.Compile(`\d+`)
	plug.Bids = make(map[int]*OpenBid)
	Roster = make(map[string]*DKPHolder)
	path, err := everquest.GetRecentRosterDump(configuration.Everquest.BaseFolder, configuration.Everquest.GuildName)
	if err != nil {
		fmt.Printf("Error finding roster dump: %s", err.Error())
	} else {
		guild := new(everquest.Guild)
		fileLog := log.New(os.Stdout, "[WARN] ", log.Lshortfile|log.Ldate|log.Ltime|log.LUTC|log.Lmsgprefix)
		fullpath := configuration.Everquest.BaseFolder + "/" + path
		apiUploadGuildRoster(fullpath)
		err := guild.LoadFromPath(fullpath, fileLog)
		if err != nil {
			fmt.Printf("Error loading roster dump: %s", err.Error())
		} else {
			loadGuildRoster(guild)
		}
	}

	Handlers = append(Handlers, plug)
	updateDKP = true
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
	fixOutrankingSecondMains()
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
	fixOutrankingSecondMains()
	// updateRosterDKP()
}

func apiUploadGuildRoster(filename string) []byte {
	Info.Printf("Uploading guild roster to API: %s", filename)
	file, err := os.Open(filename)

	if err != nil {
		Err.Println(err)
		return nil
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))

	if err != nil {
		Err.Println(err)
		return nil
	}

	io.Copy(part, file)
	writer.Close()
	request, err := http.NewRequest("POST", configuration.Main.GuildUploadAPIURL, body)

	if err != nil {
		Err.Println(err)
		return nil
	}

	request.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}

	response, err := client.Do(request)

	if err != nil {
		Err.Println(err)
		return nil
	}
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)

	if err != nil {
		Err.Println(err)
	}

	return content
}

func fixOutrankingSecondMains() {
	for _, member := range Roster {
		if member.DKPRank == SECONDMAIN {
			main := getMain(&Roster[member.Name].GuildMember)
			mainRank := &Roster[main].GuildMember
			if getDKPRank(mainRank) < SECONDMAIN {
				Roster[member.Name].Rank = Roster[main].Rank
				Roster[member.Name].PublicNote = ""
			}
		}
	}
}

func updateRosterDKP() {
	// Clear Roster
	// Roster = make(map[string]*DKPHolder)
	for _, member := range Roster { // zero out the roster
		member.DKP = 0
		member.Thirty = 0
		member.Sixty = 0
		member.Ninety = 0
		member.AllTime = 0
	}
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
			name = strings.Title(name)
			if name != "" {
				// 06/27/20
				dateString := fmt.Sprintf("%s", row[configuration.Sheets.RawSheetDateCol])
				if dateString == "" { // just to lower the logging
					dateString = "01/01/06" // set to some old date so it's not counted towards current attendance
				}
				date, err := time.Parse("01/02/06", dateString)
				if err != nil {
					Err.Printf("Error converting attendance time to time.Time at row %d: %s", i+1, err.Error())
					// continue
					date = time.Date(2006, 1, 1, 0, 0, 0, 0, time.Local) // set to some old date so it's not counted towards current attendance
				}
				dkpString := fmt.Sprintf("%s", row[configuration.Sheets.RawSheetDKPCol])
				if dkpString == "" { // just to lower the logging
					dkpString = "0"
				}
				dkpPoints, err := strconv.Atoi(dkpString)
				if err != nil {
					Err.Printf("Error converting attendance points to float at row %d: %s", i+1, err.Error())
					// continue
					dkpPoints = 0
				}
				attString := fmt.Sprintf("%s", row[configuration.Sheets.RawSheetAttendanceCol])
				if attString == "" { // just to lower the logging
					attString = "0.0"
				}
				attPoints, err := strconv.ParseFloat(attString, 64)
				if err != nil {
					Err.Printf("Error converting attendance points to float at row %d: %s", i+1, err.Error())
					// continue
					attPoints = 0.0
				}
				addDKPAttendance(name, date, dkpPoints, attPoints)
			}
		}
	}
	updateAltDKP()
}

// type RawDKP struct {
// 	Name            string `csv:"Name"`
// 	Day             string `csv:"Day"`
// 	Date            string `csv:"Date"`
// 	Raid            string `csv:"Raid"`
// 	AccrualType     string `csv:"Type"`
// 	Reason          string `csv:"Reason"`
// 	Points          string `csv:"Points"`
// 	AltOrSecondMain string `csv:"AltOrSecondMain"`
// 	Class           string `csv:"Class"`
// 	GuildRank       string `csv:"GuildRank"`
// 	Alt             string `csv:"Alt"`
// 	DKPRank         string `csv:"DKPRank"`
// 	AttendanceCalc  string `csv:"AttendanceCalc"`
// 	Level           string `csv:"Level"`
// }

func TimeStamp() string {
	ts := time.Now().Format(time.RFC3339)
	return strings.Replace(ts, ":", "", -1) // get rid of offensive colons
}

func exportSpentDKP(winners []string, winningBid int, itename string) {
	var csvDATA string
	if len(winners) < 1 {
		return
	}
	for _, winner := range winners {
		// tempDATA := csvDATA
		if _, ok := Roster[winner]; !ok { // Verify the member is in the map
			continue
		}
		main := getMain(&Roster[winner].GuildMember)
		day := time.Now().Format("Mon")
		date := time.Now().Format("01/02/06")
		points := fmt.Sprintf("-%d", winningBid)
		var alt string
		if main != winner {
			alt = winner
		}
		row := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s\n", main, day, date, date+" BIDBOT_AUTO_FILL", "Spent", itename, points, alt) // Name, Day, Date, Raid, Type, Reason, Points, AltOrSecondMain
		csvDATA += row
	}
	if csvDATA == "" {
		return
	}
	DiscordF(configuration.Discord.InvestigationChannelID, "DKP Entry for %s\n```\n%v\n```", itename, csvDATA)
}

func exportDKP(path string) {
	// TODO: Update DKP and Attendance
	// Info.Printf("Getting Attendance from Google Sheets\n")
	spreadsheetID := configuration.Sheets.RawSheetURL
	readRange := configuration.Sheets.RawSheetName
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		Err.Printf("Unable to retrieve data from sheet: %v", err)
		DiscordF(configuration.Discord.InvestigationChannelID, "Unable to read data from the DKP sheet, cannot perform backup! - %s\n", err)
		// return errors.New("Unable to retrieve data from sheet: " + err.Error())
	}

	if len(resp.Values) == 0 {
		Err.Printf("Cannot read dkp sheet: %v", resp)
	} else {
		csvFile, err := os.Create(path)
		if err != nil {
			DiscordF(configuration.Discord.InvestigationChannelID, "Unable to write DKP backup! - %s\n", err)
			// log.Printf("failed creating file: %s", err)
		}
		fmt.Printf("Exporting DKP to %s\n", path)
		csvwriter := csv.NewWriter(csvFile)
		for _, row := range resp.Values {
			// if i == 0 {
			// 	continue // skip the header
			// }
			var rec []string
			for _, record := range row {
				rec = append(rec, fmt.Sprintf("%s", record))
			}
			_ = csvwriter.Write(rec)
			csvwriter.Flush()
		}
		csvwriter.Flush()
		csvFile.Close()
		fmt.Printf("Exporting DKP to %s COMPLETE\n", path)
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

func updateAltDKP() {
	for _, member := range Roster {
		if member.GuildMember.Alt {
			if _, ok := Roster[member.GuildMember.Name]; ok {
				// fmt.Printf("Updating %s with %s' DKP", name, member.GuildMember.Name)
				Roster[member.GuildMember.Name].DKP = Roster[getMain(&member.GuildMember)].DKP
				Roster[member.GuildMember.Name].AllTime = Roster[getMain(&member.GuildMember)].AllTime
				Roster[member.GuildMember.Name].Thirty = Roster[getMain(&member.GuildMember)].Thirty
				Roster[member.GuildMember.Name].Sixty = Roster[getMain(&member.GuildMember)].Sixty
				Roster[member.GuildMember.Name].Ninety = Roster[getMain(&member.GuildMember)].Ninety
			}
		}
	}
}

func getDKPRank(member *everquest.GuildMember) DKPRank {
	if member.Rank == "Inactive" {
		return INACTIVE
	}
	if !member.Alt && member.HasRank([]string{"<<< Guild Leader >>>", "<<< Class Lead >>>", "<<< Officer >>>", "Raider"}) {
		return MAIN
	}
	if member.Alt && strings.Contains(member.PublicNote, "nd Main") || member.Alt && strings.Contains(member.PublicNote, "nd main") {
		// Check if their main has a lower rank than SECONDMAIN
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

func getMain(member *everquest.GuildMember) string {
	if member.Alt {
		if strings.Contains(member.PublicNote, "'") { // Mortimus's 2nd Main Mortimus's Alt
			s := strings.Split(member.PublicNote, "'")
			if _, ok := Roster[s[0]]; ok {
				return s[0]
			}
		}
		if strings.Contains(member.PublicNote, " ") { // Mortimus 2nd Main Mortimus Alt
			s := strings.Split(member.PublicNote, " ")
			if _, ok := Roster[s[0]]; ok {
				return s[0]
			}
		}
		// Warn.Printf("Error finding main for %s based on %s\n", member.Name, member.PublicNote)
	}
	return member.Name
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
	if (msg.Channel == "guild" && msg.Source == "You") || (msg.Channel == "raid" && msg.Source == "You") {
		{ // Check for open bid
			if p.HandleMultiBids(msg, out) {
				return // it's a multi bid, no need to process more
			}
			result := p.BidOpenMatch.FindStringSubmatch(msg.Msg)
			// p.BidOpenMatch.String()
			if len(result) > 3 {
				// result[1] == Item
				itemName := result[1]
				// result[2] == count
				count := 1
				if result[2] != "" {
					count, _ = strconv.Atoi(result[2][1:])
				}
				// result[6] == Open timer
				openTimerMin := 2
				if result[3] != "" {
					openTimerMin, _ = strconv.Atoi(result[3])
				}
				openTimerSec := 0
				if result[4] != "" {
					openTimerSec, _ = strconv.Atoi(result[4])
				}
				// fmt.Printf("Name: %s Count: %d Min: %d Sec: %d\n", itemName, count, openTimerMin, openTimerSec)
				id, _ := itemDB.FindIDByName(itemName)
				// fmt.Printf("OpenID: %d\n", id)
				p.OpenBid(id, count, openTimerMin, openTimerSec, out)
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
						p.Bids[id].CloseBids(out)
						// Remove item from map
						delete(p.Bids, id)
						Info.Printf("Closed bids on %s (x%d)\n", item.Name, count)
					} else {
						Err.Printf("Bids already closed for %s(x%d)\n", itemName, count)
					}
				}
			}
		}
	}
	if msg.Channel == "tell" {
		// fmt.Printf("Got tell!: %#+v\n", msg)
		p.HandleTell(msg)
		// result := p.BidAddMatch.FindStringSubmatch(msg.Msg)
		// if len(result) >= 2 {
		// 	itemName := result[1]
		// 	itemID, _ := itemDB.FindIDByName(itemName)
		// 	bid, _ := strconv.Atoi(result[2])
		// 	// fmt.Printf("Result: %#+v itemName: %s itemID: %d bid: %d bidder: %s msg: %#+v\n", result, itemName, itemID, bid, msg.Source, msg)
		// 	if _, ok := p.Bids[itemID]; ok {
		// 		if _, ok := Roster[msg.Source]; ok {
		// 			p.Bids[itemID].AddBid(*Roster[msg.Source], bid, *msg)
		// 		}
		// 	}
		// }
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

func (p BidPlugin) HandleTell(msg *everquest.EqLog) {
	if strings.ContainsAny(msg.Msg, "0123456789") {
		// totalItems := len(p.Bids)
		// var curItem int
		for id, item := range p.Bids {
			// curItem++
			if strings.Contains(msg.Msg, item.Item.Name) {
				bidString := strings.Replace(msg.Msg, item.Item.Name, "", 1)
				bidString = strings.Replace(bidString, "2nd", "", -1) // Remove 2nd main talk
				bidString = p.BidNumber.FindString(bidString)
				bid, err := strconv.Atoi(bidString)
				// fmt.Printf("BidString: %s Bid: %d\n", bidString, bid)
				if err != nil {
					Err.Printf("Error converting bid to number, %s\n", err)
				}
				if bid >= 0 {
					source := msg.Source
					if source == "You" {
						source = getPlayerName(configuration.Everquest.LogPath)
					}
					p.Bids[id].AddBid(*Roster[source], bid, *msg)
					return
				}
			} // else {
			// fmt.Printf("Cur: %d Total: %d\n", curItem, totalItems)
			// if curItem == totalItems {
			// 	printMessage(msg)
			// }
			//}
		}
	}
}

func printMessage(msg *everquest.EqLog) {
	// log.Printf("Printing %s\n", msg.Msg)
	if !strings.ContainsAny(msg.Msg, "0123456789") {
		i := strings.Index(msg.Msg, " '")
		var message string
		if strings.Contains(msg.Msg, "[queued]") {
			message = msg.Msg[i+3 : len(msg.Msg)-1]
		} else {
			message = msg.Msg[i+2 : len(msg.Msg)-1]
		}
		if msg.Source == "You" && strings.Contains(msg.Msg, "told") {
			// You told sacristan, 'can i get brells?'
			// [Sun Oct 24 18:24:48 2021] You told Ravnor '[queued], ty sir'
			// [Sun Oct 24 18:27:52 2021] You told Drae, 'AFK Message: Bla bla bla'
			// fmt.Printf("MSG: %#+v\n", msg.Msg)
			tTarget := strings.Title(msg.Msg[9:i])
			tTarget = strings.Replace(tTarget, ",", "", 1)
			printChan <- fmt.Sprintf("%s->%s: %s\n", color.RedString(msg.Source), color.GreenString(tTarget), color.MagentaString(message))
			// fmt.Printf("%s: %s\n", color.RedString(msg.Source), color.MagentaString(message))
		} else {
			printChan <- fmt.Sprintf("%s: %s\n", color.CyanString(msg.Source), color.YellowString(message))
			// fmt.Printf("%s: %s\n", color.CyanString(msg.Source), color.MagentaString(message))
		}
	}
}

func (p *BidPlugin) OpenBid(itemID int, quantity int, minutes int, seconds int, out io.Writer) error {
	if itemID < 0 {
		return errors.New("invalid item ID: " + strconv.Itoa(itemID))
	}
	if quantity < 1 {
		return errors.New("invalid quantity: " + strconv.Itoa(quantity))
	}
	if minutes < 1 {
		return errors.New("invalid duration: " + strconv.Itoa(minutes))
	}
	if seconds < 0 {
		return errors.New("invalid duration: " + strconv.Itoa(seconds))
	}

	if _, ok := p.Bids[itemID]; !ok { // Only open bids if item is not already in the map
		item, _ := itemDB.GetItemByID(itemID)
		bidders := make([]*Bidder, 0)
		// for i := range bidders {
		// 	bidders[i] = new(Bidder)
		// }
		p.Bids[itemID] = &OpenBid{
			Item:                 item,
			Quantity:             quantity,
			Duration:             (time.Duration(minutes) * time.Minute) + (time.Duration(seconds) * time.Second),
			Start:                time.Now(),
			End:                  time.Now().Add(time.Duration(minutes) * time.Minute).Add(time.Duration(seconds) * time.Second),
			Bidders:              bidders,
			Zone:                 currentZone,
			SecondMainBidsAsMain: configuration.Bids.SecondMainsBidAsMains,
			SecondMainMaxBid:     configuration.Bids.SecondMainAsMainMaxBid,
		}
		p.Bids[itemID].MessageID = DiscordF(configuration.Discord.LootChannelID, "> Bids open on %s (x%d) for %d minutes %d seconds.\n```%s```%s%d", item.Name, quantity, minutes, seconds, getItemDesc(item), configuration.Main.LucyURLPrefix, item.ID)
		// fmt.Fprintf(out, "> Bids open on %s (x%d) for %d minutes.\n```%s```%s%d", item.Name, quantity, minutes, getItemDesc(item), configuration.Main.LucyURLPrefix, item.ID)
		return nil
	} else {
		if p.Bids[itemID].Quantity != quantity { // Modify amount of winners
			// fmt.Fprintf(out, "Changing %s bid quantity to %d", p.Bids[itemID].Item.Name, quantity)
			header := fmt.Sprintf("> Bids open on %s (x%d) for %d minutes %d seconds.", p.Bids[itemID].Item.Name, quantity, minutes, seconds)
			err := updateHeader(configuration.Discord.LootChannelID, p.Bids[itemID].MessageID, header)
			if err != nil {
				Err.Println(err)
			}
			p.Bids[itemID].Quantity = quantity
			return nil
		}
	}
	return errors.New("bids already open for item: " + strconv.Itoa(itemID))
}

func (b *OpenBid) AddBid(player DKPHolder, amount int, msg everquest.EqLog) {
	pos := b.FindBid(player.Name)
	if pos >= 0 {
		if amount > configuration.Bids.MinimumBid {
			b.Bidders[pos].AttemptedBid = amount
			b.Bidders[pos].Message = msg
			return
		}
		b.Bidders = removeBidder(b.Bidders, pos)
		return
	} else {
		bidder := &Bidder{
			Player:       &player,
			AttemptedBid: amount,
			Message:      msg,
		}
		if !canEquip(b.Item, player.GuildMember) {
			DiscordF(configuration.Discord.InvestigationChannelID, "```diff\n-A player bid on %s that cannot use it, if it is not cancelled it will be auto investigated. %s\n```", b.Item.Name, b.Item.GetClasses())
		}
		// fmt.Printf("Bidder: %#+v\n", bidder)
		b.Bidders = append(b.Bidders, bidder)
	}
}

func (b *OpenBid) CloseBids(out io.Writer) {
	b.End = time.Now()
	// Refresh DKP
	if updateDKP {
		updateRosterDKP()
	}
	// Update max dkp based on attempted amount
	b.ApplyDKP()
	// Sort bidders by highest accounting for rank
	b.SortBids()

	// Check for ties
	ties := b.CheckTiesAndApplyWinners()
	var tieCount int
	var tieAnnounce string
	for tie := range ties {
		needsRolled = append(needsRolled, tie) // This allows for roll detection
		if tieCount == 0 {
			tieAnnounce = fmt.Sprintf("```diff\n- /rand 1000 needed for %s from %s", b.Item.Name, tie)
		} else {
			tieAnnounce = fmt.Sprintf("%s, %s", tieAnnounce, tie)
		}
		tieCount++
	}
	var tied bool
	if tieCount > 0 {
		tieAnnounce = fmt.Sprintf("%s```", tieAnnounce)
		// fmt.Fprintf(out, "%s```", tieAnnounce)
		tied = true
		err := updateMessage(configuration.Discord.LootChannelID, b.MessageID, tieAnnounce)
		if err != nil {
			Err.Println(err)
		}
	}

	// Find winning cost
	b.WinningBid = b.FindWinningBid()

	if b.WinningBid < 0 {
		b.WinningBid = 0
	}
	// Announce winner and include rot if needed
	winners := b.GetWinnerNames()
	if len(winners) < b.Quantity {
		// Fill remaining with Rots
		neededWinners := b.Quantity - len(winners)
		for i := 0; i < neededWinners; i++ {
			winners = append(winners, "Rot")
		}
	}
	if !tied {
		wonMessage := fmt.Sprintf("> %s (x%d) won for %d DKP", b.Item.Name, b.Quantity, b.WinningBid)
		err := updateHeader(configuration.Discord.LootChannelID, b.MessageID, wonMessage)
		if err != nil {
			Err.Println(err)
		}
	} else {
		wonMessage := fmt.Sprintf("> %s (x%d) won for %d DKP AFTER roll off", b.Item.Name, b.Quantity, b.WinningBid)
		err := updateHeader(configuration.Discord.LootChannelID, b.MessageID, wonMessage)
		if err != nil {
			Err.Println(err)
		}
	}
	b.GenerateInvestigation()

	winnerMessage := "```"

	var playerWon bool
	for i, win := range winners {
		if win == "Rot" {
			winnerMessage = fmt.Sprintf("%s%d: %s\n", winnerMessage, i+1, win)
		} else {
			playerWon = true
			winnerMessage = fmt.Sprintf("%s%d: %s\tCurrentDKP(%d) - WinningBid(%d) = %d DKP\n", winnerMessage, i+1, win, Roster[win].DKP, b.WinningBid, Roster[win].DKP-b.WinningBid)
		}

	}
	if playerWon { // don't require looted for rotted items
		needsLooted = append(needsLooted, b.Item.Name)
	}
	winnerMessage = fmt.Sprintf("> Winner(s)\n%s```", winnerMessage)
	// TODO: Update original message with this info appended
	err := updateMessage(configuration.Discord.LootChannelID, b.MessageID, winnerMessage)
	if err != nil {
		Err.Println(err)
	}
	err = updateMessage(configuration.Discord.LootChannelID, b.MessageID, "   v\t\t[INVESTIGATION READY]")
	if err != nil {
		Err.Println(err)
	}
	if configuration.Discord.UseDiscord {
		err = discord.MessageReactionAdd(configuration.Discord.LootChannelID, b.MessageID, configuration.Discord.InvestigationStartEmoji)
		if err != nil {
			Err.Printf("Error adding base reaction: %s", err.Error())
		}
	}
	if b.AutoInvestigate() {
		uploadArchive(b.MessageID)
	}
	// Upload csv of winner dkp changes
	exportSpentDKP(winners, b.WinningBid, b.Item.Name)
	// fmt.Fprintf(out, "%s```[%s]", winnerMessage, hash)
	// Write closed bid investigation file

}

func updateMessage(channelID, messageID, append string) error {
	if !configuration.Discord.UseDiscord {
		return nil
	}
	msg, err := discord.ChannelMessage(channelID, messageID)
	if err != nil {
		return err
	}
	content := msg.Content
	content = fmt.Sprintf("%s\n%s\n", content, append)
	_, err = discord.ChannelMessageEdit(channelID, messageID, content)
	if err != nil {
		return err
	}
	return nil
}

func updateHeader(channelID, messageID, header string) error {
	if !configuration.Discord.UseDiscord {
		return nil
	}
	msg, err := discord.ChannelMessage(channelID, messageID)
	if err != nil {
		return err
	}
	content := msg.Content
	split := strings.Split(content, "\n")
	split[0] = header
	content = strings.Join(split, "\n")
	_, err = discord.ChannelMessageEdit(channelID, messageID, content)
	if err != nil {
		return err
	}
	return nil
}

type BidInvestigation struct {
	WinningBid           int                   `json:"WinningBid"`
	ItemName             string                `json:"ItemName"`
	Quantity             int                   `json:"Quantity"`
	SecondMainBidsAsMain bool                  `json:"SecondMainBidsAsMain"`
	SecondMainMaxBid     int                   `json:"SecondMainMaxBid"`
	Started              string                `json:"Started"`
	Ended                string                `json:"Ended"`
	Bidders              []InvestigationBidder `json:"Bidders"`
	Logs                 []InvestigationLog    `json:"Logs"`
}

type InvestigationBidder struct {
	Player       string `json:"Player"`
	Main         string `json:"Main"`
	BidAttempted int    `json:"BidAttempted"`
	BidApplied   int    `json:"BidApplied"`
	DKP          int    `json:"DKP"`
	DKPRank      string `json:"DKPRank"`
	DKPRankValue int    `json:"DKPRankValue"`
	CanEquip     bool   `json:"CanEquip"`
	Message      string `json:"Message"`
	WonOrTied    bool   `json:"WonOrTied"`
}

type InvestigationLog struct {
	Message      string `json:"Message"`
	Received     string `json:"Received"`
	Player       string `json:"Player"`
	Main         string `json:"Main"`
	DKP          int    `json:"DKP"`
	DKPRank      string `json:"DKPRank"`
	DKPRankValue int    `json:"DKPRankValue"`
	CanEquip     bool   `json:"CanEquip"`
	InBidWindow  bool   `json:"InBidWindow"`
}

func (b *OpenBid) GenerateInvestigation() string {
	var Bidders []InvestigationBidder
	for _, bidder := range b.Bidders {
		Bidders = append(Bidders, InvestigationBidder{
			Player:       bidder.Player.Name,
			Main:         getMain(&bidder.Player.GuildMember),
			BidAttempted: bidder.AttemptedBid,
			BidApplied:   bidder.Bid,
			DKP:          bidder.Player.DKP,
			DKPRank:      DKPRankToString(bidder.Player.DKPRank),
			DKPRankValue: int(bidder.Player.DKPRank),
			CanEquip:     canEquip(b.Item, bidder.Player.GuildMember),
			Message:      bidder.Message.Msg,
			WonOrTied:    bidder.WonOrTied,
		})
	}
	var Logs []InvestigationLog
	for _, log := range investigation.Messages {
		var gMember *DKPHolder
		if log.Source == "You" {
			log.Source = getPlayerName(configuration.Everquest.LogPath)
		}
		if _, ok := Roster[log.Source]; ok {
			gMember = Roster[log.Source]
		} else {
			Err.Printf("Cannot find %s in roster\n", log.Source)
			gMember = genUnknownMember(log.Source)
		}
		// fmt.Printf("Main of %s is %s :: Msg: %s\n", log.Source, getMain(&gMember.GuildMember), log.Msg)
		BidLog := InvestigationLog{
			Message:      log.Msg,
			Received:     log.T.Format(time.RFC822),
			Player:       log.Source,
			Main:         getMain(&gMember.GuildMember),
			DKP:          gMember.DKP,
			DKPRank:      DKPRankToString(gMember.DKPRank),
			DKPRankValue: int(gMember.DKPRank),
			CanEquip:     canEquip(b.Item, gMember.GuildMember),
			InBidWindow:  isBetweenTime(log.T, b.Start, b.End),
		}
		Logs = append(Logs, BidLog)
	}
	investigation := BidInvestigation{
		WinningBid:           b.WinningBid,
		ItemName:             b.Item.Name,
		Quantity:             b.Quantity,
		SecondMainBidsAsMain: b.SecondMainBidsAsMain,
		SecondMainMaxBid:     b.SecondMainMaxBid,
		Started:              b.Start.Format(time.RFC822),
		Ended:                b.End.Format(time.RFC822),
		Bidders:              Bidders,
		Logs:                 Logs,
	}
	hash := b.MessageID
	filename := hash + ".json"
	Info.Printf("Writing archive %s to file", filename)
	file, err := json.MarshalIndent(investigation, "", " ")
	if err != nil {
		Err.Printf("Error converting to JSON: %s", err.Error())
	}

	err = ioutil.WriteFile("archive/"+filename, file, 0644)
	if err != nil {
		Err.Printf("Error writing archive to file: %s", err.Error())
	}
	archives = append(archives, hash) // add to known archive
	return hash
}

func isBetweenTime(t time.Time, start, end time.Time) bool {
	return t.After(start) && t.Before(end)
}

func (b *OpenBid) AutoInvestigate() bool {
	for _, bidder := range b.Bidders {
		if bidder.Bid > 0 && !canEquip(b.Item, bidder.Player.GuildMember) {
			return true
		}
	}
	if b.WinningBid == 0 && len(b.GetWinnerNames()) != 0 {
		DiscordF(configuration.Discord.InvestigationChannelID, "```diff\n-Somehow we have a 0 dkp win again auto investigating\n```")
		return true
	}
	if b.WinningBid > 0 && len(b.GetWinnerNames()) == 0 {
		DiscordF(configuration.Discord.InvestigationChannelID, "```diff\n-Somehow we have a Rot spending DKP auto investigating\n```")
		return true
	}
	return false
}

func getArchiveList() []string { // TODO: get directory listing on archives
	var files []string

	root := "./archive"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error { // This can never have an error TODO: fix this
		name := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
		files = append(files, name)
		return nil
	})
	if err != nil {
		Err.Printf("Error reading archives: %s", err.Error())
	}
	return files
}

func isArchive(id string) bool {
	for _, arc := range archives {
		if arc == id {
			return true
		}
	}
	return false
}

func genUnknownMember(name string) *DKPHolder {
	guildMember := everquest.GuildMember{
		Name:                name,
		Level:               0,
		Class:               "Unknown",
		Rank:                "Unknown",
		Alt:                 false,
		LastOnline:          time.Now(),
		Zone:                "Unknown",
		PublicNote:          "Who am I?",
		PersonalNote:        "Who am I?",
		TributeStatus:       false,
		TrophyTributeStatus: false,
		Donations:           -1,
	}
	return &DKPHolder{
		GuildMember: guildMember,
		DKP:         0,
		DKPRank:     INACTIVE,
		Thirty:      0,
		Sixty:       0,
		Ninety:      0,
		AllTime:     0,
	}
}

func canEquip(item everquest.Item, player everquest.GuildMember) bool {
	classes := item.GetClasses()
	if len(classes) == 0 { // No class item, likely for a quest, or some other use
		return true
	}
	for _, class := range classes {
		if class == player.Class {
			return true
		}
	}
	return false
}

func (b *OpenBid) FindWinningBid() int {
	const DEBUG = false
	winningBid := configuration.Bids.MinimumBid
	var winRank DKPRank
	if len(b.Bidders) == 0 {
		return 0 // no one bid, rot
	}
	var winners int
	var lastbid int
	for i, bidder := range b.Bidders {
		if DEBUG {
			fmt.Printf("winningBid: %d winRank: %d winners: %d lastbid: %d bid: %d name: %s rank: %d i: %d\n", winningBid, winRank, winners, lastbid, bidder.Bid, bidder.Player.Name, bidder.Player.DKPRank, i)
		}
		if i == 0 && bidder.Bid == 0 {
			return 0 // ROT
		}
		if bidder.Bid == 0 {
			continue
		}
		winners++ // We don't want to include cancelled bids in winningbid calculations
		if GetEffectiveDKPRank(bidder.Player.DKPRank) > winRank || winners <= b.Quantity {
			winRank = GetEffectiveDKPRank(bidder.Player.DKPRank)
			lastbid = bidder.Bid
		} else {
			if bidder.Bid == lastbid {
				winningBid = bidder.Bid // Tie
			} else {
				winningBid = bidder.Bid + 5
			}
			if GetEffectiveDKPRank(bidder.Player.DKPRank) != winRank {
				winningBid = configuration.Bids.MinimumBid
			}
			break
		}
	}
	if DEBUG {
		fmt.Printf("winningBid: %d winRank: %d winners: %d lastbid: %d\n", winningBid, winRank, winners, lastbid)
	}
	return winningBid
}
func (b *OpenBid) CheckTiesAndApplyWinners() map[string]interface{} { // We are assuming bids are applied and sorted before this is called
	tiedPlayers := make(map[string]interface{})
	if len(b.Bidders) <= b.Quantity {
		for i := range b.Bidders { // Everyone is a winner!
			b.Bidders[i].WonOrTied = true
		}
		return tiedPlayers // more items than potential ties, so no ties
	}
	var tieBid int
	var tiedRank DKPRank
	var validBids int
	for i := range b.Bidders {
		if b.Bidders[i].Bid == 0 {
			continue // a main can cancel and a lower tier might tie
		}
		validBids++
		if validBids <= b.Quantity && tieBid != b.Bidders[i].Bid {
			b.Bidders[i].WonOrTied = true
			tieBid = b.Bidders[i].Bid
			tiedRank = GetEffectiveDKPRank(b.Bidders[i].Player.DKPRank)
			tiedPlayers = make(map[string]interface{}) // clear the tied, we might have had guaranteed winners that tied
			continue                                   // not a tie, check next bid
		}
		if validBids > b.Quantity && tieBid != b.Bidders[i].Bid {
			return tiedPlayers // we have found all the possible tie bids, so we are done
		}
		if tieBid == b.Bidders[i].Bid && GetEffectiveDKPRank(b.Bidders[i].Player.DKPRank) == tiedRank {
			b.Bidders[i].WonOrTied = true
			tiedPlayers[b.Bidders[i-1].Player.Name] = nil // ensure the original tie bid is here
			tiedPlayers[b.Bidders[i].Player.Name] = nil
		}
	}
	return tiedPlayers
}

func (b *OpenBid) GetWinnerNames() []string {
	var winners []string
	for _, bidder := range b.Bidders {
		if bidder.WonOrTied {
			winners = append(winners, bidder.Player.Name)
		}
	}
	return winners
}

func (b *OpenBid) SortBids() { // TODO: This isn't working correctly -> fixed, but check testing
	/*
		INACTIVE = iota
		SOCIAL
		ALT
		RECRUIT
		SECONDMAIN
		MAIN
	*/
	// Sort by Bid
	// fmt.Printf("Pre-Sort\n")
	// b.printBidders()
	sort.Sort(sort.Reverse(ByBid(b.Bidders)))
	// Remove each effective rank into seperate slice
	var mains, secondmains, recruits, alts, socials, inactives []*Bidder
	for i := range b.Bidders {
		if GetEffectiveDKPRank(b.Bidders[i].Player.DKPRank) == MAIN {
			mains = append(mains, b.Bidders[i])
		}
		if GetEffectiveDKPRank(b.Bidders[i].Player.DKPRank) == SECONDMAIN {
			secondmains = append(secondmains, b.Bidders[i])
		}
		if GetEffectiveDKPRank(b.Bidders[i].Player.DKPRank) == RECRUIT {
			recruits = append(recruits, b.Bidders[i])
		}
		if GetEffectiveDKPRank(b.Bidders[i].Player.DKPRank) == ALT {
			alts = append(alts, b.Bidders[i])
		}
		if GetEffectiveDKPRank(b.Bidders[i].Player.DKPRank) == SOCIAL {
			socials = append(socials, b.Bidders[i])
		}
		if GetEffectiveDKPRank(b.Bidders[i].Player.DKPRank) == INACTIVE {
			inactives = append(inactives, b.Bidders[i])
		}
	}
	// nullify b.bidders
	b.Bidders = nil
	b.Bidders = append(b.Bidders, mains...)
	b.Bidders = append(b.Bidders, secondmains...)
	b.Bidders = append(b.Bidders, recruits...)
	b.Bidders = append(b.Bidders, alts...)
	b.Bidders = append(b.Bidders, socials...)
	b.Bidders = append(b.Bidders, inactives...)
	// fmt.Printf("Post-Sort\n")
	// b.printBidders()
}

func (b *OpenBid) printBidders() {
	for _, bidder := range b.Bidders {
		fmt.Printf("%s: %d\n", bidder.Player.Name, bidder.Bid)
	}
}

type ByBid []*Bidder

func (a ByBid) Len() int           { return len(a) }
func (a ByBid) Less(i, j int) bool { return a[i].Bid < a[j].Bid }
func (a ByBid) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

type ByRank []*Bidder

func (a ByRank) Len() int { return len(a) }
func (a ByRank) Less(i, j int) bool {
	return GetEffectiveDKPRank(a[i].Player.DKPRank) < GetEffectiveDKPRank(a[j].Player.DKPRank)
}
func (a ByRank) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

type ByBidAndRank []*Bidder

func (a ByBidAndRank) Len() int { return len(a) }
func (a ByBidAndRank) Less(i, j int) bool {
	if a[i].Bid < a[j].Bid {
		return true
	}
	if a[i].Bid > a[j].Bid {
		return false
	}
	return GetEffectiveDKPRank(a[i].Player.DKPRank) < GetEffectiveDKPRank(a[j].Player.DKPRank)
}
func (a ByBidAndRank) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

func GetEffectiveDKPRank(rank DKPRank) DKPRank {
	if configuration.Bids.SecondMainsBidAsMains && rank == SECONDMAIN {
		return MAIN
	}
	return rank
}

func (b *OpenBid) ApplyDKP() {
	for i := range b.Bidders {
		if b.Bidders[i].AttemptedBid > configuration.Bids.MaxBid { // This needs to be a smaller number so overflows don't happen and break rounding
			b.Bidders[i].AttemptedBid = configuration.Bids.MaxBid
		}
		b.Bidders[i].Player.DKP = Roster[getMain(&b.Bidders[i].Player.GuildMember)].DKP // Apply the latest roster values to the bidder -> move to a function and apply secondmain/alt dkp
		if b.Bidders[i].AttemptedBid > b.Bidders[i].Player.DKP {
			if b.Bidders[i].Player.DKP < configuration.Bids.MinimumBid { // Todo: Need to make a test for this
				b.Bidders[i].Bid = configuration.Bids.MinimumBid
			} else {
				b.Bidders[i].Bid = b.Bidders[i].Player.DKP
			}
		} else {
			b.Bidders[i].Bid = b.Bidders[i].AttemptedBid
		}
		if b.Bidders[i].AttemptedBid > 0 && b.Bidders[i].AttemptedBid < configuration.Bids.MinimumBid {
			b.Bidders[i].Bid = configuration.Bids.MinimumBid
		}
		if b.Bidders[i].AttemptedBid%configuration.Bids.Increments != 0 && b.Bidders[i].Bid%configuration.Bids.Increments != 0 { // if you fail to bid in correct increments, we are setting you to minimum bid
			// We should round down
			rounded := roundDown(b.Bidders[i].AttemptedBid)
			// fmt.Printf("Rounded: %d\n", rounded)
			if rounded < configuration.Bids.MinimumBid {
				rounded = configuration.Bids.MinimumBid
			}
			// fmt.Printf("Rounded Post: %d\n", rounded)
			b.Bidders[i].Bid = rounded
		}
		if b.Bidders[i].AttemptedBid <= 0 { // Cancelled Bid
			b.Bidders[i].Bid = 0
		}
		if configuration.Bids.SecondMainsBidAsMains && b.Bidders[i].Player.DKPRank == SECONDMAIN && b.Bidders[i].Bid > configuration.Bids.SecondMainAsMainMaxBid { // limit to 200 dkp always on secondmains for primary content
			b.Bidders[i].Bid = configuration.Bids.SecondMainAsMainMaxBid
		}
	}
}

func roundDown(n int) int {
	f := float64(n)
	fAmount := float64(configuration.Bids.Increments)
	rounded := int(math.Round(f/fAmount) * fAmount)
	// fmt.Printf("f: %f fAmount: %f rounded: %d f/fAmount: %f *fAmount: %f\n", f, fAmount, rounded, math.Round(f/fAmount), math.Round(f/fAmount)*fAmount)
	if rounded > n {
		return rounded - configuration.Bids.Increments
	}
	return rounded
}

func (b *OpenBid) MainsHaveBid() bool {
	for i := range b.Bidders {
		if b.Bidders[i].Player.DKPRank == MAIN && b.Bidders[i].AttemptedBid > 0 {
			return true
		}
	}
	return false
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

func removeBidder(slice []*Bidder, pos int) []*Bidder {
	return append(slice[:pos], slice[pos+1:]...)
}

func (p *BidPlugin) HandleMultiBids(msg *everquest.EqLog, out io.Writer) bool {
	// fmt.Printf("Handling multi bid\n")
	msgStart := strings.Index(msg.Msg, "'")
	if msgStart == -1 {
		return false
	}
	bidMsg := msg.Msg[msgStart+1:]
	if !strings.Contains(bidMsg, "|") { // Not multi bid
		return false
	}
	bids := strings.Split(bidMsg, "|")
	for i := range bids { // remove extra white space
		bids[i] = strings.TrimSpace(bids[i])
	}
	notice := strings.Index(bids[len(bids)-1], " bids")
	if notice == -1 {
		notice = strings.Index(bids[len(bids)-1], " tells")
		if notice == -1 { // still no bid open indicators
			// fmt.Printf("Cannot find bid open indicator")
			return false
		}
	}
	bidinfo := bids[len(bids)-1][notice:]
	bids[len(bids)-1] = bids[len(bids)-1][:notice] // remove the bid info from the last bid
	items := make(map[string]int)
	// find duplicates and increase quantity
	for _, i := range bids {
		if val, ok := items[i]; ok {
			items[i] = val + 1
		} else {
			items[i] = 1
		}
	}
	// fmt.Printf("Notice: %s\n", bidinfo)
	// Determine how long bid will be open
	nSplit := strings.Split(bidinfo, " ")
	infoTime := nSplit[len(nSplit)-1]
	minLoc := strings.Index(infoTime, "min")
	if minLoc == -1 {
		return false
	}
	minStr := infoTime[:minLoc]
	mins, err := strconv.Atoi(minStr)
	if err != nil {
		log.Printf("Error converting min to int: %s\n", err)
		mins = 2
	}
	// TODO: get seconds?

	// Open the bids
	for item, count := range items {
		id, err := itemDB.FindIDByName(item)
		if err != nil {
			fmt.Fprintf(out, "Error finding item %s: %s\n", item, err)
			continue
		}
		// fmt.Printf("Opening %dx%d for bids for %d minutes\n", id, count, mins)
		p.OpenBid(id, count, mins, 0, out)
	}
	return true
}
