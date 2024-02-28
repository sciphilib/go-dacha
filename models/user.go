package models

import (
	"github.com/sciphilib/go-dacha/common"
)

type User struct {
	ID           uint               `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string             `json:"name"`
	Email        string             `json:"email" gorm:"unique"`
	Pass_hash    string             `json:"-"`
	LocationText common.GeoJSONText `json:"location" gorm:"-"`
	LocationEWKB []byte             `gorm:"column:location" json:"-"`
	PhoneNumber  string             `json:"phone_number" gorm:"unique"`
}
