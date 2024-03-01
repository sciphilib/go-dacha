package models

import (
	"time"
)

// swagger:model AdResponse
type AdResponse struct {
	ID          uint          `json:"id"`
	Title       string        `json:"title"`
	Price       string        `json:"price"`
	Subcategory SubcategoryAd `json:"subcategory"` // Предполагается, что Subcategory - это структура с полями id, name и category
	User        UserAd        `json:"user"`        // Предполагается, что User - это структура с полями id, name, email, phone_number, и location
	Datetime    time.Time     `json:"datetime"`
	Pictures    []string      `json:"pictures"`
	Location    LocationAd    `json:"location"` // Предполагается, что Location - это структура с полями type и coordinates
}

// swagger:model UserAd
type UserAd struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	PhoneNumber string     `json:"phone_number"`
	Location    LocationAd `json:"location"`
}

// swagger:model SubcategoryAd
type SubcategoryAd struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

// Location represents a geographical coordinate.
// swagger:model Location
type LocationAd struct {
	// Coordinates is an array of two float numbers.
	Type string `json:"type"`
	// Example: [123.45, 67.89]
	Coordinates [2]float64 `json:"coordinates"`
}
