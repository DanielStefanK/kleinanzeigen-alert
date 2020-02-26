package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/danielstefank/kleinanzeigen-alert/pkg/scraper"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Config struct {
	Token  string
	ChatId int64
}

var token string
var mychatId int64

func main() {
	file, _ := os.Open("../configs/config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := Config{}
	err := decoder.Decode(&config)
	if err != nil {
		log.Panic(err)
	}
	token = config.Token
	mychatId = config.ChatId

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	lastAds := scraper.GetAds(1, "nas server", 1932, 200)

	for true {
		ads := scraper.GetAds(1, "nas server", 1932, 200)
		newlyAdded := findNew(ads, lastAds)
		lastAds = append(ads[:0:0], ads...)

		if len(newlyAdded) > 0 {
			sendUpdate(newlyAdded, bot)
		}

		time.Sleep(time.Minute)
	}
}

func findNew(arr1 []scraper.Ad, arr2 []scraper.Ad) []scraper.Ad {
	newAds := make([]scraper.Ad, 0, 0)
	for _, s1 := range arr1 {
		found := false
		for _, s2 := range arr2 {
			if s1.Id == s2.Id {
				found = true
				break
			}
		}
		if !found {
			newAds = append(newAds, s1)
		}
	}

	return newAds
}

func sendUpdate(ads []scraper.Ad, bot *tgbotapi.BotAPI) {

	for _, ad := range ads {

		msg := tgbotapi.NewMessage(mychatId, formatAd(ad))
		bot.Send(msg)
	}
}

func formatAd(ad scraper.Ad) string {
	var b strings.Builder
	f := fmt.Sprintf
	b.WriteString(f("__%s__\n", ad.Title))
	b.WriteString(f("%s\n", ad.Price))
	b.WriteString(f("%s", ad.Link))

	return b.String()
}
