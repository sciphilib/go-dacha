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

	if geom != nil {
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

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	// id := mux.Vars(r)["id"]
	// var user models.User

	// if err := models.DB.Where("id = ?", id).First(&user).Error; err != nil {
	// 	utils.RespondWithError(w, http.StatusNotFound, "User not found")
	// 	return
	// }

	// var input UserInput

	// body, _ := io.ReadAll(r.Body)
	// _ = json.Unmarshal(body, &input)

	// validate := validator.New()
	// err := validate.Struct(input)

	// if err != nil {
	// 	utils.RespondWithError(w, http.StatusBadRequest, "Validation Error")
	// 	return
	// }

	// user.Name = input.Name
	// user.Email = input.Email
	// user.Password = input.Password

	// models.DB.Save(&user)

	// w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(user)
}
