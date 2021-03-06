package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/oauth2"
	"google.golang.org/api/sheets/v4"
)

// srv is the global to connect to google sheets
var srv *sheets.Service

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
