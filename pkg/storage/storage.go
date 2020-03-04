package storage

import (
	"github.com/danielstefank/kleinanzeigen-alert/pkg/scraper"
	"github.com/rs/xid"
)

type Query struct {
	Id       string
	ChatId   int64
	LastAds  []scraper.Ad
	Term     string
	Radius   int
	City     int
	CityName string
}

//TODO: use actual db
type Storage struct {
	db []Query
}

func (s *Storage) AddQuery(query Query) {
	s.db = append(s.db, query)
}

func (s *Storage) GetQueries() []Query {
	return s.db
}

func (s *Storage) ListForChatId(chatId int64) []Query {
	foundQueries := make([]Query, 0, 0)

	for _, q := range s.db {
		if q.ChatId == chatId {
			foundQueries = append(foundQueries, q)
		}
	}

	return foundQueries
}

func (s *Storage) findQueryById(id string) (int, *Query) {
	for idx, q := range s.db {
		if q.Id == id {
			return idx, &q
		}
	}
	return -1, nil
}

func (s *Storage) RemoveById(id string) *Query {
	for idx, item := range s.db {
		if item.Id == id {
			s.db = append(s.db[:idx], s.db[idx+1:]...)
			return &item
		}
	}

	return nil
}

func NewStorage() *Storage {
	s := new(Storage)
	s.db = make([]Query, 0, 0)

	return s
}

func NewQuery(term string, city string, radius int, chatId int64) (*Query, string) {
	cityId, cityName := scraper.FindCityId(city)

	if cityId == 0 {
		return nil, "could not find cityid"
	}

	q := new(Query)
	q.Id = xid.New().String()
	q.LastAds = scraper.GetAds(1, term, cityId, radius)
	q.ChatId = chatId
	q.CityName = cityName
	q.Term = term
	q.City = cityId
	q.Radius = radius
	return q, ""
}

func (s *Storage) GetLatest(id string) []scraper.Ad {

	idx, q := s.findQueryById(id)

	if q == nil {
		return make([]scraper.Ad, 0, 0)
	}

	latest := q.getAds()
	diff := findDiff(latest, q.LastAds)
	s.db[idx].LastAds = latest
	return diff
}

func (q Query) getAds() []scraper.Ad {
	return scraper.GetAds(1, q.Term, q.City, q.Radius)
}

func findDiff(arr1 []scraper.Ad, arr2 []scraper.Ad) []scraper.Ad {
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
