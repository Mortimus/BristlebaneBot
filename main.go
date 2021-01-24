package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

var bidLock sync.Mutex
var bids = map[string]*BidItem{}
var itemDB map[string]int
var itemLock sync.Mutex
var discord *discordgo.Session
var investigation Investigation
var currentTime time.Time // for simulating time
var archives []string     // stores all known archive files for recall
var roster = map[string]*Player{}
var rosterLock sync.Mutex

// var needsReact map[string]*bool
// var needsReactLock sync.Mutex

// srv is the global to connect to google sheets
var srv *sheets.Service

const (
	cInactive = iota
	cAlt
	cRecruit
	cSecondMain
	cMain
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	l := LogInit("getClient-main.go")
	defer l.End()
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	// tokFile := "token.json"
	l.InfoF("Fake loading token from file")
	tok, err := tokenFromFile("")
	if err != nil {
		l.InfoF("Token failed to load, loading from web")
		tok = getTokenFromWeb(config)
		l.InfoF("Saving token")
		saveToken("", tok)
	}
	l.DebugF("Using Token: %+v", tok)
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	l := LogInit("getTokenFromWeb-main.go")
	defer l.End()
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)
	l.InfoF("Requesting user navigate to: %s", authURL)
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		l.FatalF("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		l.FatalF("Unable to retrieve token from web: %v", err)
	}
	l.InfoF("Return token: %+v", tok)
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	l := LogInit("tokenFromFile-main.go")
	defer l.End()
	// f, err := os.Open(file)
	// if err != nil {
	// 	return nil, err
	// }
	// defer f.Close()
	tok := &oauth2.Token{}
	tok.AccessToken = configuration.AccessToken
	tok.Expiry = configuration.Expiry
	tok.RefreshToken = configuration.RefreshToken
	tok.TokenType = configuration.TokenType
	// err = json.NewDecoder(f).Decode(tok)
	l.InfoF("Returning token: %+v", tok)
	return tok, nil
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	l := LogInit("saveToken-main.go")
	defer l.End()
	// fmt.Printf("Saving credential file to: %s\n", path)
	// f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	// if err != nil {
	// 	log.Fatalf("Unable to cache oauth token: %v", err)
	// }
	// defer f.Close()
	// json.NewEncoder(f).Encode(token)
	configuration.AccessToken = token.AccessToken
	configuration.Expiry = token.Expiry
	configuration.RefreshToken = token.RefreshToken
	configuration.TokenType = token.TokenType
	l.InfoF("Saved token to configuration")
	saveConfig()
}

// Inst is an installed struct for google
type Inst struct {
	ClientID                string   `json:"client_id"`
	ProjectID               string   `json:"project_id"`
	AuthURI                 string   `json:"auth_uri"`
	TokenURI                string   `json:"token_uri"`
	AuthProviderx509CertURL string   `json:"auth_provider_x509_cert_url"`
	ClientSecret            string   `json:"client_secret"`
	RedirectURIs            []string `json:"redirect_uris"`
}

// Gtoken is required by google
type Gtoken struct {
	Installed Inst `json:"installed"`
}

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
	itemDB = loaditemDB(configuration.LucyItems)
	archives = getArchiveList()
	// loadRoster(configuration.GuildRosterPath)
	loadRoster(configuration.EQBaseFolder + "/" + getRecentRosterDump(configuration.EQBaseFolder))

	gtoken := &Gtoken{
		Installed: Inst{
			ClientID:                configuration.ClientID,
			ProjectID:               configuration.ProjectID,
			AuthURI:                 configuration.AuthURI,
			TokenURI:                configuration.TokenURI,
			AuthProviderx509CertURL: configuration.AuthProviderx509CertURL,
			ClientSecret:            configuration.ClientSecret,
			RedirectURIs:            configuration.RedirectURIs,
		},
	}
	l.InfoF("Marshalling gToken: %+v", gtoken)
	bToken, err := json.Marshal(gtoken)
	if err != nil {
		l.FatalF("error marshalling gtoken")
	}

	// b, err := ioutil.ReadFile("credentials.json")
	// if err != nil {
	// 	log.Fatalf("Unable to read client secret file: %v", err)
	// }

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(bToken, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		l.FatalF("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err = sheets.New(client)
	if err != nil {
		l.FatalF("Unable retrieve Sheets client: %v", err)
	}

	// Create a new Discord session using the provided bot token.
	discord, err = discordgo.New("Bot " + configuration.DiscordToken)
	if err != nil {
		l.FatalF("Error creating Discord session: %v", err)
	}
	defer discord.Close()
	// discord.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)
	// updateDKP()
	// dumpPlayers()
	// fmt.Println(itemDB["Vyemm's Fang"])
	go bufferedRead(configuration.EQLogPath, configuration.ReadEntireLog)

	// // Register the messageCreate func as a callback for MessageCreate events.
	// dg.AddHandler(messageCreate)
	discord.AddHandler(reactionAdd)

	// Open a websocket connection to Discord and begin listening.
	err = discord.Open()
	if err != nil {
		l.FatalF("Error opening connection with Discord: %v", err)
		return
	}

	// daemon.SdNotify(false, "READY=1")

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	l.InfoF("Bot is now running")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func reactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	l := LogInit("reactionAdd-main.go")
	defer l.End()
	// fmt.Printf("Reaction Added! Emoji: %#+v MessageID: %s\n", m.Emoji, m.MessageID)
	// fmt.Printf("isEmoji: %t isArchive: %t\n", isEmoji, isArchive(m.MessageID))
	if m.Emoji.Name == configuration.InvestigationStartEmoji && isPriviledged(s, m.UserID) && getPrivReactions(s, m.MessageID, configuration.InvestigationStartEmoji) == configuration.InvestigationStartMinReq+1 && isArchive(m.MessageID) {
		// fmt.Printf("Investigating!\n")
		// TODO: Make this check if they are officers
		uploadArchive(m.MessageID)
	}
}

func getPrivReactions(s *discordgo.Session, messageID string, emoji string) int {
	l := LogInit("getReactions-main.go")
	defer l.End()
	msg, err := discord.ChannelMessage(configuration.LootChannelID, messageID)
	if err != nil {
		l.ErrorF("Error getting message: %s", err.Error())
		return -1
	}
	var privEmoji int
	for _, react := range msg.Reactions {
		if react.Emoji.Name == emoji {
			return react.Count
		}
		// TODO: This returns nil sometimes?
		// if react.Emoji.Name == emoji && isPriviledged(s, react.Emoji.User.ID) {
		// 	privEmoji++
		// }
	}
	return privEmoji
	// l.ErrorF("Cannot find emoji")
	// return -1 // Emoji not found
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
			// dumpBids()
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
		// Close Bid
		r, _ := regexp.Compile(configuration.RegexClosedBid) // TODO: Force this to match only the bidmaster
		result := r.FindStringSubmatch(log.Msg)
		if len(result) > 0 {
			itemID := isItem(result[1])
			if itemID > 0 { // item numbers are positive
				if result[2] == "" {
					// openBid(result[1], 1, itemID)
					if _, ok := bids[result[1]]; ok { // Verify bid open, then set end time to start time to close it
						bids[result[1]].End = bids[result[1]].Start // force the bid to show as done
					}
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
		// Open Bid
		r, _ = regexp.Compile(configuration.RegexOpenBid) // TODO: Make it NOT match if "CLOSED" or "wins" is in this, otherwise we will open aditional bids - also if we have a dedicated box, we can match that
		result = r.FindStringSubmatch(log.Msg)
		if len(result) > 0 {
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
		r, _ := regexp.Compile(configuration.RegexTellBid)
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
	updateDKP() // update player dkp with bid
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
	Player Player `json:"Player"`
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
	if !hasEnoughDKP(user, amount) {
		amount = -5
	}
	if _, ok := roster[user]; !ok {
		return
	}

	bid := Bid{
		Bidder: user,
		Player: *roster[user],
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
	winners := b.getWinners(b.Count)
	for i, winner := range winners {
		// winAmount = winner.Amount
		if i > 0 {
			response = fmt.Sprintf("%s, and %s", response, winner.Player.Name)
		} else {
			response = fmt.Sprintf("%s for %d is %s", response, winner.Amount, winner.Player.Name)
		}
	}
	response = fmt.Sprintf("%s\n[%s]", response, getPlayerName(configuration.EQLogPath))
	l.InfoF(response)
	var fields []*discordgo.MessageEmbedField
	for _, winner := range winners {
		displayName := winner.Player.Name
		if winner.Player.Name != winner.Player.Main {
			displayName = fmt.Sprintf("%s (%s)", winner.Player.Name, winner.Player.Main)
		}
		fields = append(fields, &discordgo.MessageEmbedField{
			Name:  displayName,
			Value: strconv.Itoa(winner.Amount),
		})
	}
	// var author discordgo.MessageEmbedAuthor
	// author.Name = sig.String()
	// var provider discordgo.MessageEmbedProvider
	// provider.Name = sig.String()
	var footer discordgo.MessageEmbedFooter
	footer.Text = getPlayerName(configuration.EQLogPath) // TODO: Pull this from the log being monitored
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
	err = discord.MessageReactionAdd(configuration.LootChannelID, dMsg.ID, configuration.InvestigationStartEmoji)
	if err != nil {
		l.ErrorF("Error adding base reaction: %s", err.Error())
	}
	// addReact(dMsg.ID)
	b.InvestigationLogs = investigation
	// Write bid to archive
	writeArchive(dMsg.ID, *b)
}

// Winner is the user who won items
type Winner struct {
	Player Player `json:"Player"`
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
	// sort.Sort(sort.Reverse(ByBid(b.Bids)))
	b.Bids = sortBids(b.Bids)
	// TODO: Account for ties
	winningbid := b.Bids[count].Amount + configuration.BidIncrements
	if b.Bids[0].Bidder != "Rot" && winningbid < configuration.MinimumBid {
		winningbid = configuration.MinimumBid
	}
	if winningbid > b.Bids[0].Amount { // account for ties
		winningbid = b.Bids[0].Amount
		if winningbid == b.Bids[count-1].Amount && b.Bids[count-1].Bidder != "Rot" {
			// A ROLL OFF IS NEEDED
			// Determine AMOUNT of ties
			var ties int
			for _, bidder := range b.Bids {
				if bidder.Amount == winningbid && b.Bids[0].Player.Rank == bidder.Player.Rank { // is tied winner
					ties++
				}
			}
			count = ties // show winners == amount of ties to imply roll off
			discord.ChannelMessageSend(configuration.LootChannelID, "Roll off required!")
		}
	}
	var winners []Winner
	rot := &Player{
		Name:  "Rot",
		DKP:   0,
		Main:  "Rot",
		Level: 0,
		Class: "Necromancer",
		Rank:  cInactive,
		Alt:   true,
	}
	for i := 0; i < count; i++ {
		var win Player
		if _, ok := roster[b.Bids[i].Bidder]; ok {
			win = *roster[b.Bids[i].Bidder]
		} else {
			win = *rot
		}
		winner := Winner{
			Player: win,
			Amount: winningbid,
		}
		winners = append(winners, winner)
	}
	b.Winners = winners
	return winners
}

func sortBids(bids []Bid) []Bid {
	var mains []Bid
	var secondmains []Bid
	var recruits []Bid
	var alts []Bid
	var inactives []Bid

	for _, bid := range bids {
		switch bid.Player.Rank {
		case cMain:
			mains = append(mains, bid)
		case cSecondMain:
			secondmains = append(secondmains, bid)
		case cRecruit:
			recruits = append(recruits, bid)
		case cAlt:
			alts = append(alts, bid)
		case cInactive:
			inactives = append(inactives, bid)
		}
	}
	sort.Sort(sort.Reverse(ByBid(mains)))
	sort.Sort(sort.Reverse(ByBid(secondmains)))
	sort.Sort(sort.Reverse(ByBid(recruits)))
	sort.Sort(sort.Reverse(ByBid(alts)))
	sort.Sort(sort.Reverse(ByBid(inactives)))
	var winners []Bid
	winners = append(winners, mains...)
	winners = append(winners, secondmains...)
	winners = append(winners, recruits...)
	winners = append(winners, alts...)
	winners = append(winners, inactives...)
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

func loaditemDB(file string) map[string]int {
	l := LogInit("loaditemDB-main.go")
	defer l.End()
	csvfile, err := os.Open(file)
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}
	defer csvfile.Close()
	itemDB := make(map[string]int)

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
		// fmt.Printf("Item: %s ID: %d\n", record[1], itemID)
		// fmt.Println(itemDB[record[1]])
	}
	l.InfoF("Loaded itemDB")
	return itemDB
}

func isItem(name string) int {
	l := LogInit("isItem-main.go")
	defer l.End()
	itemLock.Lock()
	defer itemLock.Unlock()
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

// func checkReactions() {
// 	discord.MessageReactions(configuration.LootChannelID, "801309136020570154", ":mag_right:", 100, "", "")
// }

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

func loadRoster(file string) {
	l := LogInit("loadRoster-main.go")
	defer l.End()
	csvfile, err := os.Open(file)
	if err != nil {
		log.Fatalln("Couldn't open the roster csv file", err)
	}
	defer csvfile.Close()

	// Parse the file
	r := csv.NewReader(csvfile)
	r.Comma = '\t'
	//r := csv.NewReader(bufio.NewReader(csvfile))

	// Iterate through the records
	headerSkipped := true // guild dumps have no header
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
			l.ErrorF("Error loading roster: %s", err.Error())
		}
		var player Player
		player.Name = record[0]
		// fmt.Printf("ID: %s Name: %s\n", record[0], record[1])
		player.Level, err = strconv.Atoi(record[1])
		if err != nil {
			log.Fatal(err)
		}
		player.Class = record[2]
		if record[4] == "A" {
			player.Alt = true
			player.Rank = cAlt // Default to alt for now
			// Figure out if secondmain, alt, recruit
			// Figure out alt's main
			r, _ := regexp.Compile(configuration.RegexIsAlt) // TODO: Make it NOT match if "CLOSED" or "wins" is in this, otherwise we will open aditional bids - also if we have a dedicated box, we can match that
			Altresult := r.FindStringSubmatch(record[7])
			if len(Altresult) > 0 {
				player.Main = Altresult[1]
			}
			// Figure out secondmain and it's main
			r, _ = regexp.Compile(configuration.RegexIs2ndMain) // TODO: Make it NOT match if "CLOSED" or "wins" is in this, otherwise we will open aditional bids - also if we have a dedicated box, we can match that
			SecondMainResult := r.FindStringSubmatch(record[7])
			if len(SecondMainResult) > 0 {
				player.Main = SecondMainResult[1]
				player.Rank = cSecondMain
			}
		} // defaults to false
		if !player.Alt && isRaider(record[3]) { // not an alt and a raider, so is a main
			player.Rank = cMain
			player.Main = player.Name
		}
		if record[3] == "Recruit" {
			player.Rank = cRecruit
			player.Main = player.Name
		}
		if record[3] == "Inactive" {
			player.Rank = cInactive
			player.Main = player.Name
		}
		roster[player.Name] = &player
	}
	l.InfoF("Loaded Roster")
}

func dumpPlayers() {
	for _, p := range roster {
		fmt.Printf("%#+v\n", p)
	}
}

// Player is represented by Name, Level, Class, Rank, Alt, Last On, Zone, Note, Tribute Status, Unk_1, Unk_2, Last Donation, Private Note
type Player struct {
	Name  string `json:"Name"`
	Main  string `json:"Main"` // Name of player's main
	Level int    `json:"Level"`
	Class string `json:"Class"`
	Rank  int    `json:"Rank"` // this is a meta field, its not direct from the rank column
	Alt   bool   `json:"Alt"`
	DKP   int    `json:"DKP"` // this is filled in post from google sheets
}

func isRaider(rank string) bool {
	for _, r := range configuration.GuildRaidingRoles {
		if r == rank {
			return true
		}
	}
	return false
}

func updatePlayerDKP(name string, dkp int) {
	l := LogInit("updatePlayerDKP-commands.go")
	defer l.End()
	rosterLock.Lock()
	defer rosterLock.Unlock()
	if _, ok := roster[name]; ok {
		roster[name].DKP = dkp
		return
	}
	l.ErrorF("Cannot find player to update DKP: %s giving them 0 dkp", name)
	DiscordF("Error configuring %s's DKP, are they on the DKP sheet, Roster Dump, and are the Guild Notes correct?", name)
	roster[name] = &Player{
		Name:  name,
		Main:  name,
		DKP:   0,
		Level: 0,
		Class: "Unknown",
		Rank:  cInactive,
		Alt:   false,
	}
}

func updateDKP() {
	l := LogInit("lookupPlayer-commands.go")
	defer l.End()
	spreadsheetID := configuration.DKPSheetURL
	readRange := configuration.DKPSummarySheetName
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		l.ErrorF("Unable to retrieve data from sheet: %v", err)
		return
	}

	if len(resp.Values) == 0 {
		l.ErrorF("Cannot read dkp sheet: %v", resp)
		// log.Println("No data found.")
	} else {
		// var lastClass string
		for _, row := range resp.Values {
			// if row[0] == "Necromancer" {
			// 	fmt.Printf("%s: %s\n", row[2], row[6])
			// }
			// l.TraceF("Player: %s Target: %s", row[configuration.DKPSRosterSheetPlayerCol], strings.TrimSpace(tar))
			name := fmt.Sprintf("%s", row[configuration.DKPSummarySheetPlayerCol])
			name = strings.TrimSpace(name)
			if name != "" {
				sDKP := fmt.Sprintf("%s", row[configuration.DKPSummarySheetDKPCol])
				sDKP = strings.ReplaceAll(sDKP, ",", "")
				dkp, err := strconv.Atoi(sDKP)
				if err != nil {
					l.ErrorF("Error converting DKP to int: %s", err.Error())
					continue
				}
				updatePlayerDKP(name, dkp)
			}
		}
	}
}

func hasEnoughDKP(name string, amount int) bool {
	l := LogInit("hasEnoughDKP-commands.go")
	defer l.End()
	rosterLock.Lock()
	defer rosterLock.Unlock()
	var bHasDKP bool
	if _, ok := roster[name]; ok {
		bHasDKP = true
	}
	if amount < 10 || (bHasDKP && roster[name].DKP >= amount) { // You can always spend 10dkp
		return true
	}
	l.WarnF("%s does not have %d DKP but tried to spend it", name, amount)
	return false
}

func isPriviledged(s *discordgo.Session, userID string) bool {
	// TODO: Fix this
	return true
	l := LogInit("isPriviledged-main.go")
	defer l.End()
	guildID := configuration.DiscordGuildID
	l.InfoF("GuildID: %+v\nUserID: %+v", guildID, userID)
	member, err := s.State.Member(guildID, userID)
	if err != nil {
		// if member, err = s.GuildMember(guildID, userID); err != nil {
		// 	return false, err
		// }
		l.ErrorF("Error: %s", err.Error())
	}
	l.InfoF("Member: %+v", member)
	for _, roleID := range member.Roles {
		role, err := s.State.Role(guildID, roleID)
		if err != nil {
			l.ErrorF("Error: %s", err.Error())
			return false
		}
		for _, cRole := range configuration.DiscordPrivRoles {
			l.InfoF("Crole: %v vs role.Name: %v", cRole, role.Name)
			if cRole == role.Name {
				l.InfoF("Role found, authorizing: %s == %s", cRole, role.Name)
				return true
			}
		}
	}
	return false
}

// func addReact(msgID string) {
// 	needsReactLock.Unlock()
// 	defer needsReactLock.Lock()
// 	*needsReact[msgID] = true
// }

// func removeReact(msgID string) {
// 	needsReactLock.Unlock()
// 	defer needsReactLock.Lock()
// 	delete(needsReact, msgID)
// }

// func checkReacts() {
// 	l := LogInit("checkReacts-commands.go")
// 	defer l.End()
// 	needsReactLock.Unlock()
// 	defer needsReactLock.Lock()
// 	for k := range needsReact {
// 		err := discord.MessageReactionAdd(configuration.LootChannelID, k, configuration.InvestigationStartEmoji)
// 		if err != nil {
// 			l.ErrorF("Error adding base reaction: %s", err.Error())
// 		} else {
// 			removeReact(k)
// 		}
// 	}
// }

func getPlayerName(logFile string) string {
	// l := LogInit("getPlayerName-commands.go")
	// defer l.End()
	logFile = filepath.Base(logFile)
	extension := filepath.Ext(logFile)
	name := logFile[0 : len(logFile)-len(extension)]
	split := strings.Split(name, "_")
	// fmt.Printf("LogFile: %s\nExtension: %s\nName: %s\nSplit: %#+v\n", logFile, extension, name, split)
	if len(split) < 3 {
		return "Unknown Player"
	}
	return split[2] + "." + split[1]
}

// DiscordF provides a printf to a discord channel
func DiscordF(format string, v ...interface{}) {
	l := LogInit("DiscordF-commands.go")
	defer l.End()
	msg := fmt.Sprintf(format, v...)
	_, err := discord.ChannelMessageSend(configuration.InvestigationChannelID, msg)
	if err != nil {
		l.ErrorF("Failed to send to discord: %s", err.Error())
	}
}

func getRecentRosterDump(path string) string {
	l := LogInit("getRecentRosterDump-commands.go")
	defer l.End()
	var files []string

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if strings.HasPrefix(filepath.Base(path), configuration.GuildName) {
			files = append(files, filepath.Base(path))
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	// for _, file := range files {
	// 	fmt.Println(file)
	// }
	return files[len(files)-1] // return last file - should be latest
	// It looks like files are already sorted by date, we don't need this
	// var times []time.Time
	// for _, file := range files {
	// 	// Remove extension
	// 	file := strings.TrimSuffix(file, filepath.Ext(file))
	// 	spltFile := strings.Split(file, "-")
	// 	if len(spltFile) > 2 { // should always happen
	// 		timeString := spltFile[1] + "-" + spltFile[2] // only parse the time
	// 		t, err := time.Parse("20060102-150405", timeString)
	// 		if err != nil {
	// 			l.ErrorF("Error parsing time of roster dump: %s", err.Error())
	// 		}
	// 		times = append(times, t)
	// 	}
	// }
	// return ""
}

// // ByTime is for finding the most recent item
// type ByTime []time.Time

// func (a ByTime) Len() int           { return len(a) }
// func (a ByTime) Less(i, j int) bool { return a[i].Before(a[j]) }
// func (a ByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
