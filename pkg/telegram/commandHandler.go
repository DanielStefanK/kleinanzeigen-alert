package telegram

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/danielstefank/kleinanzeigen-alert/pkg/model"
	"github.com/danielstefank/kleinanzeigen-alert/pkg/scraper"

	"github.com/danielstefank/kleinanzeigen-alert/pkg/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Bot will store the token the internal telegram bto and the storage
type Bot struct {
	token       string
	internalBot *tgbotapi.BotAPI
	storage     *storage.Storage
}

// CreateBot will create a new bot with the given token and storage
func CreateBot(token string, storage *storage.Storage) *Bot {
	bot := new(Bot)
	bot.token = token
	bot.storage = storage
	return bot
}

// Init will create the internal bot
func (b *Bot) Init() {
	bot, err := tgbotapi.NewBotAPI(b.token)

	b.internalBot = bot

	if err != nil {
		log.Panic("could not initalize bot")
		os.Exit(1)
	}
}

// Start starts the bot and listens for commands this will not return
func (b *Bot) Start() {
	lastUpdateID := -1

	for {
		u := tgbotapi.NewUpdate(lastUpdateID + 1)
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
					queries := b.storage.ListForChatID(update.Message.Chat.ID)
					b.sendQueries(update.Message.Chat.ID, queries)
				}()
			case "add":
				go func() {
					msg := "success"
					q, success := getQueryFromArgs(update.Message.CommandArguments(), update.Message.Chat.ID, b.storage)

					if !success {
						msg = "use add like this '/add {SearchTerm}, {CityId}, {Radius}'"
					} else {
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
						id, err := strconv.ParseUint(strings.Trim(args, " "), 10, 0)

						if err != nil {
							msg = "could not parse ID"
						} else {
							removedQ := b.storage.RemoveByID(uint(id), update.Message.Chat.ID)
							if removedQ == nil {
								msg = "Query not found"
							} else {
								msg = fmt.Sprintf("Removed query for %s", removedQ.Term)
							}
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

			lastUpdateID = update.UpdateID
		}
	}
}

// SendAds send the given ad the the given chatid
func (b *Bot) SendAds(chatID int64, ads []scraper.Ad) {
	for _, ad := range ads {
		msg := tgbotapi.NewMessage(chatID, formatAd(ad))
		b.internalBot.Send(msg)
	}
}

func (b *Bot) sendQueries(chatID int64, queries []model.Query) {
	if len(queries) == 0 {
		msg := tgbotapi.NewMessage(chatID, "No queries try adding one with /add")
		b.internalBot.Send(msg)
	} else {
		for _, q := range queries {
			msg := tgbotapi.NewMessage(chatID, formatQuery(q))
			b.internalBot.Send(msg)
		}
	}
}

func formatQuery(q model.Query) string {
	var b strings.Builder
	f := fmt.Sprintf
	b.WriteString(f("Term: %s\n", q.Term))
	b.WriteString(f("Radius: %v\n", q.Radius))
	b.WriteString(f("City: %v - %s\n", q.City, q.CityName))
	b.WriteString(f("ID: %v", q.ID))

	return b.String()
}

func formatAd(ad scraper.Ad) string {
	var b strings.Builder
	f := fmt.Sprintf
	b.WriteString(f("%s\n", ad.Title))
	b.WriteString(f("%s\n", ad.Price))
	b.WriteString(f("%s\n", ad.ID))
	b.WriteString(f("%s", ad.Link))

	return b.String()
}

func getQueryFromArgs(args string, chatID int64, s *storage.Storage) (*model.Query, bool) {
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

	q, cityError := s.AddNewQuery(term, city, radius, chatID)

	if cityError != "" {
		return nil, false
	}

	return q, true
}

func generateRoveBtn(id string) tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(fmt.Sprintf("/remove %s", id)),
		),
	)

}
