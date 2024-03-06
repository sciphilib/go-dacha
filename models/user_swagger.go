package models

// swagger:model UserResponse
type UserResponse struct {
	ID          uint         `json:"id"`
	Name        string       `json:"name"`
	Email       string       `json:"email"`
	Location    UserLocation `json:"location"`
	PhoneNumber string       `json:"phone_number"`
}

// swagger:model UserUpdate
type UserUpdateSwagger struct {
	Name        string       `json:"name"`
	Location    UserLocation `json:"location"`
	PhoneNumber string       `json:"phone_number"`
}

// Location represents a geographical coordinate.
// swagger:model Location
type UserLocation struct {
	// Coordinates is an array of two float numbers.
	Type string `json:"type"`
	// Example: [123.45, 67.89]
	Coordinates [2]float64 `json:"coordinates"`
}
