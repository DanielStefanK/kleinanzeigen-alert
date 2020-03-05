package main

import (
	"log"
	"os"
	"time"

	"github.com/danielstefank/kleinanzeigen-alert/pkg/model"

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
	defer s.CloseDB()
	bot := telegram.CreateBot(token, s)
	bot.Init()
	go bot.Start()

	for {
		log.Output(1, "Fetching ads")
		for _, q := range s.GetQueries() {
			go func(query model.Query) {
				new := s.GetLatest(query.ID)
				bot.SendAds(query.ChatID, new)
			}(q)
		}

		time.Sleep(time.Minute)
	}
}
