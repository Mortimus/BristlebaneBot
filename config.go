package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

const configPath = "config.json"
const failPath = "C:\\"

var configuration Configuration

// Configuration stores all our user defined variables
type Configuration struct {
	LogLevel                     int       `json:"LogLevel"`                     // 0=Off,1=Error,2=Warn,3=Info,4=Debug
	LogPath                      string    `json:"LogPath"`                      // Where to write logs to
	EQLogPath                    string    `json:"EQLogPath"`                    // Where to read logs from
	EQBaseLogLine                string    `json:"EQBaseLogLine"`                // Regex for a eq log line
	ReadEntireLog                bool      `json:"ReadEntireLog"`                // Read the entire log or only new
	LogPollRate                  int       `json:"LogPollRate"`                  // How often to read the log if it reaches EOF in seconds
	LucyItems                    string    `json:"LucyItems"`                    // Lucy items .txt
	BidTimerMinutes              int       `json:"BidTimerMinutes"`              // How many minutes bids are open
	MinimumBid                   int       `json:"MinimumBid"`                   // Minimum Bid Amount
	BidIncrements                int       `json:"BidIncrements"`                // Bids must be a multiple of this
	LucyURLPrefix                string    `json:"LucyURLPrefix"`                // Lucy URL prefix for creating item links
	DiscordToken                 string    `json:"DiscordToken"`                 // Discord Bot Token for Authentication
	LootChannelID                string    `json:"LootChannelID"`                // Channel ID on where to send open Bids and winners
	InvestigationChannelID       string    `json:"InvestigationChannelID"`       // Channel ID on where to send archive json files for investigation
	InvestigationLogLimitMinutes int       `json:"InvestigationLogLimitMinutes"` // How many tells to append to archives for investigation
	DiscordLootIcon              string    `json:"DiscordLootIcon"`              // Icon to show for discord rich message
	InvestigationStartEmoji      string    `json:"InvestigationStartEmoji"`      // Emoji required to start an investigation
	InvestigationStartMinReq     int       `json:"InvestigationStartMinReq"`     // Amount of emoji required from priv users to start an investigation
	GuildRosterPath              string    `json:"GuildRosterPath"`              // Path to the guild name_server-timestamp.txt guild dump including alts and offline
	GuildRaidingRoles            []string  `json:"GuildraidingRoles"`            // Roles that would be a raider, raider+officer/leader/etc
	RegexIsAlt                   string    `json:"RegexIsAlt"`                   // Regex to find out alt's main
	RegexIs2ndMain               string    `json:"RegexIs2ndMain"`               // Regex to find out if secondmain and main's name
	AccessToken                  string    `json:"access_token"`                 // Google Access Token
	TokenType                    string    `json:"token_type"`                   // Google Token Type
	RefreshToken                 string    `json:"refresh_token"`                // Google Refresh Token
	Expiry                       time.Time `json:"expiry"`                       // Google Expiration Date
	ClientID                     string    `json:"client_id"`                    // Google Client ID
	ProjectID                    string    `json:"project_id"`                   // Google Project ID
	AuthURI                      string    `json:"auth_uri"`                     // Google Auth URI
	TokenURI                     string    `json:"token_uri"`                    // Google Token URI
	AuthProviderx509CertURL      string    `json:"auth_provider_x509_cert_url"`  // Google Cert URL
	ClientSecret                 string    `json:"client_secret"`                // Google Client Secret
	RedirectURIs                 []string  `json:"redirect_uris"`                // Google Redirect URIs
	DKPSheetURL                  string    `json:"DKPSheetURL"`                  // Google sheets url for the DKP sheet
	DKPSummarySheetName          string    `json:"DKPSummarySheetName"`          // Google sheets sheet name for the DKP lookup
	DKPSummarySheetPlayerCol     int       `json:"DKPSummarySheetPlayerCol"`     // Google sheet sheet column for player names
	DKPSummarySheetDKPCol        int       `json:"DKPSummarySheetDKPCol"`        // Google sheet sheet column for dkp
	DiscordPrivRoles             []string  `json:"DiscordPrivRoles"`             // Discord Roles that are allowed to start investigations
	DiscordGuildID               string    `json:"DiscordGuildID"`               // Discord Guild ID
}

func init() {
	readConfig()
	log.Printf("Configuration loaded:\n %+v\n", configuration)
}

func readConfig() error {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Println(err)
	}
	if _, err := os.Stat(dir + "/" + configPath); os.IsNotExist(err) {
		// path/to/whatever does not exist
		dir, _ = os.Getwd()
	}
	if _, err := os.Stat(dir + "/" + configPath); os.IsNotExist(err) {
		// path/to/whatever does not exist
		dir = failPath
	}
	if _, err := os.Stat(dir + "/" + configPath); os.IsNotExist(err) {
		// path/to/whatever does not exist
		log.Fatal(err)
	}
	file, err := os.OpenFile(dir+"/"+configPath, os.O_RDONLY, 0444)
	defer file.Close()
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configuration)
	if err != nil {
		return err
	}
	return nil
}

func saveConfig() error {
	marshalledConfig, _ := json.MarshalIndent(configuration, "", "\t")
	err := ioutil.WriteFile(configPath, marshalledConfig, 0644)
	if err != nil {
		return err
	}
	log.Printf("Config Saved to %s\n", configPath)
	return nil
}
