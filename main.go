package main

import (
	"log"
	"os"
	"time"

	"github.com/danielstefank/kleinanzeigen-alert/pkg/storage"
	"github.com/danielstefank/kleinanzeigen-alert/pkg/telegram"
)

var token string

func main() {
	token = os.Getenv("TELEGRAM_APITOKEN")

	if token == "" {
		log.Panic("could read API token")
		os.Exit(1)
	}

	s := storage.NewStorage()
	bot := telegram.CreateBot(token, s)
	bot.Init()
	go bot.Start()

	for true {
		log.Output(1, "Fetching ads")
		for _, q := range s.GetQueries() {
			go func(query storage.Query) {
				new := s.GetLatest(q.Id)
				bot.SendAds(q.ChatId, new)
			}(q)
		}

		time.Sleep(time.Minute)
	}
}
