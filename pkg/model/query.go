package model

import (
	"github.com/jinzhu/gorm"
)

// Query that is beeing sored
type Query struct {
	gorm.Model
	ChatID           int64 `gorm:"index:chatid"`
	LastAds          []Ad
	Term             string `gorm:"type:varchar(100)"`
	Radius           int
	City             int
	CityName         string `gorm:"type:varchar(100)"`
	MaxPrice         *int
	MinPrice         *int
	CustomLink       *string `gorm:"type:varchar(1000)"`
	FailedPreviously bool
}

// AfterDelete delete all assiciated ads
func (u *Query) AfterDelete(tx *gorm.DB) (err error) {
	tx.Where("query_id = ?", u.ID).Delete(&Ad{})
	return
}
