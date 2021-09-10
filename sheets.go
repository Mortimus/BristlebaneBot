package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	everquest "github.com/Mortimus/goEverquest"
	"golang.org/x/oauth2"
	"google.golang.org/api/sheets/v4"
)

// srv is the global to connect to google sheets
var srv *sheets.Service

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	// tokFile := "token.json"
	Info.Printf("Fake loading token from file")
	tok, err := tokenFromFile("")
	if err != nil {
		Info.Printf("Token failed to load, loading from web")
		tok = getTokenFromWeb(config)
		Info.Printf("Saving token")
		saveToken("", tok)
	}
	Debug.Printf("Using Token: %+v", tok)
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)
	Info.Printf("Requesting user navigate to: %s", authURL)
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		Err.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		Err.Fatalf("Unable to retrieve token from web: %v", err)
	}
	Info.Printf("Return token: %+v", tok)
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	// f, err := os.Open(file)
	// if err != nil {
	// 	return nil, err
	// }
	// defer f.Close()
	tok := &oauth2.Token{}
	tok.AccessToken = configuration.Google.AccessToken
	tok.Expiry = configuration.Google.Expiry
	tok.RefreshToken = configuration.Google.RefreshToken
	tok.TokenType = configuration.Google.TokenType
	// err = json.NewDecoder(f).Decode(tok)
	Info.Printf("Returning token: %+v", tok)
	return tok, nil
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	// fmt.Printf("Saving credential file to: %s\n", path)
	// f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	// if err != nil {
	// 	log.Fatalf("Unable to cache oauth token: %v", err)
	// }
	// defer f.Close()
	// json.NewEncoder(f).Encode(token)
	configuration.Google.AccessToken = token.AccessToken
	configuration.Google.Expiry = token.Expiry
	configuration.Google.RefreshToken = token.RefreshToken
	configuration.Google.TokenType = token.TokenType
	Info.Printf("Saved token to configuration")
	configuration.save(configPath)
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

// func updateDKP() {
// 	spreadsheetID := configuration.Sheets.DKPSheetURL
// 	readRange := configuration.Sheets.DKPSummarySheetName
// 	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
// 	if err != nil {
// 		Err.Printf("Unable to retrieve data from sheet: %v", err)
// 		return
// 	}

// 	if len(resp.Values) == 0 {
// 		Err.Printf("Cannot read dkp sheet: %v", resp)
// 		// log.Println("No data found.")
// 	} else {
// 		// var lastClass string
// 		for i, row := range resp.Values {
// 			if i == 0 { // skip the header
// 				continue
// 			}
// 			name := fmt.Sprintf("%s", row[configuration.Sheets.DKPSummarySheetPlayerCol])
// 			name = strings.TrimSpace(name)
// 			if name != "" {
// 				sDKP := fmt.Sprintf("%s", row[configuration.Sheets.DKPSummarySheetDKPCol])
// 				sDKP = strings.ReplaceAll(sDKP, ",", "")
// 				dkp, err := strconv.Atoi(sDKP)
// 				if err != nil {
// 					Err.Printf("Error converting DKP to int: %s", err.Error())
// 					continue
// 				}
// 				updatePlayerDKP(name, dkp)
// 			}
// 		}
// 	}
// }

func findWhoNeedsSpell(s everquest.Spell) []string {
	spreadsheetID := configuration.Sheets.SpellSheetURL
	classes := s.GetClasses()
	var players []string
	for _, class := range classes {
		if class == "Unknown" {
			continue
		}
		Info.Printf("Finding who from class %s needs %s\n", class, s.Name)
		resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, class).Do()
		if err != nil {
			Err.Printf("Unable to retrieve data from sheet: %v", err)
			return nil
		}

		if len(resp.Values) == 0 {
			Err.Printf("Cannot read spell sheet: %v", resp)
			// log.Println("No data found.")
		} else {
			// var lastClass string
			for i, row := range resp.Values {
				// fmt.Printf("I: %d Config: %d\n", i, configuration.SpellSheetDataRowStart)
				if i < configuration.Sheets.SpellSheetDataRowStart-1 {
					continue
				}
				// fmt.Println(row)
				if len(row) <= configuration.Sheets.SpellSheetSpellCol {
					continue
				}
				spellName := fmt.Sprintf("%s", row[configuration.Sheets.SpellSheetSpellCol])
				if "Spell: "+s.Name == spellName || strings.Replace(s.Name, "Ancient ", "Ancient: ", 1) == spellName { // Ancients are dumb
					// fmt.Printf("h: %d data: %s\n", configuration.SpellSheetPlayerStartCol, row[configuration.SpellSheetPlayerStartCol])
					for h := configuration.Sheets.SpellSheetPlayerStartCol; h < len(row); h++ {
						rowString := fmt.Sprintf("%s", row[h])
						if rowString == "FALSE" {
							player := fmt.Sprintf("%s", resp.Values[configuration.Sheets.SpellSheetPlayerRow][h])
							players = append(players, player)
							Info.Printf("Player: %s needs %s\n", player, spellName)
						}
					}
					break
				}
			}
		}
	}
	if len(players) == 0 {
		players = append(players, "no one")
	}
	return players
}
