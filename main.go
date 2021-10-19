package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	everquest "github.com/Mortimus/goEverquest"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

// var bidLock sync.Mutex
// var bids = map[string]*BidItem{}

var printChan = make(chan string)

var itemDB everquest.ItemDB
var spellDB everquest.SpellDB

var investigation Investigation
var currentTime time.Time // for simulating time
var archives []string     // stores all known archive files for recall
// var roster = map[string]*Player{}
// var rosterLock sync.Mutex

// var ChatLogs chan everquest.EqLog

// var currentZone string

// var raidDumps int

// var raidStart time.Time

// var nextDump time.Time
// var needsDump bool

var Debug, Warn, Err, Info *log.Logger

// const (
// 	cInactive = iota
// 	cAlt
// 	cRecruit
// 	cSecondMain
// 	cMain
// )

func init() {
	// Initialize log handlers
	LogFile, err := os.OpenFile(configuration.Log.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	Warn = log.New(LogFile, "[WARN] ", log.Lshortfile|log.Ldate|log.Ltime|log.LUTC|log.Lmsgprefix)
	Err = log.New(LogFile, "[ERR] ", log.Lshortfile|log.Ldate|log.Ltime|log.LUTC|log.Lmsgprefix)
	Info = log.New(LogFile, "[INFO] ", log.Lshortfile|log.Ldate|log.Ltime|log.LUTC|log.Lmsgprefix)
	Debug = log.New(LogFile, "[DEBUG] ", log.Lshortfile|log.Ldate|log.Ltime|log.LUTC|log.Lmsgprefix)
	if configuration.Log.Level < 0 {
		Warn.SetOutput(ioutil.Discard)
	}
	if configuration.Log.Level < 1 {
		Err.SetOutput(ioutil.Discard)
	}
	if configuration.Log.Level < 2 {
		Info.SetOutput(ioutil.Discard)
	}
	if configuration.Log.Level < 3 {
		Debug.SetOutput(ioutil.Discard)
	}
	itemDB.LoadFromFile(configuration.Everquest.ItemDB, Err, Info)
	spellDB.LoadFromFile(configuration.Everquest.SpellDB, Err)

	archives = getArchiveList()
	// loadRoster(configuration.GuildRosterPath)

	// Setup google sheets
	gtoken := &Gtoken{
		Installed: Inst{
			ClientID:                configuration.Google.ClientID,
			ProjectID:               configuration.Google.ProjectID,
			AuthURI:                 configuration.Google.AuthURI,
			TokenURI:                configuration.Google.TokenURI,
			AuthProviderx509CertURL: configuration.Google.AuthProviderx509CertURL,
			ClientSecret:            configuration.Google.ClientSecret,
			RedirectURIs:            configuration.Google.RedirectURIs,
		},
	}
	Info.Printf("Marshalling gToken: %+v", gtoken)
	bToken, err := json.Marshal(gtoken)
	if err != nil {
		Err.Fatalf("error marshalling gtoken")
	}

	// b, err := ioutil.ReadFile("credentials.json")
	// if err != nil {
	// 	log.Fatalf("Unable to read client secret file: %v", err)
	// }

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(bToken, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		Err.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err = sheets.New(client)
	if err != nil {
		Err.Fatalf("Unable retrieve Sheets client: %v", err)
	}
	seedBosses()
}

func main() {
	var err error
	// Create a new Discord session using the provided bot token.
	discord, err = discordgo.New("Bot " + configuration.Discord.Token)
	if err != nil {
		Err.Fatalf("Error creating Discord session: %v", err)
	}
	defer discord.Close()
	discord.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsAll)

	// Load the roaster for character lookups, and rank determination
	// loadRoster(configuration.Everquest.BaseFolder + "/" + getRecentRosterDump(configuration.Everquest.BaseFolder)) // needs to run AFTER discord is initialized
	// Create channel for chat logs
	ChatLogs := make(chan everquest.EqLog)
	// Create channel to allow log reading to gracefully stop
	Quit := make(chan bool, 1)

	// Print all the plugins versions/etc
	printPlugins(Info.Writer())
	// Read logs on dedicated thread
	go everquest.BufferedLogRead(configuration.Everquest.LogPath, configuration.Main.ReadEntireLog, configuration.Main.LogPollRate, ChatLogs, Quit)
	// Parse logs on dedicated thread
	go parseLogs(ChatLogs, Quit)

	// Add handler so we can monitor reaction to messages
	discord.AddHandler(reactionAdd)

	// Open a websocket connection to Discord and begin listening.
	err = discord.Open()
	if err != nil {
		Err.Fatalf("Error opening connection with Discord: %v", err)
		return
	}

	// daemon.SdNotify(false, "READY=1")

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	Info.Printf("Bot is now running")
	path, err := everquest.GetRecentRosterDump(configuration.Everquest.BaseFolder, configuration.Everquest.GuildName)
	if err != nil {
		fmt.Printf("Error finding roster dump: %s", err.Error())
	} else {
		if isDumpOutOfDate(strings.TrimSuffix(filepath.Base(path), filepath.Ext(filepath.Base(path)))) {
			DiscordF(configuration.Discord.InvestigationChannelID, "**Roster dump is out of date, please update!**")
		}
	}

	DiscordF(configuration.Discord.InvestigationChannelID, "**BidBot online**\n> Secondmains bid as mains: %t\n> Secondmain max bid: %d (0 means infinite)", configuration.Bids.SecondMainsBidAsMains, configuration.Bids.SecondMainAsMainMaxBid)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	for {
		select {
		case <-sc:
			return
		case pMsg := <-printChan:
			fmt.Printf("%s", pMsg)
		}
	}
	// <-sc
	// Quit <- true
}

// func printHUD() {
// 	fmt.Print("\033[H\033[2J") // clear the terminal should only work in windows
// 	fmt.Printf("Time: %s\n", currentTime.String())
// 	fmt.Printf("Player: %s\tZone: %s\n", getPlayerName(configuration.Everquest.LogPath), currentZone)
// 	fmt.Printf("Guild Members: %d\n", len(Roster))
// 	fmt.Printf("SecondMainsBidAsMains: %t\tSecondMainMaxBidAsMain: %d\n", configuration.Bids.SecondMainsBidAsMains, configuration.Bids.SecondMainAsMainMaxBid)
// 	fmt.Printf("Waiting to be looted: %d\n", len(needsLooted))
// 	for num, item := range needsLooted {
// 		fmt.Printf("#%d: %s\n", num+1, item)
// 	}
// }

func parseLogs(ChatLogs chan everquest.EqLog, quit <-chan bool) {
	Info.Printf("Parsing logs")
	// printHUD()
	for msgs := range ChatLogs {
		currentTime = msgs.T
		if (msgs.Channel == "guild" && msgs.Source == "You") || msgs.Channel == "tell" {
			investigation.addLog(msgs)
			// printMessage(&msgs)
		}
		if msgs.Channel == "tell" || (msgs.Source == "You" && strings.Contains(msgs.Msg, "told")) {
			printMessage(&msgs)
		}
		//checkClosedBids()
		//parseLogLine(msgs) // Old, should be replaced with plugin system below
		for _, handler := range Handlers {
			switch handler.OutputChannel() {
			case STDOUT:
				handler.Handle(&msgs, os.Stdout)
			case BIDOUT:
				handler.Handle(&msgs, &BidWriter)
			case INVESTIGATEOUT:
				handler.Handle(&msgs, &InvestigateWriter)
			case RAIDOUT:
				handler.Handle(&msgs, &RaidWriter)
			case SPELLOUT:
				handler.Handle(&msgs, &SpellWriter)
			case FLAGOUT:
				handler.Handle(&msgs, &FlagWriter)
			case PARSEOUT:
				handler.Handle(&msgs, &ParseWriter)
			default:
				handler.Handle(&msgs, os.Stdout)
			}
		}
		select {
		case <-quit:
			return
		default:
		}
	}
}

// func parseLogLine(log everquest.EqLog) {
// 	// if !needsDump && getTime().Round(5*time.Minute) == nextDump.Round(5*time.Minute) {
// 	// 	DiscordF(configuration.Discord.InvestigationChannelID, "Time for another hourly raid dump!")
// 	// 	needsDump = true
// 	// }
// 	if log.Channel == "guild" {
// 		investigation.addLog(log)
// 		// Close Bid
// 		r, _ := regexp.Compile(configuration.Bids.RegexClosedBid) // TODO: Force this to match only the bidmaster
// 		result := r.FindStringSubmatch(log.Msg)
// 		if len(result) > 0 {
// 			if log.Source != "You" {
// 				return
// 			}
// 			itemName := strings.TrimSpace(result[1])
// 			itemName = strings.ToLower(itemName)
// 			itemID := isItem(itemName)
// 			if itemID > 0 { // item numbers are positive
// 				if _, ok := bids[itemName]; ok { // Verify bid open, then set end time to start time to close it
// 					bids[itemName].End = bids[itemName].Start // force the bid to show as done
// 				}
// 			} else {
// 				Warn.Printf("Cannot find item %s has id %d\n", itemName, itemID)
// 				DiscordF(configuration.Discord.InvestigationChannelID, "**[ERROR] Cannot find item %s has id %d, for closing, please retry.**\n", itemName, itemID)
// 			}
// 			return
// 		}
// 		// Open Bid
// 		r, _ = regexp.Compile(configuration.Bids.RegexOpenBid) // TODO: Make it NOT match if "CLOSED" or "wins" is in this, otherwise we will open aditional bids - also if we have a dedicated box, we can match that
// 		result = r.FindStringSubmatch(log.Msg)
// 		if len(result) > 0 {
// 			if log.Source != "You" {
// 				return
// 			}
// 			itemName := strings.TrimSpace(result[1])
// 			itemName = strings.ToLower(itemName)
// 			itemID := isItem(itemName)
// 			if itemID > 0 { // item numbers are positive
// 				if result[2] == "" {
// 					openBid(itemName, 1, itemID)
// 				} else {
// 					count, err := strconv.Atoi(result[2][1:])
// 					if err != nil {
// 						Err.Printf("Error converting item count to int: %s", err.Error())
// 					}
// 					openBid(itemName, count, itemID)
// 				}
// 			} else {
// 				Warn.Printf("Cannot find item %s has id %d\n", itemName, itemID)
// 				DiscordF(configuration.Discord.InvestigationChannelID, "**[ERROR] Cannot find item %s has id %d, bids will NOT be recorded. Item list needs updated and someone needs to manually take bids.**\n", itemName, itemID)
// 			}
// 			return
// 		}
// 	}
// 	// if log.Channel == "say" { // Guzz says, 'Hail, a spell jammer'
// 	// 	if checkFlagGivers(log.Msg) {
// 	// 		DiscordF(configuration.Discord.FlagChannelID, "%s got the flag from %s\n", log.Source, currentZone)
// 	// 	}
// 	// }

// 	// if strings.Contains(strings.ToLower(log.Channel), strings.ToLower(configuration.Everquest.ParseChannel)) {
// 	// 	if strings.Contains(log.Msg, configuration.Everquest.ParseIdentifier) {
// 	// 		i := strings.Index(log.Msg, "'")
// 	// 		out := strings.ReplaceAll(log.Msg[i:], "'", "")
// 	// 		DiscordF(configuration.Discord.ParseChannelID, "> %s provided a parse\n```%s```", log.Source, out)
// 	// 	}
// 	// }
// 	if log.Channel == "tell" {
// 		investigation.addLog(log)
// 		// Add Bid
// 		r, _ := regexp.Compile(configuration.Bids.RegexTellBid)
// 		result := r.FindStringSubmatch(log.Msg)
// 		if len(result) > 0 {
// 			// var bid string
// 			// var err error
// 			bidClean := strings.ReplaceAll(result[2], ",", "")
// 			bid, err := strconv.Atoi(bidClean)
// 			if err != nil {
// 				Err.Printf("Error converting bid to int: %s", result[2])
// 			}
// 			itemName := strings.TrimSpace(result[1])
// 			itemName = strings.ToLower(itemName)
// 			if isItem(strings.TrimSpace(itemName)) > 0 && isBidOpen(strings.TrimSpace(itemName)) { // item names don't get that long
// 				if bid == 0 && bids[strings.TrimSpace(itemName)].bidderExists(log.Source) != -1 { // Bidder wants to cancel
// 					Info.Printf("%s is trying to remove their bid\n", log.Source)
// 					bids[strings.TrimSpace(itemName)].Bids = removeFromSlice(bids[strings.TrimSpace(itemName)].Bids, bids[strings.TrimSpace(itemName)].bidderExists(log.Source))
// 				}
// 				if bid >= 10 {
// 					bids[strings.TrimSpace(itemName)].addBid(log.Source, strings.TrimSpace(itemName), bid)
// 				}
// 			}
// 			return
// 		}
// 	}
// 	// if log.Channel == "system" {
// 	// 	// if strings.Contains(log.Msg, "Outputfile") {
// 	// 	// 	outputName := log.Msg[21:] // Filename Outputfile sent data to
// 	// 	// 	// if strings.Contains(log.Msg, "RaidRoster") {
// 	// 	// 	// 	Info.Printf("Raid Dump exported, uploading")
// 	// 	// 	// 	// upload to discord
// 	// 	// 	// 	uploadRaidDump(outputName)
// 	// 	// 	// }
// 	// 	// 	if strings.Contains(log.Msg, configuration.Everquest.GuildName) {
// 	// 	// 		Info.Printf("Guild Dump exported, uploading")
// 	// 	// 		// upload to discord
// 	// 	// 		uploadGuildDump(outputName)
// 	// 	// 	}
// 	// 	// }
// 	// 	// if strings.Contains(log.Msg, "You have entered ") && !strings.Contains(log.Msg, "function.") { // You have entered Vex Thal. NOT You have entered an area where levitation effects do not function.
// 	// 	// 	currentZone = log.Msg[17 : len(log.Msg)-1]
// 	// 	// 	printHUD()
// 	// 	// 	Info.Printf("Changing zone to %s\n", currentZone)
// 	// 	// }
// 	// 	// Item Looted
// 	// 	r, _ := regexp.Compile(configuration.Everquest.RegexLoot)
// 	// 	result := r.FindStringSubmatch(log.Msg)
// 	// 	if len(result) > 0 {
// 	// 		if strings.Contains(result[2], "Spell: ") || strings.Contains(result[2], "Ancient: ") { // TODO: Include "Ancient: "
// 	// 			// TODO: Lookup who needs the spell and add it to the loot message
// 	// 			cleanSpellName := strings.Replace(result[2], "Spell: ", "", 1)
// 	// 			cleanSpellName = strings.Replace(cleanSpellName, "Ancient: ", "Ancient ", 1)
// 	// 			spellID, _ := spellDB.FindIDByName(cleanSpellName)
// 	// 			spell, _ := spellDB.GetSpellByID(spellID)
// 	// 			var notNecro bool
// 	// 			for _, classCanUse := range spell.GetClasses() {
// 	// 				if classCanUse == "Necromancer" {
// 	// 					var canUseString string
// 	// 					for i, player := range findWhoNeedsSpell(spell) {
// 	// 						if i != 0 {
// 	// 							canUseString += ", "
// 	// 						}
// 	// 						canUseString += player
// 	// 					}
// 	// 					DiscordF(configuration.Discord.SpellDumpChannelID, "%s looted %s from %s needed by %+s", result[1], spell.Name, result[3], canUseString)
// 	// 				} else {
// 	// 					notNecro = true
// 	// 				}
// 	// 			}
// 	// 			if notNecro {
// 	// 				DiscordF(configuration.Discord.SpellDumpChannelID, "%s looted %s from %s usable by %s", result[1], spell.Name, result[3], spell.Classes)
// 	// 			}
// 	// 			if len(spell.GetClasses()) == 0 {
// 	// 				// TODO: do a broader search for a spell with said name that has classes
// 	// 				DiscordF(configuration.Discord.SpellDumpChannelID, "%s looted %s from %s usable by %s", result[1], spell.Name, result[3], spell.Classes)
// 	// 			}
// 	// 		}
// 	// 		if isSpellProvider(result[2]) {
// 	// 			// TODO: Lookup what class will get this and add it to the loot message
// 	// 			DiscordF(configuration.Discord.SpellDumpChannelID, "%s looted %s from %s", result[1], result[2], result[3])
// 	// 		}
// 	// 		for _, item := range needsLooted { // Notify that someone looted a bid upon item
// 	// 			if strings.ToLower(result[2]) == item {
// 	// 				DiscordF(configuration.Discord.LootChannelID, "%s looted %s from %s", result[1], result[2], result[3])
// 	// 				removeLootFromLooted(item)
// 	// 				break // We only want to remove 1 item per loot (multi bid items we want to see all winners loot them)
// 	// 			}
// 	// 		}
// 	// 	}
// 	// }
// }

// func removeFromSlice(slice []Bid, s int) []Bid {
// 	return append(slice[:s], slice[s+1:]...)
// }

// func checkFlagGivers(msg string) bool {
// 	for _, flaggiver := range configuration.Everquest.FlagGiver {
// 		if strings.Contains(msg, flaggiver) {
// 			return true
// 		}
// 	}
// 	return false
// }

// func openBid(item string, count int, id int) {
// 	if _, ok := bids[item]; ok { // if the bid already exist, remove old
// 		// bids[item] = &BidItem{}
// 		// delete(bids, item)
// 		Info.Printf("Ignoring duplicate bid item: %s\n", item)
// 		return
// 	}
// 	Info.Printf("Opening bid for: %s x%d\n", item, count)
// 	updateDKP() // update player dkp with bid
// 	bids[item] = &BidItem{}
// 	bids[item].SecondMainsBidAsMains = configuration.Bids.SecondMainsBidAsMains   // For investigation purposes and debug replay
// 	bids[item].SecondMainAsMainMaxBid = configuration.Bids.SecondMainAsMainMaxBid // For investigation purposes and debug replay
// 	bids[item].Zone = currentZone
// 	bids[item].Item = item
// 	bids[item].Count = count
// 	bids[item].ID = id
// 	bids[item].URL = configuration.Main.LucyURLPrefix + strconv.Itoa(id)
// 	bids[item].startBid()
// 	printHUD()
// 	// lookup item id
// }

// func isBidOpen(item string) bool {
// 	if _, ok := bids[item]; ok {
// 		return true
// 	}
// 	Warn.Printf("Bid not open for: %s\n", item)
// 	return false
// }

// Bid defines a bid on an item from a player
// type Bid struct {
// 	Bidder string `json:"Bidder"`
// 	Player Player `json:"Player"`
// 	Amount int    `json:"Amount"`
// }

// BidItem defines bids on an item
// type BidItem struct {
// 	Item                   string        `json:"Item"`
// 	ID                     int           `json:"ID"`
// 	URL                    string        `json:"URL"`
// 	Count                  int           `json:"Count"`
// 	Bids                   []Bid         `json:"Bids"`
// 	Start                  time.Time     `json:"Start"`
// 	End                    time.Time     `json:"End"`
// 	Winners                []Winner      `json:"Winners"`
// 	InvestigationLogs      Investigation `json:"InvestigationLogs"`
// 	SecondMainsBidAsMains  bool          `json:"SecondMainsBidAsMains"`
// 	SecondMainAsMainMaxBid int           `json:"SecondMainAsMainMaxBid"`
// 	RollOff                bool          `json:"RollOff"`
// 	RollOffWinner          string        `json:"RollOffWinner"` // Added post win
// 	Zone                   string        `json:"Zone"`
// }

// func (b *BidItem) startBid() {
// 	b.Start = getTime()
// 	b.End = b.Start.Add(time.Duration(time.Duration(configuration.Bids.OpenBidTimer) * time.Minute))
// 	// time.NewTimer(3 * time.Minute)
// 	id, _ := itemDB.FindIDByName(b.Item)
// 	item, _ := itemDB.GetItemByID(id)
// 	_, err := discord.ChannelMessageSend(configuration.Discord.LootChannelID, fmt.Sprintf("> Bids open on **%s** x%d for %d minutes\n```%s```\n", item.Name, b.Count, configuration.Bids.OpenBidTimer, getItemDesc(item)))
// 	if err != nil {
// 		Err.Printf("Failed to open bid: %s", err.Error())
// 	}
// }

// func (b *BidItem) isBidEnded() bool {
// 	return getTime().After(b.End)
// }

// func (b *BidItem) addBid(user string, item string, amount int) {
// 	bidLock.Lock()
// 	defer bidLock.Unlock()
// 	if _, ok := roster[user]; !ok { // user doesn't exist, no bids to add
// 		Err.Printf("User %s does not exist, refusing bid", user)
// 		return
// 	}
// 	// TODO: Lookup user's dkp and make sure they can spend (Players can ALWAYS spend 10 dkp)
// 	if amount != 0 && amount%configuration.Bids.Increments != 0 {
// 		Err.Printf("Bid from %s for %s is not an increment of %d: %d -> Rounding down", user, item, configuration.Bids.Increments, amount)
// 		amount = roundDown(amount)
// 	}
// 	if !hasEnoughDKP(roster[user].Main, amount) { // this works but roster should have their dkp
// 		amount = getMaxDKP(roster[user].Main)
// 	}

// 	bid := Bid{
// 		Bidder: user,
// 		Player: *roster[user],
// 		Amount: amount,
// 	}
// 	existing := b.bidderExists(user)
// 	if existing >= 0 {
// 		b.Bids[existing] = bid
// 	} else {
// 		b.Bids = append(b.Bids, bid)
// 	}
// 	itemID, _ := itemDB.FindIDByName(b.Item)
// 	itemInstance, _ := itemDB.GetItemByID(itemID)
// 	if !canUse(itemInstance, *roster[user]) && len(itemInstance.GetClasses()) > 0 && amount > 0 {
// 		DiscordF(configuration.Discord.InvestigationChannelID, "A class that is not %s bid on %s\n", itemInstance.GetClasses(), itemInstance.Name)
// 	}

// 	// b.Bids = append(b.Bids, bid)
// 	Info.Printf("Adding Bid: Player: %s Main: %s MaxDKP: %d\n", bid.Player.Name, bid.Player.Main, bid.Player.DKP)
// }

// func canUse(item everquest.Item, player Player) bool {
// 	classes := item.GetClasses()
// 	for _, class := range classes {
// 		if class == player.Class {
// 			return true
// 		}
// 	}
// 	return false
// }

// func (b *BidItem) bidderExists(user string) int {
// 	Info.Printf("Checking if bidder %s exists in %#+v", user, b)
// 	for k, bidder := range b.Bids {
// 		if bidder.Bidder == user {
// 			return k
// 		}
// 	}
// 	return -1
// }

// func (b *BidItem) closeBid() {
// 	// sig := uuid.New()
// 	Info.Printf("Bid ended for : %s x%d", b.Item, b.Count)
// 	// Handle Bid winnner
// 	response := fmt.Sprintf("Winner(s) of %s (%s) x%d", b.Item, b.URL, b.Count)
// 	// var winAmount int
// 	winners := b.getWinners(b.Count)
// 	for i, winner := range winners {
// 		// winAmount = winner.Amount
// 		if i > 0 {
// 			response = fmt.Sprintf("%s, and %s", response, winner.Player.Name)
// 		} else {
// 			response = fmt.Sprintf("%s for %d is %s", response, winner.Amount, winner.Player.Name)
// 		}
// 	}
// 	response = fmt.Sprintf("%s\n[%s]", response, getPlayerName(configuration.Everquest.LogPath))
// 	Info.Printf(response)
// 	var fields []*discordgo.MessageEmbedField
// 	for _, winner := range winners {
// 		displayName := winner.Player.Name
// 		if winner.Player.Name != winner.Player.Main {
// 			displayName = fmt.Sprintf("%s (%s)", winner.Player.Name, winner.Player.Main)
// 		}
// 		if winner.Player.Name != "Rot" { // Rot is left on corpse
// 			needsLooted = append(needsLooted, b.Item)
// 		}
// 		fields = append(fields, &discordgo.MessageEmbedField{
// 			Name:  displayName,
// 			Value: strconv.Itoa(winner.Amount),
// 		})
// 	}
// 	// var author discordgo.MessageEmbedAuthor
// 	// author.Name = sig.String()
// 	// var provider discordgo.MessageEmbedProvider
// 	// provider.Name = sig.String()
// 	var footer discordgo.MessageEmbedFooter
// 	footer.Text = getPlayerName(configuration.Everquest.LogPath) + " - " + b.Zone
// 	footer.IconURL = configuration.Discord.LootIcon

// 	embed := discordgo.MessageEmbed{
// 		URL:    b.URL,
// 		Title:  b.Item,
// 		Type:   discordgo.EmbedTypeRich,
// 		Fields: fields,
// 		Footer: &footer,
// 	}
// 	dMsg, err := discord.ChannelMessageSendEmbed(configuration.Discord.LootChannelID, &embed)
// 	// _, err := discord.ChannelMessageSend(configuration.LootChannelID, response)
// 	if err != nil {
// 		Err.Printf("Error sending discord message: %s", err.Error())
// 	}
// 	err = discord.MessageReactionAdd(configuration.Discord.LootChannelID, dMsg.ID, configuration.Discord.InvestigationStartEmoji)
// 	if err != nil {
// 		Err.Printf("Error adding base reaction: %s", err.Error())
// 	}
// 	// addReact(dMsg.ID)
// 	b.InvestigationLogs = investigation
// 	// Write bid to archive
// 	writeArchive(dMsg.ID, *b)
// 	printHUD()
// }

// // Winner is the user who won items
// type Winner struct {
// 	Player Player `json:"Player"`
// 	Amount int    `json:"Amount"`
// }

// func (b *BidItem) getWinners(count int) []Winner {
// 	// Account for no bids (rot)
// 	for i := 0; i < count+1; i++ {
// 		fakeBid := Bid{
// 			Bidder: "Rot",
// 			Amount: 0,
// 		}
// 		b.Bids = append(b.Bids, fakeBid)
// 	}
// 	// sort.Sort(sort.Reverse(ByBid(b.Bids)))
// 	// b.Bids = sortBids(b.Bids)
// 	// TODO: Account for ties
// 	winningbid := b.Bids[count].Amount + configuration.Bids.Increments
// 	if len(b.Bids) > count && b.Bids[count-1].Amount == b.Bids[count].Amount && b.Bids[count].Bidder != "Rot" {
// 		b.RollOff = true // for logging/replay purposes
// 		winningbid = b.Bids[count].Amount
// 		var rollers string
// 		lowestBidder := count - 1

// 		for i := count - 1; i < len(b.Bids); i++ {
// 			if b.Bids[i].Bidder != "Rot" && b.Bids[i].Amount == winningbid {
// 				if i == lowestBidder {
// 					rollers = fmt.Sprintf("Roll off between %s", b.Bids[i].Bidder)
// 				} else {
// 					rollers = fmt.Sprintf("%s and %s", rollers, b.Bids[i].Bidder)
// 					count++
// 				}
// 				needsRolled = append(needsRolled, b.Bids[i].Bidder)
// 			}
// 		}
// 		rollers = fmt.Sprintf("%s required!", rollers)
// 		Info.Printf("%s", rollers)
// 		DiscordF(configuration.Discord.LootChannelID, "%s", rollers)
// 	}
// 	if b.Bids[0].Bidder != "Rot" && winningbid < configuration.Bids.MinimumBid {
// 		winningbid = configuration.Bids.MinimumBid
// 	}
// 	if b.Bids[count].Player.Rank != b.Bids[count-1].Player.Rank { // If we outrank them minimum bid is 10
// 		if configuration.Bids.SecondMainsBidAsMains && (b.Bids[count].Player.Rank == cMain || b.Bids[count].Player.Rank == cSecondMain) && (b.Bids[count-1].Player.Rank == cMain || b.Bids[count-1].Player.Rank == cSecondMain) {
// 			Info.Printf("%s was outranked by %s but is a main vs secondmain so ignoring", b.Bids[count].Player.Name, b.Bids[count-1].Player.Name)
// 		} else {
// 			Info.Printf("%s's rank of %d outranks %s's rank of %d", b.Bids[count-1].Player.Name, b.Bids[count-1].Player.Rank, b.Bids[count].Player.Name, b.Bids[count].Player.Rank)
// 			winningbid = configuration.Bids.MinimumBid
// 		}
// 	}
// 	// if winningbid > b.Bids[0].Amount { // account for ties
// 	// 	winningbid = b.Bids[0].Amount
// 	// 	if winningbid == b.Bids[count-1].Amount && b.Bids[count-1].Bidder != "Rot" {
// 	// 		// A ROLL OFF IS NEEDED
// 	// 		// Determine AMOUNT of ties
// 	// 		Info.Printf("Roll off required!: %#+v vs %#+v\n", b.Bids[count], b.Bids[count-1])
// 	// 		var ties int
// 	// 		for _, bidder := range b.Bids {
// 	// 			// TODO: account for secondmains bidding as mains
// 	// 			if bidder.Amount == winningbid && b.Bids[0].Player.Rank == bidder.Player.Rank { // is tied winner
// 	// 				ties++
// 	// 			}
// 	// 		}
// 	// 		count = ties     // show winners == amount of ties to imply roll off
// 	// 		b.RollOff = true // for logging/replay purposes
// 	// 		discord.ChannelMessageSend(configuration.Discord.LootChannelID, "Roll off required!")
// 	// 	}
// 	// }
// 	var winners []Winner
// 	rot := &Player{
// 		Name:  "Rot",
// 		DKP:   0,
// 		Main:  "Rot",
// 		Level: 0,
// 		Class: "Necromancer",
// 		Rank:  cInactive,
// 		Alt:   true,
// 	}
// 	for i := 0; i < count; i++ {
// 		var win Player
// 		if _, ok := roster[b.Bids[i].Bidder]; ok {
// 			win = *roster[b.Bids[i].Bidder]
// 		} else {
// 			win = *rot
// 		}
// 		winner := Winner{
// 			Player: win,
// 			Amount: winningbid,
// 		}
// 		winners = append(winners, winner)
// 	}
// 	b.Winners = winners
// 	return winners
// }

// func sortBids(bids []Bid) []Bid {
// 	var mains []Bid
// 	var secondmains []Bid
// 	var recruits []Bid
// 	var alts []Bid
// 	var inactives []Bid

// 	for _, bid := range bids {
// 		switch bid.Player.Rank {
// 		case cMain:
// 			mains = append(mains, bid)
// 		case cSecondMain:
// 			secondmains = append(secondmains, bid)
// 		case cRecruit:
// 			recruits = append(recruits, bid)
// 		case cAlt:
// 			alts = append(alts, bid)
// 		case cInactive:
// 			inactives = append(inactives, bid)
// 		}
// 	}
// 	if len(mains) == 0 {
// 		Info.Printf("No mains bid on this item")
// 	}
// 	if configuration.Bids.SecondMainsBidAsMains { // we don't need to do anything if no mains bid
// 		if configuration.Bids.SecondMainAsMainMaxBid > 0 && len(mains) > 0 { // We need to lower their bids that are > 200 only if mains bid
// 			// looping like this uses a copy I think so we need to make a new slice
// 			var fixedSMains []Bid
// 			for _, sMain := range secondmains {
// 				if sMain.Amount > configuration.Bids.SecondMainAsMainMaxBid { // They bid too much, so lower it to the max allowed
// 					Info.Printf("Secondmain %s bid more than the max allowed of %d, setting to max\n", sMain.Player.Name, configuration.Bids.SecondMainAsMainMaxBid)
// 					sMain.Amount = configuration.Bids.SecondMainAsMainMaxBid
// 				}
// 				fixedSMains = append(fixedSMains, sMain)
// 			}
// 			mains = append(mains, fixedSMains...)
// 		} else {
// 			mains = append(mains, secondmains...)
// 		}
// 	}
// 	sort.Sort(sort.Reverse(ByBid(mains)))
// 	if !configuration.Bids.SecondMainsBidAsMains {
// 		sort.Sort(sort.Reverse(ByBid(secondmains)))
// 	}
// 	sort.Sort(sort.Reverse(ByBid(recruits)))
// 	sort.Sort(sort.Reverse(ByBid(alts)))
// 	sort.Sort(sort.Reverse(ByBid(inactives)))
// 	var winners []Bid
// 	winners = append(winners, mains...)
// 	if !configuration.Bids.SecondMainsBidAsMains {
// 		winners = append(winners, secondmains...)
// 	}
// 	winners = append(winners, recruits...)
// 	winners = append(winners, alts...)
// 	winners = append(winners, inactives...)
// 	return winners
// }

// ByBid is for finding the highest bidders
// type ByBid []Bid

// func (a ByBid) Len() int           { return len(a) }
// func (a ByBid) Less(i, j int) bool { return a[i].Amount < a[j].Amount }
// func (a ByBid) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// func dumpBids() {
// 	l := LogInit("dumpBids-main.go")
// 	defer l.End()
// 	for i, bid := range bids {
// 		l.TraceF("%s: %#+v\n", i, bid)
// 	}
// }

func isItem(name string) int {
	// itemLock.Lock()
	// defer itemLock.Unlock()
	// name = strings.ToLower(name) // Make it lowercase to match database
	// if _, ok := itemDB[name]; ok {
	// 	return itemDB[name]
	// }
	// Warn.Printf("Cannot find item: %s\n", name)
	id, err := itemDB.FindIDByName(name)
	if err != nil {
		return -1
	}
	return id
}

// func checkClosedBids() {
// 	for k, bi := range bids {
// 		if bi.isBidEnded() {
// 			bi.closeBid()
// 			delete(bids, k) // delete key
// 		}
// 	}
// }

// func writeArchive(name string, data BidItem) {
// 	Info.Printf("Writing archive %s to file", name)
// 	file, err := json.MarshalIndent(data, "", " ")
// 	if err != nil {
// 		Err.Printf("Error converting to JSON: %s", err.Error())
// 	}

// 	err = ioutil.WriteFile("archive/"+name+".json", file, 0644)
// 	if err != nil {
// 		Err.Printf("Error writing archive to file: %s", err.Error())
// 	}
// 	archives = append(archives, name) // add to known archive
// }

// Investigation is raw logs during a specified time-frame for verifying failed bids
type Investigation struct {
	Messages []everquest.EqLog `json:"Messages"`
}

func (i *Investigation) addLog(l everquest.EqLog) {
	i.Messages = append(i.Messages, l)
	if len(i.Messages) > configuration.Main.InvestigationLogLimitMinutes { // remove the oldest log
		copy(i.Messages[0:], i.Messages[1:])
		i.Messages[len(i.Messages)-1] = everquest.EqLog{} // or the zero value of T
		i.Messages = i.Messages[:len(i.Messages)-1]
	}
}

func getTime() time.Time {
	if configuration.Main.ReadEntireLog { // We are simulating/testing things, we need to use time from logs
		return currentTime
	}
	return time.Now()
}

func uploadArchive(id string) {
	file, err := os.Open("archive/" + id + ".json") // TODO: Account for linux, and maliciousness
	if !configuration.Discord.UseDiscord {
		return
	}
	if err != nil {
		Err.Printf("Error finding archive: %s", err.Error())
		discord.ChannelMessageSend(configuration.Discord.InvestigationChannelID, "Error uploading investigation: "+id)
	} else {
		discord.ChannelFileSend(configuration.Discord.InvestigationChannelID, id+".json", file)
	}
}

// var hourly int
// var bosses int

// func uploadRaidDump(filename string) {
// 	file, err := os.Open(configuration.Everquest.BaseFolder + "/" + filename)
// 	stamp := time.Now().Format("20060102")
// 	if err != nil {
// 		Err.Printf("Error finding Raid Dump: %s", err.Error())
// 		discord.ChannelMessageSend(configuration.Discord.RaidDumpChannelID, "Error uploading Raid Dump: "+filename)
// 	} else {
// 		if raidDumps == 0 {
// 			DiscordF(configuration.Discord.RaidDumpChannelID, "%s uploaded an on-time raid dump at %s for %s", getPlayerName(configuration.Everquest.LogPath), getTime().String(), currentZone)
// 			raidDumps++
// 			// Start timer
// 			raidStart = getTime().Round(1 * time.Hour)
// 			nextDump = raidStart.Add(1 * time.Hour)
// 			discord.ChannelFileSend(configuration.Discord.RaidDumpChannelID, stamp+"_raid_start.txt", file)
// 		} else {
// 			if needsDump && getTime().Round(1*time.Hour) == nextDump {
// 				raidDumps++
// 				nextDump = nextDump.Add(1 * time.Hour)
// 				needsDump = false
// 				hourly++
// 				hString := strconv.Itoa(hourly)
// 				DiscordF(configuration.Discord.RaidDumpChannelID, "%s uploaded an hourly raid dump at %s for %s", getPlayerName(configuration.Everquest.LogPath), getTime().String(), currentZone)
// 				discord.ChannelFileSend(configuration.Discord.RaidDumpChannelID, stamp+"_hour"+hString+".txt", file)
// 			} else {
// 				bosses++
// 				hBosses := strconv.Itoa(bosses)
// 				DiscordF(configuration.Discord.RaidDumpChannelID, "%s uploaded a boss kill raid dump at %s for %s", getPlayerName(configuration.Everquest.LogPath), getTime().String(), currentZone)
// 				discord.ChannelFileSend(configuration.Discord.RaidDumpChannelID, stamp+"_boss"+hBosses+".txt", file)
// 			}
// 		}
// 		// discord.ChannelFileSend(configuration.Discord.RaidDumpChannelID, filename, file)
// 	}
// }

// func uploadGuildDump(filename string) { // TODO: this crashes bidbot if u can't upload?
// 	var guild everquest.Guild
// 	guild.LoadFromPath(configuration.Everquest.BaseFolder+"/"+filename, Err)
// 	//Encode the data
// 	postBody, _ := json.Marshal(guild)
// 	responseBody := bytes.NewBuffer(postBody)
// 	// Create a Bearer string by appending string access token
// 	var bearer = "Bearer " + configuration.Main.GuildUploadLicense

// 	// Create a new request using http
// 	req, err := http.NewRequest(http.MethodPost, configuration.Main.GuildUploadAPIURL, responseBody)
// 	if err != nil {
// 		log.Println("Error creating request.\n[ERROR] -", err)
// 	}

// 	// add authorization header to the req
// 	req.Header.Add("Authorization", bearer)
// 	req.Header.Add("Content-Type", "application/json")

// 	// Send req using http Client
// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		log.Println("Error on response.\n[ERROR] -", err)
// 	}
// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		log.Println("Error while reading the response bytes:", err)
// 	}
// 	log.Println(string([]byte(body)))
// }

// func checkReactions() {
// 	discord.MessageReactions(configuration.LootChannelID, "801309136020570154", ":mag_right:", 100, "", "")
// }

// func loadRoster(file string) {
// 	csvfile, err := os.Open(file)
// 	if err != nil {
// 		Err.Fatalln("Couldn't open the roster csv file", err)
// 	}
// 	defer csvfile.Close()

// 	// Parse the file
// 	r := csv.NewReader(csvfile)
// 	r.Comma = '\t'
// 	//r := csv.NewReader(bufio.NewReader(csvfile))

// 	// Iterate through the records
// 	headerSkipped := true // guild dumps have no header
// 	for {
// 		// Read each record from csv
// 		record, err := r.Read()
// 		if !headerSkipped {
// 			headerSkipped = true
// 			// skip header line
// 			continue
// 		}
// 		if err == io.EOF {
// 			break
// 		}
// 		if err != nil {
// 			Err.Printf("Error loading roster: %s", err.Error())
// 		}
// 		var player Player
// 		player.Name = record[0]
// 		// fmt.Printf("ID: %s Name: %s\n", record[0], record[1])
// 		player.Level, err = strconv.Atoi(record[1])
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		player.Class = record[2]
// 		if record[4] == "A" {
// 			player.Alt = true
// 			player.Rank = cAlt // Default to alt for now
// 			// Figure out if secondmain, alt, recruit
// 			// Figure out alt's main
// 			r, _ := regexp.Compile(configuration.Everquest.RegexIsAlt) // TODO: Make it NOT match if "CLOSED" or "wins" is in this, otherwise we will open aditional bids - also if we have a dedicated box, we can match that
// 			Altresult := r.FindStringSubmatch(record[7])
// 			if len(Altresult) > 0 {
// 				player.Main = Altresult[1]
// 			}
// 			// Figure out secondmain and it's main
// 			r, _ = regexp.Compile(configuration.Everquest.RegexIsSecondMain) // TODO: Make it NOT match if "CLOSED" or "wins" is in this, otherwise we will open aditional bids - also if we have a dedicated box, we can match that
// 			SecondMainResult := r.FindStringSubmatch(record[7])
// 			if len(SecondMainResult) > 0 {
// 				player.Main = SecondMainResult[1]
// 				player.Rank = cSecondMain
// 			}
// 		} // defaults to false
// 		if !player.Alt && isRaider(record[3]) { // not an alt and a raider, so is a main
// 			player.Rank = cMain
// 			player.Main = player.Name
// 		}
// 		if record[3] == "Recruit" || record[3] == "Member" { // Members are Recruits for bidding purposes
// 			player.Rank = cRecruit
// 			player.Main = player.Name
// 		}
// 		if record[3] == "Inactive" {
// 			player.Rank = cInactive
// 			player.Main = player.Name
// 		}
// 		roster[player.Name] = &player
// 	}
// 	Info.Printf("Loaded Roster")
// }

// func dumpPlayers() {
// 	for _, p := range roster {
// 		fmt.Printf("%#+v\n", p)
// 	}
// }

// Player is represented by Name, Level, Class, Rank, Alt, Last On, Zone, Note, Tribute Status, Unk_1, Unk_2, Last Donation, Private Note
// type Player struct {
// 	Name  string `json:"Name"`
// 	Main  string `json:"Main"` // Name of player's main
// 	Level int    `json:"Level"`
// 	Class string `json:"Class"`
// 	Rank  int    `json:"Rank"` // this is a meta field, its not direct from the rank column
// 	Alt   bool   `json:"Alt"`
// 	DKP   int    `json:"DKP"` // this is filled in post from google sheets
// }

// func isRaider(rank string) bool {
// 	for _, r := range configuration.Everquest.GuildRaidingRanks {
// 		if r == rank {
// 			return true
// 		}
// 	}
// 	return false
// }

// func updatePlayerDKP(name string, dkp int) {
// 	rosterLock.Lock()
// 	defer rosterLock.Unlock()
// 	if _, ok := roster[name]; ok {
// 		roster[name].DKP = dkp
// 		return
// 	}
// 	Err.Printf("Cannot find player to update DKP: %s giving them 0 dkp", name)
// 	// DiscordF("Error configuring %s's DKP, are they on the DKP sheet, Roster Dump, and are the Guild Notes correct?", name)
// 	roster[name] = &Player{
// 		Name:  name,
// 		Main:  name,
// 		DKP:   0,
// 		Level: 0,
// 		Class: "Unknown",
// 		Rank:  cInactive,
// 		Alt:   false,
// 	}
// }

// func hasEnoughDKP(name string, amount int) bool {
// 	rosterLock.Lock()
// 	defer rosterLock.Unlock()
// 	var bHasDKP bool
// 	if _, ok := roster[name]; ok {
// 		bHasDKP = true
// 	}
// 	if amount < 10 || (bHasDKP && roster[name].DKP >= amount) { // You can always spend 10dkp
// 		return true
// 	}
// 	Warn.Printf("%s does not have %d DKP but tried to spend it", name, amount)
// 	return false
// }

// func getMaxDKP(name string) int {
// 	rosterLock.Lock()
// 	defer rosterLock.Unlock()
// 	if _, ok := roster[name]; ok {
// 		if roster[name].DKP < 10 { // You can always spend 10dkp
// 			return 10
// 		}
// 		return roster[name].DKP
// 	}
// 	Err.Printf("Cannot obtain max DKP for %s", name)
// 	return -5
// }

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
// 			Err.Printf("Error adding base reaction: %s", err.Error())
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
	return split[1]
}

func getPlayerServer(logFile string) string {
	// l := LogInit("getPlayerName-commands.go")
	// defer l.End()
	logFile = filepath.Base(logFile)
	extension := filepath.Ext(logFile)
	name := logFile[0 : len(logFile)-len(extension)]
	split := strings.Split(name, "_")
	// fmt.Printf("LogFile: %s\nExtension: %s\nName: %s\nSplit: %#+v\n", logFile, extension, name, split)
	if len(split) < 3 {
		return "unknown"
	}
	return split[2]
}

// func getRecentRosterDump(path string) string {
// 	var files []string

// 	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
// 		if strings.HasPrefix(filepath.Base(path), configuration.Everquest.GuildName) {
// 			files = append(files, filepath.Base(path))
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		panic(err)
// 	}
// 	if !isDumpOutOfDate(files[len(files)-1]) {
// 		DiscordF(configuration.Discord.InvestigationChannelID, "**Guild dump %s is out of date, this needs updated with ALL members (including offline and alts) before bidbot is ran**", files[len(files)-1])
// 	}

// 	return files[len(files)-1] // return last file - should be latest
// 	// It looks like files are already sorted by date, we don't need this
// 	// var times []time.Time
// 	// for _, file := range files {
// 	// 	// Remove extension
// 	// 	file := strings.TrimSuffix(file, filepath.Ext(file))
// 	// 	spltFile := strings.Split(file, "-")
// 	// 	if len(spltFile) > 2 { // should always happen
// 	// 		timeString := spltFile[1] + "-" + spltFile[2] // only parse the time
// 	// 		t, err := time.Parse("20060102-150405", timeString)
// 	// 		if err != nil {
// 	// 			Err.Printf("Error parsing time of roster dump: %s", err.Error())
// 	// 		}
// 	// 		times = append(times, t)
// 	// 	}
// 	// }
// 	// return ""
// }

func isDumpOutOfDate(dump string) bool {
	// Vets of Norrath_aradune-20210124-083635
	// location, err := time.LoadLocation("America/Chicago")
	// if err != nil {
	// 	Err.Printf("Error parsing tz : %s", err.Error())
	// }
	// t := getTime()
	zone, _ := time.Now().Zone()
	name := strings.Split(dump, "-") // seperate by hypen so [1] is the day we care about
	format := "20060102MST"
	logDate, err := time.Parse(format, name[1]+zone)
	// logDate = logDate.In(location)
	if err != nil {
		Err.Printf("Error parsing time of guild dump : %s", err.Error())
	}
	// fmt.Printf("LogDate: %s Before: %s After: %s Dump: %s Result: %t\n", logDate.String(), time.Now().String(), time.Now().Add(-24*time.Hour).String(), dump, logDate.Before(time.Now()) && logDate.After(time.Now().Add(-24*time.Hour)))
	return !(logDate.Before(time.Now()) && logDate.After(time.Now().Add(-24*time.Hour)))
}

// // ByTime is for finding the most recent item
// type ByTime []time.Time

// func (a ByTime) Len() int           { return len(a) }
// func (a ByTime) Less(i, j int) bool { return a[i].Before(a[j]) }
// func (a ByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
