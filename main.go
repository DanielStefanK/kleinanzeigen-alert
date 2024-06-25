package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/danielstefank/kleinanzeigen-alert/pkg/model"

	"github.com/danielstefank/kleinanzeigen-alert/pkg/storage"
	"github.com/danielstefank/kleinanzeigen-alert/pkg/telegram"
)

var token string

var f = fmt.Sprintf

const fetchDuration = time.Second * 60

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	debug := flag.Bool("debug", false, "sets log level to debug")

	flag.Parse()

	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger().With().Timestamp().Logger()
	} else {
		log.Logger = log.With().Caller().Logger().With().Timestamp().Logger()
	}

	token = os.Getenv("TELEGRAM_APITOKEN")

	if token == "" {
		log.Panic().Msg("could read API token")
		os.Exit(1)
	}

	s := storage.NewStorage()
	defer s.CloseDB()
	bot := telegram.CreateBot(token, s)
	bot.Init()
	go bot.Start()

	cleanupTicker := time.NewTicker(time.Hour)

	go func() {
		for {
			<-cleanupTicker.C
			log.Info().Msg("Removing old ads.")
			deleted, err := s.DeleteOlderAds()

			if err != nil {
				log.Error().Err(err).Msg("could not delete old ads")
				return
			}

			log.Info().Int64("affected_ads", deleted).Msg("Old ads removed. Sleeping for 1 hour.")
		}
	}()

	for {
		queries := s.GetQueries()

		log.Info().Int("number_of_queries", len(queries)).Msg("fetching ads")
		for _, q := range queries {
			go func(query model.Query) {
				new, err := s.GetLatest(query.ID)

				if err != nil {
					if query.FailedPreviously {
						s.RemoveByID(query.ID, query.ChatID)
						bot.SendMsg(query.ChatID, f("Anzeigen für %s (ID: %d) konnten nicht geladen werden. Das Problem ist erneut aufgetreten. Die Query wurde gelöscht.", query.Term, query.ID))
					} else {
						bot.SendMsg(query.ChatID, f("Anzeigen konnten für %s (ID: %d) nicht geladen werden. Falls das Problem weiterhin besteht, wird die Query gelöscht. Der Bot könnte überlastet sein, oder die Query enthält Fehler.", query.Term, query.ID))
						query.FailedPreviously = true
						s.UpdateQuery(query.ID, true)
					}
					return
				}

				if query.FailedPreviously {
					s.UpdateQuery(query.ID, false)
				}

				log.Debug().Int("number_of_new_ads", len(new)).Msg("new ads found")
				err = bot.SendAds(query.ChatID, new, query)
				if err != nil {
					affected, err := s.RemoveByChatID(query.ChatID)
					if err != nil {
						log.Error().Err(err).
							Msg("could not remove  queries for blocked/deactivated user")
					} else {
						log.Info().
							Int("number_of_removed_queries", affected).
							Msg("removed queries for blocked/deactivated user")
					}
				}
			}(q)
		}

		time.Sleep(fetchDuration)
	}
}
