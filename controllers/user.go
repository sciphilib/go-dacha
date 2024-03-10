package controllers

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/ewkb"
	"github.com/paulmach/orb/geojson"
	"github.com/sciphilib/go-dacha/common"
	"github.com/sciphilib/go-dacha/models"
	"github.com/sciphilib/go-dacha/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	validate *validator.Validate
)

const (
	signingKey = "ldkfjalksdjflksj#32141#@@$!@"
)

type TokenClaims struct {
	UserId int `json:"id"`
}

type UserInput struct {
	Name        string            `json:"name" validate:"required"`
	Email       string            `json:"email" validate:"required,email"`
	Password    string            `json:"password" validate:"required"`
	Location    *geojson.Geometry `json:"location" validate:""`
	PhoneNumber string            `json:"phone_number" validate:"required"`
}

type UserUpdate struct {
	Name        string            `json:"name" validate:"required"`
	Location    *geojson.Geometry `json:"location" validate:""`
	PhoneNumber string            `json:"phone_number" validate:"required"`
}

// GetAllUsers godoc
// @Summary Get all users
// @Description Retrieves a list of all users with their locations in GeoJSON format
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} models.UserResponse "A list of users"
// @Failure 500 {object} string "Internal Server Error"
// @Router /users [get]
func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	var result []struct {
		models.User
		LocationText string `json:"location"`
	}

	err := models.DB.Raw(`
        SELECT users.*, ST_AsGeoJSON(users.location::geometry) AS location_text
        FROM users`).Scan(&result).Error

	if err != nil {
		log.Printf("Request error: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	users := make([]models.User, len(result))
	for i, r := range result {
		if r.User.LocationEWKB == nil {
			r.User.LocationText = common.GeoJSONText{Data: json.RawMessage("{}")}
		} else {
			r.User.LocationText = common.GeoJSONText{Data: json.RawMessage(r.LocationText)}
		}
		users[i] = r.User
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(users); err != nil {
		log.Printf("Request error: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var result struct {
		models.User
		LocationText string `json:"location"`
	}

	err := models.DB.Raw(`
        SELECT users.*, ST_AsGeoJSON(users.location::geometry) AS location_text
        FROM users WHERE users.id = ?`, id).Scan(&result).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.RespondWithError(w, http.StatusNotFound, "User is not found")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}

	if result.User.LocationEWKB == nil {
		result.User.LocationText = common.GeoJSONText{Data: json.RawMessage("{}")}
	} else {
		result.User.LocationText = common.GeoJSONText{Data: json.RawMessage(result.LocationText)}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result.User); err != nil {
		log.Printf("Serialization error: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Error encoding response")
	}
}

func orbToEWKB(geom orb.Geometry, srid int) ([]byte, error) {
	data, err := ewkb.Marshal(geom, srid, binary.LittleEndian)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// RegisterUser godoc
// @Summary Register a new user
// @Description Creates a new user with the provided information
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.UserInputS true "User data for registration"
// @Success 200 {object} map[string]interface{} "id, token" "ID and token of the newly registered user"
// @Failure 400 {object} string "Validation Error"
// @Failure 500 {object} string "Internal Server Error"
// @Router /users/registration [post]
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var userInput UserInput

	body, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(body, &userInput)

	validate = validator.New()

	err := validate.Struct(userInput)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Validation Error")
		return
	}

	var (
		locationEWKB []byte
		geom         orb.Geometry
	)

	if userInput.Location != nil {
		geom = userInput.Location.Geometry()
		locationEWKB, err = orbToEWKB(geom, 4326)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Location Validation Error")
		}
	}

	hashedPassword, err := HashPassword(userInput.Password)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Password Hashing Error")
	}

	user := &models.User{
		Name:         userInput.Name,
		Email:        userInput.Email,
		Pass_hash:    hashedPassword,
		LocationEWKB: locationEWKB,
		PhoneNumber:  userInput.PhoneNumber,
	}

	if err := models.DB.Create(user).Error; err != nil {
		utils.RespondWithError(w, http.StatusForbidden, "Failed to create new user")
		return
	}

	token, _ := GenerateToken()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": user.ID, "token": token})
}

// AuthenticateUser godoc
// @Summary Authenticate a user
// @Description Authenticates a user and returns a token
// @Tags users
// @Accept json
// @Produce json
// @Param credentials body models.AuthInputS true "User credentials for authentication"
// @Success 200 {object} map[string]interface{} "id, token" "ID and token of the authenticated user"
// @Failure 400 {object} string "Incorrect password or validation error"
// @Failure 404 {object} string "User not found"
// @Failure 500 {object} string "Internal Server Error"
// @Router /users/authentication [post]
func AuthenticateUser(w http.ResponseWriter, r *http.Request) {
	type AuthInput struct {
		Email    string `json:email`
		Password string `json:password`
	}

	var authInput AuthInput

	body, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(body, &authInput)

	validate = validator.New()

	err := validate.Struct(authInput)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Validation Error")
		return
	}

	var user models.User
	err = models.DB.Where("email = ?", authInput.Email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.RespondWithError(w, http.StatusNotFound, "User not found")
		} else {
			log.Printf("Request error: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}

	if !CheckPasswordHash(authInput.Password, user.Pass_hash) {
		utils.RespondWithError(w, http.StatusBadRequest, "Incorrect password")
		return
	}

	token, _ := GenerateToken()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": user.ID, "token": token})
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Deletes a user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 204 "User successfully deleted"
// @Failure 404 {object} string "User not found"
// @Failure 500 {object} string "Internal Server Error"
// @Router /users/{id} [delete]
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var exists struct {
		ID int `gorm:"column:id"`
	}
	err := models.DB.Raw(`SELECT id FROM users WHERE id = ?`, id).Scan(&exists).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.RespondWithError(w, http.StatusNotFound, "User not found")
		} else {
			log.Printf("Request error: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}

	err = models.DB.Exec(`DELETE FROM users WHERE id = ?`, id).Error
	if err != nil {
		log.Printf("Error deleting user: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Error deleting user")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GenerateToken() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	return token.SignedString([]byte(signingKey))
}

// UpdateUser godoc
// @Summary Update user details
// @Description Updates details of an existing user by ID.
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body models.UserUpdateSwagger true "User data to update"
// @Success 200 {object} models.UserResponse "Successfully updated user details"
// @Failure 400 {object} string "Validation Error"
// @Failure 404 {object} string "User not found"
// @Router /users/{id} [put]
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	var (
		locationEWKB []byte
		geom         orb.Geometry
		input        UserUpdate
		user         models.User
	)

	id := mux.Vars(r)["id"]

	if err := models.DB.Where("id = ?", id).First(&user).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "User not found")
		return
	}

	body, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(body, &input)

	validate := validator.New()
	err := validate.Struct(input)

	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Validation Error")
		return
	}

	if input.Location != nil {
		geom = input.Location.Geometry()
		locationEWKB, err = orbToEWKB(geom, 4326)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Location Validation Error")
		}
	}

	user.Name = input.Name
	user.LocationEWKB = locationEWKB
	user.PhoneNumber = input.PhoneNumber

	models.DB.Save(&user)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
