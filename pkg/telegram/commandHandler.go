package telegram

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"

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
		log.Panic().Err(err).Msg("could not initialize bot")
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
			log.Error().Err(err).Msg("could not get latetst message updates for telegram bot")
			continue
		}

		for update := range updates {

			if update.Message == nil { // ignore any non-Message updates
				continue
			}

			log.Debug().
				Str("telegram_msg", update.Message.Text).
				Str("telegram_username", update.Message.Chat.UserName).
				Msg("Got new message")

			switch update.Message.Command() {
			case "start":
				log.Debug().Str("telegram_username", update.Message.Chat.UserName).Msg("Starting bot.")
				b.sendMsgRaw(generateHelpText(), update.Message.Chat.ID)
			case "help":
				b.sendMsgRaw(generateHelpText(), update.Message.Chat.ID)
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
						msg = "Um eine Suche hinzuzufügen schreibe <code>/add {Suchbegriff}, {Stadt/PLZ}, {Radius}, {Max Preis ohne \"€\", \",\",\".\"}, {Min Preis ohne \"€\", \",\",\".\"}?</code>"
					} else {
						msg = fmt.Sprintf("Suche für <b>%s</b> in <b>%s</b> hinzugefügt. ID: <b>%d</b>", q.Term, q.CityName, q.ID)
						log.Debug().
							Str("telegram_username", update.Message.Chat.UserName).
							Str("term", q.Term).
							Str("city", q.CityName).
							Int("radius", q.Radius).
							Msg("added new query.")
					}

					b.sendMsgRaw(msg, update.Message.Chat.ID)
				}()
			case "remove":
				go func() {
					msg := "success"
					args := update.Message.CommandArguments()

					if len(args) == 0 {
						msg = "Um zu entfernen schreibe <code>/remove {ID}</code>. Die ID bekommst du vom <code>/list</code> Befehl."
					} else {
						id, err := strconv.ParseUint(strings.Trim(args, " "), 10, 0)

						if err != nil {
							msg = "Konnte ID nicht lesen. Diese sollte eine ganze positive Zahl sein."
						} else {
							removedQ := b.storage.RemoveByID(uint(id), update.Message.Chat.ID)
							if removedQ == nil {
								msg = "Suche nicht gefunden."
							} else {
								msg = fmt.Sprintf("Suche für %s entfernt", removedQ.Term)
								log.Debug().
									Str("telegram_username", update.Message.Chat.UserName).
									Str("term", removedQ.Term).
									Str("city", removedQ.CityName).
									Int("radius", removedQ.Radius).
									Msg("query removed")
							}
						}

					}

					b.sendMsgRaw(msg, update.Message.Chat.ID)
				}()
			case "clear":
				go func() {
					b.sendMsgRaw("kommt bald.", update.Message.Chat.ID)
				}()
			default:
				b.sendMsgRaw("Das Kommando kenne ich nicht.", update.Message.Chat.ID)
			}

			lastUpdateID = update.UpdateID
		}
	}
}

// SendAds send the given ad the the given chatId
func (b *Bot) SendAds(chatID int64, ads []scraper.Ad, q model.Query) error {
	for _, ad := range ads {
		err := b.sendMsg(formatAd(ad, q.Term, int(q.ID)), formatAdRaw(ad, q.Term, int(q.ID)), chatID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *Bot) sendQueries(chatID int64, queries []model.Query) {
	if len(queries) == 0 {
		b.sendMsg(
			"Keine Suchen gefunden. Füge ein mit <code>/add</code> hinzu.",
			"Keine Suchen gefunden. Füge ein mit /add hinzu.",
			chatID)
	} else {
		for _, q := range queries {
			b.sendMsg(formatQuery(q), formatQueryRaw(q), chatID)
		}
	}
}

func (b *Bot) sendMsgRaw(msg string, chatID int64) error {
	return b.sendMsg(msg, msg, chatID)
}

func (b *Bot) sendMsg(msg string, raw string, chatID int64) error {
	telegramMessage := tgbotapi.NewMessage(chatID, msg)
	telegramMessage.ParseMode = tgbotapi.ModeHTML

	_, err := b.internalBot.Send(telegramMessage)

	if err != nil {
		if err.Error() == blocked {
			log.Info().Msg("the bot was blocked by the user. could not send message.")
			return errors.New("user blocked the bot")
		}

		if err.Error() == deactivated {
			log.Info().Msg("the bot was deactivated. could not send message.")
			return errors.New("user is deactivated")
		}

		if strings.HasPrefix(err.Error(), "Bad Request: can't parse entities") {
			log.Info().Msg("msg has invalid html. trying to send raw data.")
			telegramMessage := tgbotapi.NewMessage(chatID, raw)

			_, err := b.internalBot.Send(telegramMessage)

			if err != nil {
				log.Warn().Err(err).Str("send_message", msg).Msg("could not send telegram message")
			}
		}
	}

	return nil
}

func formatQuery(q model.Query) string {
	var b strings.Builder
	f := fmt.Sprintf
	b.WriteString(f("Suchbegriff: <b>%s</b>\n", q.Term))
	b.WriteString(f("Radius: <b>%v km</b>\n", q.Radius))
	b.WriteString(f("Stadt: <b>%s</b>\n", q.CityName))
	b.WriteString(f("ID: <b>%v</b>", q.ID))

	if q.MaxPrice != nil {
		b.WriteString(f("\nMax Preis: <b>%v €</b>", *q.MaxPrice))
	}

	if q.MinPrice != nil {
		b.WriteString(f("\nMin Preis: <b>%v €</b>", *q.MinPrice))
	}

	return b.String()
}

func formatQueryRaw(q model.Query) string {
	var b strings.Builder
	f := fmt.Sprintf
	b.WriteString(f("Suchbegriff: %s\n", q.Term))
	b.WriteString(f("Radius: %v km\n", q.Radius))
	b.WriteString(f("Stadt: %s\n", q.CityName))
	b.WriteString(f("ID: %v", q.ID))

	if q.MaxPrice != nil {
		b.WriteString(f("Max Preis: %v €", q.MaxPrice))
	}

	if q.MinPrice != nil {
		b.WriteString(f("Max Preis: %v €", q.MinPrice))
	}

	return b.String()
}

func formatAd(ad scraper.Ad, term string, id int) string {
	var b strings.Builder
	f := fmt.Sprintf
	b.WriteString(f("<b>%s</b> - %s\n", ad.Title, ad.Price))
	b.WriteString(f("For search \"%s\" (ID: %v)\n", term, id))
	b.WriteString(f("<a href=\"%s\">hier</a>", ad.Link))

	return b.String()
}

func formatAdRaw(ad scraper.Ad, term string, id int) string {
	var b strings.Builder
	f := fmt.Sprintf
	b.WriteString(f("%s - %s\n", ad.Title, ad.Price))
	b.WriteString(f("For search \"%s\" (ID: %v)\n", term, id))
	b.WriteString(f("Link: %s", ad.Link))

	return b.String()
}

func getQueryFromArgs(args string, chatID int64, s *storage.Storage) (*model.Query, bool) {
	arr := strings.SplitN(args, ",", -1)

	if len(arr) < 2 || len(arr) > 5 {
		return nil, false
	}

	term := arr[0]
	city := arr[1]

	radius, err := strconv.Atoi(strings.Trim(arr[2], " "))
	if err != nil {
		return nil, false
	}

	var q *model.Query

	if len(arr) > 3 {
		price, err := strconv.Atoi(strings.Trim(arr[3], " "))

		if err != nil {
			return nil, false
		}

		if len(arr) > 4 {
			minPrice, err := strconv.Atoi(strings.Trim(arr[4], " "))

			if err != nil {
				return nil, false
			}

			q, err = s.AddNewQuery(term, city, radius, &price, &minPrice, chatID)
		} else {
			q, err = s.AddNewQuery(term, city, radius, &price, nil, chatID)
		}

	} else {
		q, err = s.AddNewQuery(term, city, radius, nil, nil, chatID)
	}

	if err != nil {
		log.Warn().Err(err).
			Str("term", q.Term).
			Str("city", q.CityName).
			Int("radius", q.Radius).
			Msg("could not create query")

		return nil, false
	}

	return q, true
}

func generateHelpText() string {
	var b strings.Builder
	f := fmt.Sprintf
	b.WriteString(f("<u>Hinzufügen von Suchen</u>\n"))
	b.WriteString(f("schreibe <code>/add {Suchbegriff}, {Stadt/PLZ}, {Radius}, {Max Preis ohne \"€\", \",\",\".\"}?, {Min Preis ohne \"€\", \",\",\".\"}?</code>\n"))
	b.WriteString(f("z.B. <code>/add Fahrrad, Köln, 20</code>\n"))
	b.WriteString(f("Dies führt jede minute eine Suche aus und du kommst die neuesten Einträge hier.\n"))

	b.WriteString(f("\n"))
	b.WriteString(f("<u>Listen von alles Suchen</u>\n"))
	b.WriteString(f("schreibe <code>/list</code>\n"))
	b.WriteString(f("Dies listet alle deine aktuellen Suchen\n"))

	b.WriteString(f("\n"))
	b.WriteString(f("<u>Entfernen von Suchen</u>\n"))
	b.WriteString(f("schreibe <code>/remove {ID}</code>\n"))
	b.WriteString(f("Die ID erhältst du aus dem List Befehl. Dies Löscht die Suche und du erhältst für sie keine Nachrichten mehr.\n"))

	return b.String()
}
