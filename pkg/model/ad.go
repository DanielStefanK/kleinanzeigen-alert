package model

import "time"

// Ad that is beeing stored
type Ad struct {
	ID        uint   `gorm:"primary_key"`
	EbayID    string `gorm:"type:varchar(255)"`
	QueryID   uint   `gorm:"index:ad_queryid"`
	Location  string `gorm:"type:varchar(510)"`
	CreatedAt time.Time
}
