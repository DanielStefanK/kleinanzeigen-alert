package model

import "time"

// Ad that is beeing stored
type Ad struct {
	ID        uint   `gorm:"primary_key"`
	EbayID    string `gorm:"type:varchar(255)"`
	QueryID   uint   `gorm:"index:ad_queryid"`
	CreatedAt time.Time
}
