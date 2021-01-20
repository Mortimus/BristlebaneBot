package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

var bidLock sync.Mutex
var bids = map[string]*BidItem{}
var itemDB map[string]int
var discord *discordgo.Session
var investigation Investigation
var currentTime time.Time // for simulating time
var archives []string     // stores all known archive files for recall

func main() {
	// Open Configuration and set log output
	configFile, err := os.OpenFile(configuration.LogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer configFile.Close()
	log.SetOutput(configFile)
	l := LogInit("main-main.go")
	defer l.End()

	loadItemDB(configuration.LucyItems)
	// bufferedRead(configuration.EQLogPath, configuration.ReadEntireLog)

	// Create a new Discord session using the provided bot token.
	discord, err = discordgo.New("Bot " + configuration.DiscordToken)
	if err != nil {
		l.FatalF("Error creating Discord session: %v", err)
	}
	defer discord.Close()
	discord.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)
	// uploadArchive("1ad7487f-e1d4-4b15-ba52-100ac34a245e")
	bufferedRead(configuration.EQLogPath, configuration.ReadEntireLog)

	// // Register the messageCreate func as a callback for MessageCreate events.
	// dg.AddHandler(messageCreate)
	discord.AddHandler(reactionAdd)

	// daemon.SdNotify(false, "READY=1")

	// Wait here until CTRL-C or other term signal is received.
	// fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	// l.InfoF("Bot is now running")
	// sc := make(chan os.Signal, 1)
	// signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	// <-sc
}

func reactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	fmt.Printf("Reaction Added! Emoji: %s MessageID: %s", m.Emoji.ID, m.MessageID)
	if m.Emoji.ID == configuration.InvestigationStartEmoji && isArchive(m.MessageID) {
		for _, archiveID := range archives {
			if m.MessageID == archiveID {
				uploadArchive("archive\\" + archiveID + ".json")
			}
		}
	}
}

func bufferedRead(path string, fromStart bool) {
	l := LogInit("bufferedRead-main.go")
	defer l.End()
	file, err := os.Open(path)
	if err != nil {
		l.ErrorF("error opening buffered file: %v", err)
		return
	}
	if !fromStart {
		file.Seek(0, 2) // move to end of file
	}
	bufferedReader := bufio.NewReader(file)
	r, _ := regexp.Compile(configuration.EQBaseLogLine)
	for {
		checkClosedBids()
		str, err := bufferedReader.ReadString('\n')
		if err == io.EOF {
			dumpBids()
			time.Sleep(time.Duration(configuration.LogPollRate) * time.Second) // 1 eq tick = 6 seconds
			continue
		}
		if err != nil {
			l.ErrorF("error opening buffered file: %v", err)
			return
		}

		results := r.FindAllStringSubmatch(str, -1) // this really needs converted to single search
		if results == nil {
			time.Sleep(3 * time.Second)
		} else {
			readLogLine(results)
		}
	}
}

func readLogLine(results [][]string) {
	l := LogInit("readLogLine-main.go")
	defer l.End()
	t := eqTimeConv(results[0][1])
	msg := strings.TrimSuffix(results[0][2], "\r")
	log := &EqLog{
		T:       t,
		Msg:     msg,
		Channel: getChannel(msg),
		Source:  getSource(msg),
	}
	parseLogLine(*log)
}

func eqTimeConv(t string) time.Time {
	l := LogInit("eqTimeConv-main.go")
	defer l.End()
	// Get local time zone
	localT := time.Now()
	zone, _ := localT.Zone()
	// fmt.Println(zone, offset)

	// Parse Time
	cTime, err := time.Parse("Mon Jan 02 15:04:05 2006 MST", t+" "+zone)
	if err != nil {
		l.ErrorF("Error parsing time, defaulting to now: %s\n", err.Error())
		cTime = time.Now()
	}
	return cTime
}

// EqLog represents a single line of eq logging
type EqLog struct {
	T       time.Time `json:"Time"`
	Msg     string    `json:"Message"`
	Channel string    `json:"Channel"`
	Source  string    `json:"Source"`
}

func getChannel(msg string) string {
	l := LogInit("getChannel-main.go")
	defer l.End()
	m := strings.Split(msg, " ")
	if len(m) > 4 {
		if m[3] == "guild," || m[4] == "guild," {
			return "guild"
		}
		if m[3] == "group," || m[4] == "group," {
			return "group"
		}
		if m[3] == "raid," || m[4] == "raid," {
			return "raid"
		}
		if m[1] == "tells" && m[2] == "you," {
			return "tell"
		}
		// fmt.Printf("Default: %s\n", m[2])
		return "system"
	}
	if len(m) > 1 && m[1] == "tells" {
		// return m[3]
		return m[0]
		// return strings.TrimRight(m[3], ",")
	}
	return "system"
}

func getSource(msg string) string {
	l := LogInit("getSource-main.go")
	defer l.End()
	m := strings.Split(msg, " ")
	return m[0]
}

func parseLogLine(log EqLog) {
	l := LogInit("getSource-main.go")
	defer l.End()
	currentTime = log.T
	if log.Channel != "system" && log.Channel != "guild" && log.Channel != "group" && log.Channel != "raid" {
		// fmt.Printf("Channel: %s\n", l.channel)
	}
	if log.Channel == "guild" {
		investigation.addLog(log)
		// Open Bid
		r, _ := regexp.Compile(`'(.+?)x?(\d)?\s+bids\sto\s.+,.+(\d+).*'`) // TODO: Make it NOT match if "CLOSED" or "wins" is in this, otherwise we will open aditional bids - also if we have a dedicated box, we can match that
		result := r.FindStringSubmatch(log.Msg)
		if len(result) > 0 {
			// var bid string
			// var err error
			// bidClean := strings.ReplaceAll(result[2], ",", "")
			// count, err := strconv.Atoi(bidClean)
			// if err != nil {
			// 	l.ErrorF("Error converting count to int: %s", result[2])
			// }
			itemID := isItem(result[1])
			if itemID > 0 { // item numbers are positive
				if result[2] == "" {
					openBid(result[1], 1, itemID)
				} else {
					count, err := strconv.Atoi(result[2])
					if err != nil {
						l.ErrorF("Error converting item count to int: %s", err.Error())
					}
					openBid(result[1], count, itemID)
				}

			}
			return
		}
	}
	if log.Channel == "tell" {
		investigation.addLog(log)
		// Add Bid
		r, _ := regexp.Compile(`'(.+)\s+(\d+).*'`)
		result := r.FindStringSubmatch(log.Msg)
		if len(result) > 0 {
			// var bid string
			// var err error
			bidClean := strings.ReplaceAll(result[2], ",", "")
			bid, err := strconv.Atoi(bidClean)
			if err != nil {
				l.ErrorF("Error converting bid to int: %s", result[2])
			}
			if isItem(result[1]) > 0 && bid >= 10 && isBidOpen(result[1]) { // item names don't get that long
				// addBid(log.source, result[1], bid)
				bids[result[1]].addBid(log.Source, result[1], bid)
			}
			return
		}
	}
}

func openBid(item string, count int, id int) {
	l := LogInit("openBid-main.go")
	defer l.End()
	if _, ok := bids[item]; ok { // if the bid already exist, remove old
		// bids[item] = &BidItem{}
		// delete(bids, item)
		l.InfoF("Ignoring duplicate bid item: %s\n", item)
		return
	}
	l.InfoF("Opening bid for: %s x%d\n", item, count)
	bids[item] = &BidItem{}
	bids[item].Item = item
	bids[item].Count = count
	bids[item].ID = id
	bids[item].URL = configuration.LucyURLPrefix + strconv.Itoa(id)
	bids[item].startBid()
	// lookup item id
}

func isBidOpen(item string) bool {
	l := LogInit("isBidOpen-main.go")
	defer l.End()
	if _, ok := bids[item]; ok {
		return true
	}
	l.WarnF("Bid not open for: %s\n", item)
	return false
}

// Bid defines a bid on an item from a player
type Bid struct {
	Bidder string `json:"Bidder"`
	Amount int    `json:"Amount"`
}

// BidItem defines bids on an item
type BidItem struct {
	Item              string        `json:"Item"`
	ID                int           `json:"ID"`
	URL               string        `json:"URL"`
	Count             int           `json:"Count"`
	Bids              []Bid         `json:"Bids"`
	Start             time.Time     `json:"Start"`
	End               time.Time     `json:"End"`
	Winners           []Winner      `json:"Winners"`
	InvestigationLogs Investigation `json:"InvestigationLogs"`
}

func (b *BidItem) startBid() {
	l := LogInit("startBid-main.go")
	defer l.End()
	b.Start = getTime()
	b.End = b.Start.Add(time.Duration(time.Duration(configuration.BidTimerMinutes) * time.Minute))
	// time.NewTimer(3 * time.Minute)
	_, err := discord.ChannelMessageSend(configuration.LootChannelID, fmt.Sprintf("Bids open on %s x%d for %d minutes", b.Item, b.Count, configuration.BidTimerMinutes))
	if err != nil {
		l.ErrorF("Failed to open bid: %s", err.Error())
	}
}

func (b *BidItem) isBidEnded() bool {
	return getTime().After(b.End)
}

func (b *BidItem) addBid(user string, item string, amount int) {
	l := LogInit("addBid-main.go")
	defer l.End()
	bidLock.Lock()
	defer bidLock.Unlock()

	bid := Bid{
		Bidder: user,
		// item:   item,
		Amount: amount,
	}
	existing := b.bidderExists(user)
	if existing >= 0 {
		b.Bids[existing] = bid
	} else {
		b.Bids = append(b.Bids, bid)
	}
	// TODO: Lookup user's dkp and make sure they can spend (Players can ALWAYS spend 10 dkp)
	if amount%configuration.BidIncrements != 0 {
		l.ErrorF("Bid from %s for %s is not an increment of %d: %d -> Skipping Bid", user, item, configuration.BidIncrements, amount)
		return
	}
	// b.Bids = append(b.Bids, bid)
	l.InfoF("Adding Bid: %#+v\n", bid)
}

func (b *BidItem) bidderExists(user string) int {
	for k, bidder := range b.Bids {
		if bidder.Bidder == user {
			return k
		}
	}
	return -1
}

func (b *BidItem) closeBid() {
	l := LogInit("closeBid-main.go")
	defer l.End()
	// sig := uuid.New()
	l.InfoF("Bid ended for : %s x%d", b.Item, b.Count)
	// Handle Bid winnner
	response := fmt.Sprintf("Winner(s) of %s (%s) x%d", b.Item, b.URL, b.Count)
	// var winAmount int
	for i, winner := range b.getWinners(b.Count) {
		// winAmount = winner.Amount
		if i > 0 {
			response = fmt.Sprintf("%s, and %s", response, winner.Name)
		} else {
			response = fmt.Sprintf("%s for %d is %s", response, winner.Amount, winner.Name)
		}
	}
	response = fmt.Sprintf("%s\n[%s]", response, "Mortimus") // TODO: Pull this name from the log being monitored
	l.InfoF(response)
	var fields []*discordgo.MessageEmbedField
	for _, winner := range b.getWinners(b.Count) {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  winner.Name,
			Value: strconv.Itoa(winner.Amount),
		})
	}
	// var author discordgo.MessageEmbedAuthor
	// author.Name = sig.String()
	// var provider discordgo.MessageEmbedProvider
	// provider.Name = sig.String()
	var footer discordgo.MessageEmbedFooter
	footer.Text = "Mortimus" // TODO: Pull this from the log being monitored
	footer.IconURL = configuration.DiscordLootIcon

	embed := discordgo.MessageEmbed{
		URL:    b.URL,
		Title:  fmt.Sprintf("%s", b.Item),
		Type:   discordgo.EmbedTypeRich,
		Fields: fields,
		Footer: &footer,
	}
	dMsg, err := discord.ChannelMessageSendEmbed(configuration.LootChannelID, &embed)
	// _, err := discord.ChannelMessageSend(configuration.LootChannelID, response)
	if err != nil {
		l.ErrorF("Error sending discord message: %s", err.Error())
	}
	err = discord.MessageReactionAdd(configuration.LootChannelID, dMsg.ChannelID, configuration.InvestigationStartEmoji)
	if err != nil {
		l.ErrorF("Error adding base reaction: %s", err.Error())
	}
	b.InvestigationLogs = investigation
	// Write bid to archive
	writeArchive(dMsg.ID, *b)
}

// Winner is the user who won items
type Winner struct {
	Name   string `json:"Name"`
	Amount int    `json:"Amount"`
}

func (b *BidItem) getWinners(count int) []Winner {
	// Account for no bids (rot)
	for i := 0; i < count+1; i++ {
		fakeBid := Bid{
			Bidder: "Rot",
			Amount: 0,
		}
		b.Bids = append(b.Bids, fakeBid)
	}
	sort.Sort(sort.Reverse(ByBid(b.Bids)))
	winningbid := b.Bids[count].Amount + configuration.BidIncrements
	if b.Bids[0].Bidder != "Rot" && winningbid < configuration.MinimumBid {
		winningbid = configuration.MinimumBid
	}
	if winningbid > b.Bids[0].Amount { // account for ties
		winningbid = b.Bids[0].Amount
	}
	var winners []Winner
	for i := 0; i < count; i++ {
		winner := Winner{
			Name:   b.Bids[i].Bidder,
			Amount: winningbid,
		}
		winners = append(winners, winner)
	}
	b.Winners = winners
	return winners
}

// ByBid is for finding the highest bidders
type ByBid []Bid

func (a ByBid) Len() int           { return len(a) }
func (a ByBid) Less(i, j int) bool { return a[i].Amount < a[j].Amount }
func (a ByBid) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func dumpBids() {
	l := LogInit("dumpBids-main.go")
	defer l.End()
	for i, bid := range bids {
		l.TraceF("%s: %#+v\n", i, bid)
	}
}

func loadItemDB(file string) {
	l := LogInit("loadItemDB-main.go")
	defer l.End()
	csvfile, err := os.Open(file)
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}
	defer csvfile.Close()
	itemDB = make(map[string]int)

	// Parse the file
	r := csv.NewReader(csvfile)
	//r := csv.NewReader(bufio.NewReader(csvfile))

	// Iterate through the records
	headerSkipped := false
	for {
		// Read each record from csv
		record, err := r.Read()
		if !headerSkipped {
			headerSkipped = true
			// skip header line
			continue
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Printf("ID: %s Name: %s\n", record[0], record[1])
		itemID, err := strconv.Atoi(record[0])
		if err != nil {
			log.Fatal(err)
		}
		itemDB[record[1]] = itemID
	}
}

func isItem(name string) int {
	l := LogInit("isItem-main.go")
	defer l.End()
	if _, ok := itemDB[name]; ok {
		return itemDB[name]
	}
	l.WarnF("Cannot find item: %s\n", name)
	return -1
}

func checkClosedBids() {
	l := LogInit("checkClosedBids-main.go")
	defer l.End()
	for k, bi := range bids {
		if bi.isBidEnded() {
			bi.closeBid()
			delete(bids, k) // delete key
		}
	}
}

func writeArchive(name string, data BidItem) {
	l := LogInit("writeArchive-main.go")
	defer l.End()
	l.InfoF("Writing archive %s to file", name)
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		l.ErrorF("Error converting to JSON: %s", err.Error())
	}

	err = ioutil.WriteFile("archive/"+name+".json", file, 0644)
	if err != nil {
		l.ErrorF("Error writing archive to file: %s", err.Error())
	}
	archives = append(archives, name) // add to known archive
}

// Investigation is raw logs during a specified time-frame for verifying failed bids
type Investigation struct {
	Messages []EqLog `json:"Messages"`
}

func (i *Investigation) addLog(l EqLog) {
	i.Messages = append(i.Messages, l)
	if len(i.Messages) > configuration.InvestigationLogLimitMinutes { // remove the oldest log
		copy(i.Messages[0:], i.Messages[1:])
		i.Messages[len(i.Messages)-1] = EqLog{} // or the zero value of T
		i.Messages = i.Messages[:len(i.Messages)-1]
	}
}

func getTime() time.Time {
	if configuration.ReadEntireLog { // We are simulating/testing things, we need to use time from logs
		return currentTime
	}
	return time.Now()
}

func uploadArchive(id string) {
	l := LogInit("uploadArchive-main.go")
	defer l.End()
	file, err := os.Open("archive/" + id + ".json") // TODO: Account for linux, and maliciousness
	if err != nil {
		l.ErrorF("Error finding archive: %s", err.Error())
		discord.ChannelMessageSend(configuration.InvestigationChannelID, "Error uploading investigation: "+id)
	} else {
		discord.ChannelFileSend(configuration.InvestigationChannelID, id+".json", file)
	}
}

func checkReactions() {
	discord.MessageReactions(configuration.LootChannelID, "801309136020570154", ":mag_right:", 100, "", "")
}

func getArchiveList() []string { // TODO: get directory listing on archives
	l := LogInit("getArchiveList-main.go")
	defer l.End()
	var files []string

	root := "./archive"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		name := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
		files = append(files, name)
		return nil
	})
	if err != nil {
		l.ErrorF("Error reading archives: %s", err.Error())
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
