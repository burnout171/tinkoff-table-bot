package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"google.golang.org/api/sheets/v4"
)

var spreadsheetID = os.Getenv("SHEET_ID")

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

func updateTable(connection *sheets.Service, input string) (int64, error) {
	receivedKey, sum := parseInput(input)
	month, day := currentDate()
	workingRange := fmt.Sprintf("%s!H%d:I%d", month, day+1, day+1)
	receivedRange, err := connection.Spreadsheets.Values.Get(spreadsheetID, workingRange).Do()
	if err != nil {
		return -1, err
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
	updateResponse, err := connection.Spreadsheets.Values.Update(spreadsheetID, workingRange, &resultRange).ValueInputOption("RAW").Do()
	if (err != nil) {
		return -1, err
	}
	return updateResponse.UpdatedCells, nil
}

func getDailyBalance(connection *sheets.Service) (string, error) {
	month, day := currentDate()
	workingRange := fmt.Sprintf("%s!K%d", month, day+1)
	return getSimpleSheetData(connection, workingRange)
}

func getMonthlyBalance(connection *sheets.Service) (string, error) {
	month, _ := currentDate()
	workingRange := fmt.Sprintf("%s!K33", month)
	return getSimpleSheetData(connection, workingRange)
}

func getMonthlyAccumulation(connection *sheets.Service) (string, error) {
	month, _ := currentDate()
	workingRange := fmt.Sprintf("%s!D21", month)
	return getSimpleSheetData(connection, workingRange)
}

func getSimpleSheetData(connection *sheets.Service, workingRange string) (string, error) {
	receivedRange, err := connection.Spreadsheets.Values.Get(spreadsheetID, workingRange).Do()
	if err != nil {
		return "", err
	}
	return receivedRange.Values[0][0].(string), nil
}

func processCommand(connection *sheets.Service, command string) (string, error) {
	switch command {
	case "db":
		return getDailyBalance(connection)
	case "mb":
		return getMonthlyBalance(connection)
	case "ma":
		return getMonthlyAccumulation(connection)
	default:
		return "Unknown command", nil
	}
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		log.Fatalf("Could not connect to telegram: %v", err)
	}
	if debug, _ := strconv.ParseBool(os.Getenv("ENABLE_DEBUG")); debug == true {
		bot.Debug = true
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	tableService := TableService{}
	connection := tableService.CreateConnection()
	log.Printf("Authorized on account %s", bot.Self.UserName)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		if update.Message.IsCommand() {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
			text, err := processCommand(connection, update.Message.Command())
			if (err != nil) {
				log.Printf("Following error accured: %v", err)
				msg.Text = "Some error accured"
			} else {
				msg.Text = text
			}
			bot.Send(msg)
			continue
		}
		_, err := updateTable(connection, update.Message.Text)
		var replyMessage tgbotapi.MessageConfig
		if (err != nil) {
			log.Printf("Following error accured: %v", err)
			replyMessage = tgbotapi.NewMessage(update.Message.Chat.ID, "Some error accured")
		} else {
			balance, err := getDailyBalance(connection)
			var replyText string
			if (err != nil) {
				replyText = "Баланс обновлен"
			} else {
				replyText = "Остаток на день " + balance
			}
			replyMessage = tgbotapi.NewMessage(update.Message.Chat.ID, replyText)
		}
		replyMessage.ReplyToMessageID = update.Message.MessageID
		bot.Send(replyMessage)
	}
}
