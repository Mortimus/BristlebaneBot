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

var printChan = make(chan string)

var itemDB everquest.ItemDB
var spellDB everquest.SpellDB

var investigation Investigation
var currentTime time.Time // for simulating time
var archives []string     // stores all known archive files for recall

var Debug, Warn, Err, Info *log.Logger

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
