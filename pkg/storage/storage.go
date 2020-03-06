package storage

import (
	"log"
	"time"

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
		log.Panic(err.Error())
	}

	db.AutoMigrate(&model.Query{})
	db.AutoMigrate(&model.Ad{})

	s.db = db
	return s
}

//CloseDB closes the created tb connection
func (s *Storage) CloseDB() {
	s.db.Close()
}

// AddNewQuery adds a new query to the db
func (s *Storage) AddNewQuery(term string, city string, radius int, chatID int64) (*model.Query, string) {
	cityID, cityName := scraper.FindCityID(city)
	if cityID == 0 {
		return nil, "could not find cityid"
	}

	query := model.Query{ChatID: chatID, Term: term, Radius: radius, City: cityID, CityName: cityName}

	s.db.NewRecord(query)
	s.db.Create(&query)

	latestAds := scraper.GetAds(1, term, cityID, radius)
	s.storeLatestAds(latestAds, query.ID)

	return &query, ""
}

// GetQueries gets all the queries from the db
func (s *Storage) GetQueries() []model.Query {
	queries := make([]model.Query, 0, 0)
	s.db.Find(&queries)

	return queries
}

// ListForChatID gets all the queries for specified chatId
func (s *Storage) ListForChatID(chatID int64) []model.Query {
	queries := make([]model.Query, 0, 0)
	s.db.Where(&model.Query{ChatID: chatID}).Find(&queries)
	return queries
}

// FindQueryByID fidn a query by the given id
func (s *Storage) FindQueryByID(id uint) *model.Query {
	q := model.Query{}
	s.db.Where("id = ?", id).First(&q)
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
func (s *Storage) DeleteOlderAds() {
	s.db.Where("created_at < ?", time.Now().AddDate(0, 0, -7)).Delete(model.Ad{})
}

func (s *Storage) storeLatestAds(ads []scraper.Ad, qID uint) {
	for _, item := range ads {
		ad := model.Ad{EbayID: item.ID, QueryID: qID}
		s.db.NewRecord(&ad)
		s.db.Create(&ad)
	}
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
