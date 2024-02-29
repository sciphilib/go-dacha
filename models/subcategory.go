package models

type Subcategory struct {
	ID         uint     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name       string   `json:"name"`
	CategoryID uint     `json:"category_id"`
	Category   Category `gorm:"foreignKey:CategoryID" json:"-"`
}
