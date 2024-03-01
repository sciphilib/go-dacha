package models

import (
	"time"

	"github.com/lib/pq"
	"github.com/sciphilib/go-dacha/common"
)

type Ad struct {
	ID           uint                    `gorm:"primaryKey;autoIncrement" json:"id"`
	Title        string                  `json:"title"`
	Price        string                  `json:"price"`
	Subcategory  SubcategoryWithCategory `json:"subcategory"`
	Description  string                  `json:"description"`
	User_id      uint                    `json:"user_id"`
	Datetime     time.Time               `json:"datetime"`
	Pictures     pq.StringArray          `json:"-"`
	PicturesText []string                `json:"pictures"`
	LocationText common.GeoJSONText      `json:"location" gorm:"-"`
	LocationEWKB []byte                  `gorm:"column:location" json:"-"`
}

type SubcategoryWithCategory struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	CategoryID uint   `json:"category_id"`
	Category   string `json:"category_name"`
}
