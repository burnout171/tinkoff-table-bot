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

// TableService aggregates table management functions.
type TableService struct {
}

// CreateConnection create a connection to manage the table.
func (ts *TableService) CreateConnection() *sheets.Service {
	config := ts.getConfig()
	client := ts.getClient(config)
	service, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}
	return service
}

func (ts *TableService) getConfig() *oauth2.Config {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	projectID := os.Getenv("GOOGLE_PROJECT_ID")
	authURI := os.Getenv("GOOGLE_AUTH_URI")
	tokenURI := os.Getenv("GOOGLE_TOKEN_URI")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectUris := os.Getenv("GOOGLE_REDIRECT_URIS")
	scope := "https://www.googleapis.com/auth/spreadsheets"
	if clientID != "" && projectID != "" && authURI != "" &&
		tokenURI != "" && clientSecret != "" && redirectUris != "" {
		return &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectUris,
			Scopes:       []string{scope},
			Endpoint: oauth2.Endpoint{
				AuthURL:  authURI,
				TokenURL: tokenURI,
			},
		}
	}
	credentialsBytes, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	config, err := google.ConfigFromJSON(credentialsBytes, scope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	return config
}

func (ts *TableService) getClient(config *oauth2.Config) *http.Client {
	accessToken := os.Getenv("SHEET_ACCESS_TOKEN")
	tokenType := os.Getenv("SHEET_TOKEN_TYPE")
	refreshToken := os.Getenv("SHEET_REFRESH_TOKEN")
	expireTime := os.Getenv("SHEET_TOKEN_EXPIRE_TIME")
	if accessToken != "" && tokenType != "" && refreshToken != "" && expireTime != "" {
		expiry, _ := time.Parse(time.RFC3339, expireTime)
		token := &oauth2.Token{
			AccessToken:  accessToken,
			TokenType:    tokenType,
			RefreshToken: refreshToken,
			Expiry:       expiry,
		}
		return config.Client(context.Background(), token)
	}
	tokenFile := "token.json"
	token, err := ts.tokenFromFile(tokenFile)
	if err != nil {
		token = ts.getTokenFromWeb(config)
		ts.saveToken(tokenFile, token)
	}
	return config.Client(context.Background(), token)
}

func (ts *TableService) getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the " +
		"authorization code: \n%v\n", authURL)
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}
	token, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return token
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

func (ts *TableService) saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
