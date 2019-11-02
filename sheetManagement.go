package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"google.golang.org/api/sheets/v4"
)

var spreadsheetID = os.Getenv("SHEET_ID")

func getConfig() *oauth2.Config {
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

func getClient(config *oauth2.Config) *http.Client {
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
	token, err := tokenFromFile(tokenFile)
	if err != nil {
		token = getTokenFromWeb(config)
		saveToken(tokenFile, token)
	}
	return config.Client(context.Background(), token)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the " +
		"authorization code: \n%v\n", authURL)
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}
	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func currentDate() (monthName string, day int) {
	_, month, day := time.Now().Date()
	switch month {
	case time.January:
		return "Январь", day
	case time.February:
		return "Февраль", day
	case time.March:
		return "Март", day
	case time.April:
		return "Апрель", day
	case time.May:
		return "Май", day
	case time.June:
		return "Июнь", day
	case time.July:
		return "Июль", day
	case time.August:
		return "Август", day
	case time.September:
		return "Сентябрь", day
	case time.October:
		return "Октябрь", day
	case time.November:
		return "Ноябрь", day
	case time.December:
		return "Декабрь", day
	default:
		return "", day
	}
}

func parseInput(input string) (description string, sum float64) {
	splitted := strings.Split(input, " ")
	var descriptionSlice []string
	for _, word := range splitted {
		if value, err := strconv.ParseFloat(word, 64); err == nil {
			sum += value
			continue
		}
		descriptionSlice = append(descriptionSlice, word)
	}
	return strings.Join(descriptionSlice, ", "), sum
}

func prepareKey(receivedKey string, currentKey string) string {
	if strings.Contains(currentKey, " + ") {
		currentKey = strings.Replace(currentKey, " + ", ", ", -1)
	}
	if receivedKey == "" {
		return currentKey
	}
	return currentKey + ", " + receivedKey
}

func prepareValue(sum float64, currentValue string) float64 {
	if strings.HasPrefix(currentValue, "SUM") { // In case SUM() function is used in Sheet to sum the exchanges
		currentValue = currentValue[4 : len(currentValue)-2]
		for _, number := range strings.Split(currentValue, ", ") {
			if value, err := strconv.ParseFloat(number, 64); err == nil {
				sum += value
				continue
			}
		}
		return sum
	}
	currentValue = strings.ReplaceAll(currentValue, ",", "")
	floatValue, _ := strconv.ParseFloat(currentValue, 64)
	return sum + floatValue
}

func initServiceConnection() *sheets.Service {
	config := getConfig()
	client := getClient(config)
	service, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}
	return service
}

func updateTable(connection *sheets.Service, input string) error {
	receivedKey, sum := parseInput(input)
	month, day := currentDate()
	workingRange := fmt.Sprintf("%s!H%d:I%d", month, day+1, day+1)
	receivedRange, err := connection.Spreadsheets.Values.Get(spreadsheetID, workingRange).Do()
	if err != nil {
		return err
	}
	var resultRange sheets.ValueRange
	var myValues []interface{}
	if len(receivedRange.Values) == 0 {
		myValues = []interface{}{strings.ToLower(receivedKey), sum}
	} else {
		key := prepareKey(receivedKey, receivedRange.Values[0][0].(string))
		value := prepareValue(sum, receivedRange.Values[0][1].(string))
		myValues = []interface{}{strings.ToLower(key), value}
	}
	resultRange.Values = append(resultRange.Values, myValues)
	log.Println(resultRange.Values)
	if _, err = connection.Spreadsheets.Values.Update(spreadsheetID, workingRange, &resultRange).ValueInputOption("RAW").Do(); err != nil {
		return err
	}
	return nil
}

func getDailyBalance(connection *sheets.Service) string {
	month, day := currentDate()
	workingRange := fmt.Sprintf("%s!K%d", month, day+1)
	return getSheetData(connection, workingRange)
}

func getMonthlyBalance(connection *sheets.Service) string {
	month, _ := currentDate()
	workingRange := fmt.Sprintf("%s!K33", month)
	return getSheetData(connection, workingRange)
}

func getMonthlyAccumulation(connection *sheets.Service) string {
	month, _ := currentDate()
	workingRange := fmt.Sprintf("%s!D21", month)
	return getSheetData(connection, workingRange)
}

func getSheetData(connection *sheets.Service, workingRange string) string {
	receivedRange, err := connection.Spreadsheets.Values.Get(spreadsheetID, workingRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}
	return receivedRange.Values[0][0].(string)
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		log.Fatalf("Could not connect to telegram: %v", err)
	}
	if debug, _ := strconv.ParseBool(os.Getenv("ENABLE_DEBUG")); debug == true {
		bot.Debug = true
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	connection := initServiceConnection()
	for update := range updates {
		if update.Message == nil {
			continue
		}
		if update.Message.IsCommand() {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			switch update.Message.Command() {
			case "db":
				msg.Text = getDailyBalance(connection)
			case "mb":
				msg.Text = getMonthlyBalance(connection)
			case "ma":
				msg.Text = getMonthlyAccumulation(connection)
			default:
				msg.Text = "Unknown command"
			}
			bot.Send(msg)
			continue
		}
		err := updateTable(connection, update.Message.Text)
		var replyMessage tgbotapi.MessageConfig
		if err != nil {
			log.Printf("Following error accured: %v", err)
			replyMessage = tgbotapi.NewMessage(update.Message.Chat.ID, "Some error accured")
		} else {
			replyText := "Остаток на день " + getDailyBalance(connection)
			replyMessage = tgbotapi.NewMessage(update.Message.Chat.ID, replyText)
		}
		replyMessage.ReplyToMessageID = update.Message.MessageID
		bot.Send(replyMessage)
	}
}
