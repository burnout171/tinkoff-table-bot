package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"google.golang.org/api/sheets/v4"
)

// ConnectionProperties holds all properties needed to create a connection
type ConnectionProperties struct {
	SpreadsheetID string
	ClientID string
	ProjectID string
	AuthURI string
	TokenURI string
	ClientSecret string
	RedirectUris string
	AccessToken string
	TokenType string
	RefreshToken string
	ExpireTime string
}

// TableService creates a connection and simplify interactions with it
type TableService struct {
	SpreadsheetID string
	service *sheets.Service
}

// NewTableService factory method to create a TableService
func NewTableService(properties *ConnectionProperties) (*TableService, error) {
	ts := TableService{}
	config, err := ts.getConfig(properties)
	if err != nil {
		return nil, err
	}
	client, err := ts.getClient(properties, config) 
	if err != nil {
		return nil, err
	}
	service, err := sheets.New(client)
	if err != nil {
		log.Printf("Unable to retrieve Sheets client: %v", err)
		return nil, err
	}
	ts.SpreadsheetID = properties.SpreadsheetID
	ts.service = service
	return &ts, nil
}

// GetData from the workingRange cells
func (ts *TableService) GetData(workingRange string) (*sheets.ValueRange, error) {
	return ts.service.Spreadsheets.Values.Get(ts.SpreadsheetID, workingRange).Do()
}

// UpdateData in the workingRange cells
func (ts *TableService) UpdateData(workingRange string, resultRange *sheets.ValueRange) (*sheets.UpdateValuesResponse, error) {
	return ts.service.Spreadsheets.Values.Update(ts.SpreadsheetID, workingRange, resultRange).ValueInputOption("RAW").Do()
}

func (ts *TableService) getConfig(properties *ConnectionProperties) (*oauth2.Config, error) {
	scope := "https://www.googleapis.com/auth/spreadsheets"
	if properties.ClientID != "" && properties.ProjectID != "" && properties.AuthURI != "" &&
	properties.TokenURI != "" && properties.ClientSecret != "" && properties.RedirectUris != "" {
		return &oauth2.Config {
			ClientID:     properties.ClientID,
			ClientSecret: properties.ClientSecret,
			RedirectURL:  properties.RedirectUris,
			Scopes:       []string{scope},
			Endpoint: oauth2.Endpoint{
				AuthURL:  properties.AuthURI,
				TokenURL: properties.TokenURI,
			},
		}, nil
	}
	credentialsBytes, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Printf("Unable to read client secret file: %v", err)
		return nil, err
	}
	config, err := google.ConfigFromJSON(credentialsBytes, scope)
	if err != nil {
		log.Printf("Unable to parse client secret file to config: %v", err)
		return nil, err
	}
	return config, nil
}

func (ts *TableService) getClient(properties *ConnectionProperties, config *oauth2.Config) (*http.Client, error) {
	if properties.AccessToken != "" && properties.TokenType != "" && properties.RefreshToken != "" && properties.ExpireTime != "" {
		expiry, _ := time.Parse(time.RFC3339, properties.ExpireTime)
		token := &oauth2.Token{
			AccessToken:  properties.AccessToken,
			TokenType:    properties.TokenType,
			RefreshToken: properties.RefreshToken,
			Expiry:       expiry,
		}
		return config.Client(context.Background(), token), nil
	}
	tokenFile := "token.json"
	token, err := ts.tokenFromFile(tokenFile)
	if err != nil {
		token, err = ts.getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		err := ts.saveToken(tokenFile, token)
		if err != nil {
			return nil, err
		}
	}
	return config.Client(context.Background(), token), nil
}

func (ts *TableService) getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the " +
		"authorization code: \n%v\n", authURL)
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Printf("Unable to read authorization code: %v", err)
		return nil, err
	}
	token, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Printf("Unable to retrieve token from web: %v", err)
		return nil, err
	}
	return token, nil
}

func (ts *TableService) tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func (ts *TableService) saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Printf("Unable to cache oauth token: %v", err)
		return err
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
	return nil
}
