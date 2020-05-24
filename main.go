package main

import (
	"log"
	"fmt"
	"net/http"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func configure() (*tgbotapi.BotAPI, tgbotapi.UpdatesChannel, *TableManagement) {
	token := os.Getenv("TELEGRAM_TOKEN")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Could not connect to telegram: %v", err)
	}
	if debug, _ := strconv.ParseBool(os.Getenv("ENABLE_DEBUG")); debug == true {
		bot.Debug = true
	}
	properties := &ConnectionProperties{
		SpreadsheetID: os.Getenv("SHEET_ID"),
		ClientID:      os.Getenv("GOOGLE_CLIENT_ID"),
		ProjectID:     os.Getenv("GOOGLE_PROJECT_ID"),
		AuthURI:       os.Getenv("GOOGLE_AUTH_URI"),
		TokenURI:      os.Getenv("GOOGLE_TOKEN_URI"),
		ClientSecret:  os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectUris:  os.Getenv("GOOGLE_REDIRECT_URIS"),
		AccessToken:   os.Getenv("SHEET_ACCESS_TOKEN"),
		TokenType:     os.Getenv("SHEET_TOKEN_TYPE"),
		RefreshToken:  os.Getenv("SHEET_REFRESH_TOKEN"),
		ExpireTime:    os.Getenv("SHEET_TOKEN_EXPIRE_TIME"),
	}
	tableService, err := NewTableService(properties)
	tableManagement := NewTableManagement(tableService)
	if "heroku" == os.Getenv("ENVIRONMENT") {
		bot.RemoveWebhook()
		publicURL := fmt.Sprintf("%s/%s", os.Getenv("URL"), token)
		_, err = bot.SetWebhook(tgbotapi.NewWebhook(publicURL))
		if err != nil {
			log.Fatalf("Could not register webhook: %v", err)
		}
		updates := bot.ListenForWebhook("/" + token)
		go http.ListenAndServe("0.0.0.0:" + os.Getenv("PORT"), nil)
		return bot, updates, tableManagement
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal("Could not init a connection to Telegram", err)
	}
	return bot, updates, tableManagement
}

func processCommand(tm *TableManagement, update *tgbotapi.Update) tgbotapi.MessageConfig {
	balance, err := tm.GetTableBalance(update.Message.Command())
	if err != nil {
		log.Printf("Following error accured: %v", err)
		return tgbotapi.NewMessage(update.Message.Chat.ID, "Some error accured")
	}
	return tgbotapi.NewMessage(update.Message.Chat.ID, balance)
}

func processUpdate(tm *TableManagement, update *tgbotapi.Update) tgbotapi.MessageConfig {
	_, err := tm.UpdateTableData(update.Message.Text)
	if err != nil {
		log.Printf("Following error accured: %v", err)
		return tgbotapi.NewMessage(update.Message.Chat.ID, "Some error accured")
	}
	balance, err := tm.GetTableBalance("db")
	var replyText string
	if err != nil {
		replyText = "Баланс обновлен"
	} else {
		replyText = "Остаток на день " + balance
	}
	log.Print("Ok: ", replyText)
	replyMessage := tgbotapi.NewMessage(update.Message.Chat.ID, replyText)
	replyMessage.ReplyToMessageID = update.Message.MessageID
	return replyMessage
}

func main() {
	bot, updates, tm := configure()
	log.Printf("Authorized on account %s", bot.Self.UserName)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		if update.Message.IsCommand() {
			msg := processCommand(tm, &update)
			bot.Send(msg)
			continue
		}
		replyMessage := processUpdate(tm, &update)
		bot.Send(replyMessage)
	}
}
