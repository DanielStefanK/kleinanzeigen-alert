package storage

import (
	"errors"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/danielstefank/kleinanzeigen-alert/pkg/model"
	"github.com/danielstefank/kleinanzeigen-alert/pkg/scraper"
	"github.com/jinzhu/gorm"

	// import for the database driver
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

// Storage is the main storage medium
type Storage struct {
	db *gorm.DB
}

// NewStorage creates a new Storage
func NewStorage() *Storage {
	s := new(Storage)
	db, err := gorm.Open("sqlite3", "/tmp/alert.db")

	if err != nil {
		log.Panic().Err(err).Msg("sqlite3 database could not be created")
	}

	log.Info().Msg("database was created")

	db.AutoMigrate(&model.Query{})
	db.AutoMigrate(&model.Ad{})

	s.db = db
	return s
}

//CloseDB closes the created tb connection
func (s *Storage) CloseDB() {
	log.Info().Msg("closing database")
	s.db.Close()
}

// AddNewQuery adds a new query to the db
func (s *Storage) AddNewQuery(term string, city string, radius int, chatID int64) (*model.Query, error) {
	cityID, cityName, err := scraper.FindCityID(city)

	if err != nil {
		return nil, errors.New("could not find city id")
	}

	query := model.Query{ChatID: chatID, Term: term, Radius: radius, City: cityID, CityName: cityName}

	//s.db.NewRecord(query)

	err = s.db.Create(&query).Error

	if err != nil {
		log.Error().Err(err).Msg("could not create query")
		return nil, errors.New("could not create query")
	}

	latestAds := scraper.GetAds(1, term, cityID, radius)
	err = s.storeLatestAds(latestAds, query.ID)

	if err != nil {
		log.Error().Err(err).Msg("could not store latest ads")
		return nil, errors.New("could not store latest ads")
	}

	return &query, nil
}

// GetQueries gets all the queries from the db
func (s *Storage) GetQueries() []model.Query {
	queries := make([]model.Query, 0, 0)
	err := s.db.Find(&queries).Error

	if err != nil {
		log.Error().Err(err).Msg("could not get queries")
		return queries
	}

	return queries
}

// ListForChatID gets all the queries for specified chatId
func (s *Storage) ListForChatID(chatID int64) []model.Query {
	queries := make([]model.Query, 0, 0)
	err := s.db.Where(&model.Query{ChatID: chatID}).Find(&queries).Error

	if err != nil {
		log.Error().Err(err).Msg("could not get queries for a specific chat id")
		return queries
	}

	return queries
}

// FindQueryByID find a query by the given id
func (s *Storage) FindQueryByID(id uint) *model.Query {
	q := model.Query{}
	err := s.db.Where("id = ?", id).First(&q).Error

	if err != nil {
		log.Error().Err(err).Msg("could not get a query by id")
		return &q
	}

	return &q
}

// RemoveByID removes a query by id
func (s *Storage) RemoveByID(id uint, chatID int64) *model.Query {
	q := s.FindQueryByID(id)

	if q.ChatID != chatID {
		return nil
	}

	s.db.Delete(q)

	return q
}

// RemoveByChatID removes all queries for a chat id
func (s *Storage) RemoveByChatID(chatID int64) (int, error) {
	trx := s.db.Where(&model.Query{ChatID: chatID}).Delete(&model.Query{})

	return int(trx.RowsAffected), trx.Error
}

//GetLatest fetches the latest ads from kleinanzeigen. All ads where the id is not in the db is returned and the db is updated with the latest ads
func (s *Storage) GetLatest(id uint) []scraper.Ad {
	q := s.FindQueryByID(id)

	if q == nil {
		return make([]scraper.Ad, 0, 0)
	}

	latest := scraper.GetAds(1, q.Term, q.City, q.Radius)
	diff := s.findDiff(latest, q.ID)

	s.storeLatestAds(diff, q.ID)
	return diff
}

// DeleteOlderAds deletes all ads older that 7 days
func (s *Storage) DeleteOlderAds() (int64, error) {
	trx := s.db.Where("created_at < ?", time.Now().AddDate(0, 0, -7)).Delete(model.Ad{})
	if trx.Error != nil {
		return 0, trx.Error
	}
	return trx.RowsAffected, nil
}

func (s *Storage) storeLatestAds(ads []scraper.Ad, qID uint) error {
	for _, item := range ads {
		ad := model.Ad{EbayID: item.ID, QueryID: qID}
		s.db.NewRecord(&ad)
		err := s.db.Create(&ad).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Storage) findDiff(current []scraper.Ad, qID uint) []scraper.Ad {
	newAds := make([]scraper.Ad, 0, 0)
	for _, s1 := range current {
		q := model.Ad{}
		s.db.Where("query_id = ? AND ebay_id = ?", qID, s1.ID).First(&q)
		if q.EbayID == "" {
			newAds = append(newAds, s1)
		}
	}

	return newAds
}
