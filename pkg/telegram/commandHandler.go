package telegram

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/danielstefank/kleinanzeigen-alert/pkg/scraper"

	"github.com/danielstefank/kleinanzeigen-alert/pkg/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Bot struct {
	token       string
	internalBot *tgbotapi.BotAPI
	storage     *storage.Storage
}

func CreateBot(token string, storage *storage.Storage) *Bot {
	bot := new(Bot)
	bot.token = token
	bot.storage = storage
	return bot
}

func (b *Bot) Init() {
	bot, err := tgbotapi.NewBotAPI(b.token)

	b.internalBot = bot

	if err != nil {
		log.Panic("could not initalize bot")
		os.Exit(1)
	}
}

func (b *Bot) Start() {
	lastUpdateId := -1

	for {
		u := tgbotapi.NewUpdate(lastUpdateId + 1)
		u.Timeout = 60

		updates, err := b.internalBot.GetUpdatesChan(u)

		if err != nil {
			log.Println("could not get latetst update")
			continue
		}

		for update := range updates {
			if update.Message == nil { // ignore any non-Message updates
				continue
			}

			log.Printf("Got msg: %s\n", update.Message.Text)

			if !update.Message.IsCommand() {
				log.Printf("Got msg: %s\n", update.Message.Text)
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

			switch update.Message.Command() {
			case "start":
				msg.Text = "type '/list', '/add {SearchTerm}, {CityId}, {Radius}' or /remove {ID}."
				b.internalBot.Send(msg)
			case "help":
				msg.Text = "type '/list', '/add {SearchTerm}, {CityId}, {Radius}' or /remove {ID}."
				b.internalBot.Send(msg)
			case "list":
				go func() {
					queries := b.storage.ListForChatId(update.Message.Chat.ID)
					b.sendQueries(update.Message.Chat.ID, queries)
				}()
			case "add":
				go func() {
					msg := "success"
					q, success := getQueryFromArgs(update.Message.CommandArguments(), update.Message.Chat.ID)

					if !success {
						msg = "use add like this '/add {SearchTerm}, {CityId}, {Radius}'"
					} else {
						b.storage.AddQuery(*q)
						msg = fmt.Sprintf("Added query for %s", q.Term)
					}

					b.internalBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
				}()
			case "remove":
				go func() {
					msg := "success"
					args := update.Message.CommandArguments()

					if len(args) == 0 {
						msg = "use remove like this '/remove {ID}"
					} else {
						removedQ := b.storage.RemoveById(strings.Trim(args, " "))

						if removedQ == nil {
							msg = "Query not found"
						} else {
							msg = fmt.Sprintf("Removed query for %s", removedQ.Term)
						}
					}

					b.internalBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
				}()
			case "clear":
				go func() {
					b.internalBot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "coming soon"))
				}()
			default:
				msg.Text = "I don't know that command"
				b.internalBot.Send(msg)
			}

			lastUpdateId = update.UpdateID
		}
	}
}

func (b *Bot) SendAds(chatId int64, ads []scraper.Ad) {
	for _, ad := range ads {
		msg := tgbotapi.NewMessage(chatId, formatAd(ad))
		b.internalBot.Send(msg)
	}
}

func (b *Bot) sendQueries(chatId int64, queries []storage.Query) {
	if len(queries) == 0 {
		msg := tgbotapi.NewMessage(chatId, "No queries try adding one with /add")
		b.internalBot.Send(msg)
	} else {
		for _, q := range queries {
			msg := tgbotapi.NewMessage(chatId, formatQuery(q))
			msg.ReplyMarkup
			b.internalBot.Send(msg)
		}
	}
}

func formatQuery(q storage.Query) string {
	var b strings.Builder
	f := fmt.Sprintf
	b.WriteString(f("Term: %s\n", q.Term))
	b.WriteString(f("Radius: %v\n", q.Radius))
	b.WriteString(f("City: %v - %s\n", q.City, q.CityName))
	b.WriteString(f("ID: %s", q.Id))

	return b.String()
}

func formatAd(ad scraper.Ad) string {
	var b strings.Builder
	f := fmt.Sprintf
	b.WriteString(f("%s\n", ad.Title))
	b.WriteString(f("%s\n", ad.Price))
	b.WriteString(f("%s", ad.Link))

	return b.String()
}

func getQueryFromArgs(args string, chatId int64) (*storage.Query, bool) {
	arr := strings.SplitN(args, ",", -1)

	if len(arr) != 3 {
		return nil, false
	}

	term := arr[0]
	city := arr[1]

	radius, err := strconv.Atoi(strings.Trim(arr[2], " "))

	if err != nil {
		return nil, false
	}

	q, cityError := storage.NewQuery(term, city, radius, chatId)

	if cityError != "" {
		return nil, false
	}

	return q, true
}
