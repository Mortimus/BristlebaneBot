package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/pelletier/go-toml"
)

var configuration Configuration

var configPath = "config.toml"

func init() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	configuration, err = loadConfig(exPath + "/" + configPath)
	if err != nil {
		configuration, err = loadConfig(configPath)
		if err != nil {
			panic(err)
		}
	} else {
		configPath = exPath + "/" + configPath
	}
}

type Main struct {
	ReadEntireLog                bool   `comment:"Should we read the entire character log or just new entries"`
	LogPollRate                  int    `comment:"How often to check for new entries in the character log in seconds"`
	InvestigationLogLimitMinutes int    `comment:"How many minutes of chat to log for an investigation"`
	LucyURLPrefix                string `comment:"URL prefix for generating links based on item ID"`
	GuildUploadAPIURL            string `comment:"URL for uploadiing Guild dumps to the discord bot"`
	GuildUploadLicense           string `comment:"License key for uploading guild dumps"`
}

type SpellOverride struct {
	Name    string `comment:"Name of spell to override"`
	SpellID int    `comment:"ID of overrode spell"`
}

type Bids struct {
	OpenBidTimer           int    `comment:"Number of minutes to keep bids open before auto closing"`
	MinimumBid             int    `comment:"Minimum bid accepted - 10"`
	Increments             int    `comment:"Increment multiple for bids - 5"`
	SecondMainAsMainMaxBid int    `comment:"Max value that bids can bid against mains, set to 0 for infinite"`
	MaxBid                 int    `comment:"Max bid allowed for a single item needs to be lower to allow rounding down default: 9000"`
	RegexClosedBid         string `comment:"Regex to detect a bid has been closed"`
	RegexOpenBid           string `comment:"Regex to detect a bid has been opened"`
	RegexTellBid           string `comment:"Regex to detect a bid being sent via tell"`
	CloseAutomatically     bool   `comment:"Close bids automatically after timer has expired"`
	SecondMainsBidAsMains  bool   `comment:"Will second mains be tiered the same as mains"`
}

type Discord struct {
	Token                    string   `comment:"Discord Bot Token"`
	LootChannelID            string   `comment:"Discord Channel to sent loot to"`
	InvestigationChannelID   string   `comment:"Discord Channel to sent investigations to"`
	LootIcon                 string   `comment:"Icon used for the loot in discord"`
	InvestigationStartEmoji  string   `comment:"Emoji response used to start an investigation"`
	GuildID                  string   `comment:"Discord Guild ID"`
	RaidDumpChannelID        string   `comment:"Discord channel to send raid dumps to"`
	SpellDumpChannelID       string   `comment:"Discord channel to send spell loot to"`
	FlagChannelID            string   `comment:"Discord channel to send acquired flags to"`
	ParseChannelID           string   `comment:"Discord channel to send parses to"`
	UseDiscord               bool     `comment:"Should we use discord"`
	InvestigationMinRequired int      `comment:"Number of reactions required to start investigation"`
	PrivRoles                []string `comment:"Discord roles that are considered privledged, for starting investigations"`
}

type Everquest struct {
	LogPath           string   `comment:"path to character log file"`
	ItemDB            string   `comment:"path to the eqitems item database"`
	SpellDB           string   `comment:"path to the lucydb spell database"`
	RegexIsAlt        string   `comment:"Regex to determine if character is an alt"`
	RegexIsSecondMain string   `comment:"Regex to determine if character is a 2nd main"`
	GuildName         string   `comment:"Guild name to determine guild dumps"`
	BaseFolder        string   `comment:"Base folder where eqgame.exe is located, for determining logs and dumps"`
	RegexLoot         string   `comment:"Regex to detect when an item has been looted"`
	ParseIdentifier   string   `comment:"string that a parse dump will contain"`
	ParseChannel      string   `comment:"everquest channel to monitor for parses"`
	RegexSlay         string   `comment:"Regex to detect when a mob is slain"`
	RegexRoll         string   `comment:"Regex to detect when a does a die roll"`
	GuildRaidingRanks []string `comment:"Guild Ranks that can bid on items"`
	FlagGiver         []string `comment:"log text for a character getting a flag - Hail, a planar projection"`
	DKPGiver          []string `comment:"mob names that we apply DKP for"`
	SpellProvider     []string `comment:"item that provides a spell like Spectral Parchment"`
}

type Google struct {
	AccessToken             string
	TokenType               string
	RefreshToken            string
	ClientID                string
	ProjectID               string
	AuthURI                 string
	TokenURI                string
	AuthProviderx509CertURL string
	ClientSecret            string
	RedirectURIs            []string
	Expiry                  time.Time
}

type Sheets struct {
	DKPSheetURL              string `comment:"Google sheets url for the DKP sheet"`         // Google sheets url for the DKP sheet
	DKPSummarySheetName      string `comment:"Google sheets sheet name for the DKP lookup"` // Google sheets sheet name for the DKP lookup
	RawSheetURL              string `comment:"Raw DKP/Attendance sheet URL"`                // Google sheets sheet URL for the raw data
	RawSheetName             string `comment:"Raw DKP/Attendance sheet sheet name"`
	SpellSheetURL            string `comment:"Google sheets sheet URL for the spell lookup"` // Google sheets sheet URL for the spell lookup
	BossesSheetName          string `comment:"Bossses sheet sheet name"`
	SpellSheetSpellCol       int    `comment:"Google sheet sheet column for spell names"`             // Google sheet sheet column for spell names
	SpellSheetPlayerStartCol int    `comment:"Google sheet sheet column for when player names start"` // Google sheet sheet column for when player names start
	SpellSheetDataRowStart   int    `comment:"Google sheet sheet row for when spell names start"`     // Google sheet sheet row for when spell names start
	SpellSheetPlayerRow      int    `comment:"Google sheet sheet row for when player names start"`
	RawSheetPlayerCol        int    `comment:"Raw DKP/Attendance sheet sheet column for player names"`
	RawSheetDateCol          int    `comment:"Raw DKP/Attendance sheet sheet column for date"`
	RawSheetDKPCol           int    `comment:"Raw DKP/Attendance sheet sheet column for dkp"`
	RawSheetAttendanceCol    int    `comment:"Raw DKP/Attendance sheet sheet column for attendance"`
	BossSheetZoneCol         int    `comment:"Boss sheet column for zone name"`
	BossSheetNoteCol         int    `comment:"Boss sheet column for zone name"`
	BossSheetBossCol         int    `comment:"Boss sheet column for zone name"`
	BossSheetDKPCol          int    `comment:"Boss sheet column for zone name"`
	BossSheetFTKCol          int    `comment:"Boss sheet column for zone name"`
	BossSheetisFTKCol        int    `comment:"Boss sheet column for zone name"`
	DKPSummarySheetPlayerCol int    `comment:"Google sheet sheet column for player names"` // Google sheet sheet column for player names
	DKPSummarySheetDKPCol    int    `comment:"Google sheet sheet column for dkp count"`
}

type Log struct {
	Level int    `comment:"How much to log Warn:0 Err:1 Info:2 Debug:3"`
	Path  string `comment:"Where to store the log file use linux formatting or escape slashes for windows"`
}

type Configuration struct {
	Main      Main
	Everquest Everquest
	Log       Log
	Bids      Bids
	Discord   Discord
	Google    Google
	Sheets    Sheets
	Overrides []SpellOverride `comment:"Spell that finds as wrong ID, force an ID here"`
}

func loadConfig(path string) (Configuration, error) {
	config := Configuration{}
	configFile, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = toml.Unmarshal(configFile, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func (c Configuration) save(path string) {
	out, err := toml.Marshal(c)
	if err != nil {
		Err.Printf("Error marshalling config: %s", err.Error())
	}
	err = ioutil.WriteFile(path, out, 0644)
	if err != nil {
		Err.Printf("Error writing config: %s", err.Error())
	}
}
