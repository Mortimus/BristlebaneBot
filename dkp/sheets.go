package dkp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	everquest "github.com/Mortimus/goEverquest"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type SheetsDB struct {
	Players []DKPHolder
	raw     []Approval
	lastRow int
	srv     *sheets.Service
}

// type AttendanceEntry struct {
// 	date   time.Time
// 	points float64
// }

func (db *SheetsDB) Connect(Info *log.Logger, options ...string) error {
	// Init Google Sheets
	gtoken := &Gtoken{
		Installed: Inst{
			ClientID:                options[0],
			ProjectID:               options[1],
			AuthURI:                 options[2],
			TokenURI:                options[3],
			AuthProviderx509CertURL: options[4],
			ClientSecret:            options[5],
			RedirectURIs:            options[6:],
		},
	}
	Info.Printf("Marshalling gToken: %+v", gtoken)
	bToken, err := json.Marshal(gtoken)
	if err != nil {
		return err
	}

	// b, err := ioutil.ReadFile("credentials.json")
	// if err != nil {
	// 	log.Fatalf("Unable to read client secret file: %v", err)
	// }

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(bToken, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		return err
	}
	client := getClient(config)
	ctx := context.Background()
	// db.srv, err = sheets.New(client)
	// db.srv = new(sheets.Service)
	// db.srv = &sheets.Service{}
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return err
	}
	db.srv = srv
	// db.lastRow, err = getRawDKP(srv, &db.raw, 0)
	// db.Guild = db.GetRoster()
	return db.UpdateCache(true)
}

func (db *SheetsDB) Disconnect() {
	// Do nothing
}

func (db *SheetsDB) UpdateCache(invalidate bool) error {
	var row int
	var err error
	if !invalidate {
		row = db.lastRow
	}
	db.lastRow, err = getRawDKP(db.srv, &db.raw, row)
	db.SeedPlayers() // TODO: We only want to update if not invalidating
	return err
}

func (db *SheetsDB) GetDKP() []DKPHolder {
	return db.Players
}

func (db *SheetsDB) LookupAttendance(player string, days int) float64 {
	var result float64
	for _, event := range db.raw {
		if event.Player == player && event.Date.After(time.Now().AddDate(0, 0, -1*days)) {
			result += event.Attendance
		}
	}
	result = math.Round(result*100) / 100
	return result
}

func (db *SheetsDB) LookupDKP(player string) int {
	var result int
	for _, event := range db.raw {
		if event.Player == player {
			result += event.Points
		}
	}
	return result
}

func (db *SheetsDB) SeedPlayers(guild *everquest.Guild) {
	pMap := make(map[string]*DKPHolder)
	for _, member := range guild.Members {
		if _, ok := pMap[member.Name]; !ok {
			pMap[member.Name] = &DKPHolder{}
		}
		pMap[member.Name].GuildMember = member
		pMap[member.Name].Name = member.Name
	}
	for _, approval := range db.raw {
		if _, ok := pMap[approval.Player]; !ok {
			continue
		}
		pMap[approval.Player].DKP += approval.Points
		pMap[approval.Player].Name = approval.Player
	}
	db.Players = make([]DKPHolder, 0)
	for _, holder := range pMap {
		nPlayer := DKPHolder{
			Name:        holder.Name,
			DKP:         holder.DKP,
			Thirty:      db.LookupAttendance(holder.Name, 30),
			Sixty:       db.LookupAttendance(holder.Name, 60),
			Ninety:      db.LookupAttendance(holder.Name, 90),
			Lifetime:    db.LookupAttendance(holder.Name, 9999),
			GuildMember: holder.GuildMember,
			Main:        holder.GetMain(guild.Members),
		}
		nPlayer.DKPRank = nPlayer.GetDKPRank()
		if nPlayer.Alt && nPlayer.Main != "" { // Update DKP to match main
			nPlayer.DKP = pMap[nPlayer.Main].DKP
		}
		if nPlayer.Rank == "" {
			nPlayer.Rank = "Inactive"
		}
		if nPlayer.Level == 0 {
			nPlayer.Level = 1
		}
		db.Players = append(db.Players, nPlayer)
	}
}

func (db *SheetsDB) GetDKPApprovals(page int, bApproved bool) ([]Approval, int) {
	// pages := len(db.raw)/RESULTSPERPAGE
	if page < 0 {
		page = 0
	}
	var results []Approval
	var usable []Approval
	if !bApproved {
		for _, app := range db.raw {
			if !app.Approved {
				usable = append(usable, app)
			}
		}
	} else {
		usable = db.raw
	}

	sort.Sort(sort.Reverse(ByDate(usable)))
	for i, appr := range usable {
		if i > page*RESULTSPERPAGE && i < page*RESULTSPERPAGE+RESULTSPERPAGE {
			results = append(results, appr)
		}
	}
	return results, len(usable) - 1
}

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
	Info.Printf("Using Token: %+v", tok)
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
	tok.AccessToken = AccessToken
	tok.Expiry = Expiry
	tok.RefreshToken = RefreshToken
	tok.TokenType = TokenType
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
	// AccessToken = token.AccessToken
	// Expiry = token.Expiry
	// RefreshToken = token.RefreshToken
	// TokenType = token.TokenType
	Info.Printf("Did not Save token to configuration")
	// configuration.save("config.toml")
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

func getRawDKP(srv *sheets.Service, raw *[]Approval, startRow int) (int, error) {
	Info.Printf("Getting Raw DKP from Google Sheets\n")
	// var results []DKPInput
	spreadsheetID := DKPSheetURL
	readRange := RawDKPSheetName
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		Err.Printf("Unable to retrieve data from sheet: %v", err)
		return 0, err
	}
	var currentRow int
	if startRow == 0 { // Invalidate the cache
		*raw = make([]Approval, 0)
	}
	if len(resp.Values) == 0 {
		Err.Printf("Cannot read dkp sheet: %v", resp)
		return 0, errors.New("cannot read dkp sheet")
	} else {
		for i, row := range resp.Values {
			if i == 0 || i < startRow {
				continue // skip the header and already cached rows
			}
			var dkpLine Approval
			dkpLine.Player = fmt.Sprintf("%s", row[RawDKPPlayerCol])
			dkpLine.Player = strings.TrimSpace(dkpLine.Player)
			if dkpLine.Player == "" { // Blank entry, skip
				continue
			}
			dateFormat := "01/02/06"
			dateStr := fmt.Sprintf("%s", row[RawDKPDateCol])
			dkpLine.Date, err = time.Parse(dateFormat, dateStr)
			if err != nil {
				Err.Printf("Error reading date %s\n", dateStr)
			}
			dkpLine.Event = fmt.Sprintf("%s", row[RawDKPRaidCol])
			dkpLine.Reason = fmt.Sprintf("%s", row[RawDKPReasonCol])
			pointsStr := fmt.Sprintf("%s", row[RawDKPPointsCol])
			dkpLine.Points, err = strconv.Atoi(pointsStr)
			if err != nil {
				Err.Printf("Error converting points %s\n", pointsStr)
			}
			dkpLine.Affected = fmt.Sprintf("%s", row[RawDKPAltCol])
			attStr := fmt.Sprintf("%s", row[RawDKPAttCol])
			dkpLine.Attendance, err = strconv.ParseFloat(attStr, 64)
			if err != nil {
				Err.Printf("Error converting attendance %s\n", attStr)
			}
			dkpLine.Approved = true // Sheets DB is always true
			*raw = append(*raw, dkpLine)
			// fmt.Printf("Appending %#+v\n", dkpLine)
			currentRow = i
		}
	}
	return currentRow, nil
}

func (db *SheetsDB) GetRoster() []DKPHolder {
	return db.Players
}

func (db *SheetsDB) GetPlayer(name string) DKPHolder {
	for _, mem := range db.Players {
		if mem.Name == name {
			return mem
		}
	}
	return DKPHolder{}
}

func (h *DKPHolder) GetDKPRank() int {
	if !h.Alt && (h.Rank == "<<< Raid/Class Lead/Recruitment >>>" || h.Rank == "<<< Officer >>>" || h.Rank == "Raider" || h.Rank == "<<< Guild Leader >>>") {
		return DKPMain
	}
	if !h.Alt && h.Rank == "Recruit" {
		return DKPRecruit
	}
	if h.Alt && strings.Contains(h.PublicNote, "nd Main") {
		return DKPSecondMain
	}
	if h.Alt && h.Rank != "Inactive" {
		return DKPAlt
	}
	return DKPInactive
}

func (h *DKPHolder) GetDKPRankToString() string {
	switch h.DKPRank {
	case DKPMain:
		return "Main"
	case DKPSecondMain:
		return "2nd Main"
	case DKPRecruit:
		return "Recruit"
	case DKPAlt:
		return "Alt"
	default:
		return "Inactive"
	}
}

func (h *DKPHolder) GetMain(members []everquest.GuildMember) string {
	if h.Alt {
		for _, member := range members {
			if strings.Contains(h.PublicNote, member.Name) {
				return member.Name
			}
		}
	}
	return "" // No Main
}

func (db *SheetsDB) AddApproval(app Approval, acct Account) {
	db.raw = append(db.raw, app)
	sort.Sort(ByDate(db.raw))
}
