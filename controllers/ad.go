package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/sciphilib/go-dacha/common"
	"github.com/sciphilib/go-dacha/models"
	"github.com/sciphilib/go-dacha/utils"
	"gorm.io/gorm"
)

// GetAllAds godoc
// @Summary Get all ads
// @Description Retrieves a list of all advertisements with detailed information
// @Tags advertisements
// @Accept json
// @Produce json
// @Success 200 {array} models.AdResponse "An array of advertisement objects"
// @Failure 500 {object} nil "Internal Server Error"
// @Router /ads [get]
func GetAllAds(w http.ResponseWriter, r *http.Request) {
	var result []struct {
		models.Advertisement
		SubcategoryID   uint           `json:"subcategory_id"`
		SubcategoryName string         `json:"subcategory_name"`
		CategoryName    string         `json:"category_name"`
		LocationText    string         `json:"location"`
		Pictures        pq.StringArray `gorm:"column:pictures" json:"pictures"`
	}

	err := models.DB.Raw(`
	       SELECT
	             advertisements.*,
	             subcategories.id AS subcategory_id,
	             subcategories.name AS subcategory_name,
	             categories.name AS category_name,
	             ST_AsGeoJSON(advertisements.location::geometry) AS location_text
	       FROM advertisements
	       JOIN subcategories ON subcategories.id = advertisements.subcategory_id
	       JOIN categories ON categories.id = subcategories.category_id
	    `).
		Scan(&result).Error

	if err != nil {
		log.Printf("Request error: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	type user_result struct {
		models.User
		LocationText string `json:"location"`
	}
	var json_users []user_result

	err = models.DB.Raw(`
	    SELECT *, ST_AsGeoJSON(users.location::geometry) AS location_text
	    FROM users`).
		Scan(&json_users).Error

	if err != nil {
		log.Printf("Request error: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	users := make([]models.User, len(json_users))
	for i, r := range json_users {
		if r.User.LocationEWKB == nil {
			r.User.LocationText = common.GeoJSONText{Data: json.RawMessage("{}")}
		} else {
			r.User.LocationText = common.GeoJSONText{Data: json.RawMessage(r.LocationText)}
		}
		users[i] = r.User
	}

	userMap := make(map[uint]models.User)
	for _, user := range users {
		userMap[user.ID] = user
	}

	ads := make([]models.Advertisement, len(result))
	for i, r := range result {
		r.Advertisement.LocationText = common.GeoJSONText{Data: json.RawMessage(r.LocationText)}
		r.Advertisement.Subcategory.ID = r.SubcategoryID
		r.Advertisement.Subcategory.Name = r.SubcategoryName
		r.Advertisement.Subcategory.Category = r.CategoryName
		ads[i] = r.Advertisement
		ads[i].PicturesText = make([]string, len(r.Pictures))
		copy(ads[i].PicturesText, r.Pictures)
	}

	var formattedAds []map[string]interface{}

	for _, ad := range ads {
		user, exists := userMap[ad.User_id]
		if !exists {
			log.Printf("User with ID %d not found", ad.User_id)
			continue
		}

		formattedAd := map[string]interface{}{
			"id":          ad.ID,
			"title":       ad.Title,
			"price":       ad.Price,
			"description": ad.Description,
			"subcategory": map[string]interface{}{
				"name":     ad.Subcategory.Name,
				"category": ad.Subcategory.Category,
			},
			"user":     user,
			"datetime": ad.Datetime,
			"pictures": ad.PicturesText,
			"location": ad.LocationText,
		}
		formattedAds = append(formattedAds, formattedAd)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(formattedAds); err != nil {
		log.Printf("Serialization error: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
}

// GetAd godoc
// @Summary Get an ad by id
// @Description Retrieve an advertisements by id with detailed information
// @Tags advertisements
// @Accept json
// @Produce json
// @Param id path int true "Ad ID"
// @Success 200 {object} models.AdResponse "An advertisement object"
// @Failure 404 {object} nil "Ad not found"
// @Failure 500 {object} nil "Internal Server Error"
// @Router /ads/{id} [get]
func GetAd(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var result struct {
		models.Advertisement
		SubcategoryID   uint           `json:"subcategory_id"`
		SubcategoryName string         `json:"subcategory_name"`
		CategoryName    string         `json:"category_name"`
		LocationText    string         `json:"location"`
		Pictures        pq.StringArray `gorm:"column:pictures" json:"pictures"`
	}

	err := models.DB.Raw(`
	       SELECT
	             advertisements.*,
	             subcategories.id AS subcategory_id,
	             subcategories.name AS subcategory_name,
	             categories.name AS category_name,
	             ST_AsGeoJSON(advertisements.location::geometry) AS location_text
	       FROM advertisements
	       JOIN subcategories ON subcategories.id = advertisements.subcategory_id
	       JOIN categories ON categories.id = subcategories.category_id
	       WHERE advertisements.id = ?
	    `, id).
		Scan(&result).Error

	if result.Advertisement.ID == 0 {
		utils.RespondWithError(w, http.StatusNotFound, "Ad is not found")
		return
	}

	result.Advertisement.LocationText = common.GeoJSONText{Data: json.RawMessage(result.LocationText)}
	result.Advertisement.Subcategory.ID = result.SubcategoryID
	result.Advertisement.Subcategory.Name = result.SubcategoryName
	result.Advertisement.Subcategory.Category = result.CategoryName
	ad := result.Advertisement
	ad.PicturesText = make([]string, len(result.Pictures))
	copy(ad.PicturesText, result.Pictures)

	var user_json struct {
		models.User
		LocationText string `json:"location"`
	}

	err = models.DB.Raw(`
	    SELECT users.*, ST_AsGeoJSON(users.location::geometry) AS location_text
	    FROM users WHERE users.id = ?`, ad.User_id).
		Scan(&user_json).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.RespondWithError(w, http.StatusNotFound, "Ad is not found")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}

	if err != nil {
		log.Printf("Request error: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	if user_json.User.LocationEWKB == nil {
		user_json.User.LocationText = common.GeoJSONText{Data: json.RawMessage("{}")}
	} else {
		user_json.User.LocationText = common.GeoJSONText{Data: json.RawMessage(user_json.LocationText)}
	}
	user := user_json.User

	formattedAd := map[string]interface{}{
		"id":          ad.ID,
		"title":       ad.Title,
		"price":       ad.Price,
		"description": ad.Description,
		"subcategory": map[string]interface{}{
			"name":     ad.Subcategory.Name,
			"category": ad.Subcategory.Category,
		},
		"user":     user,
		"datetime": ad.Datetime,
		"pictures": ad.PicturesText,
		"location": ad.LocationText,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(formattedAd); err != nil {
		log.Printf("Serialization error: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
}

type UserAdInput struct {
	Title       string            `json:"title" validate:"required"`
	Price       string            `json:"price" validate:"required"`
	Subcategory string            `json:"subcategory" validate:"required"`
	Category    string            `json:"category" validate:"required"`
	Description string            `json:"description"`
	UserEmail   string            `json:"user_email" validate:"required"`
	Datetime    time.Time         `json:"datetime" validate:"required"`
	Pictures    []string          `json:"pictures"`
	Location    *geojson.Geometry `json:"location" validate:"required"`
}

// CreateAd godoc
// @Summary Add a new advertisement
// @Description Adds a new advertisement with the given details
// @Tags advertisements
// @Accept json
// @Produce json
// @Param ad body models.AdInput true "Create Ad"
// @Success 200 {object} models.AdAdded "ID of the newly created ad"
// @Failure 400 {string} string "Validation Error"
// @Failure 404 {string} string "Subcategory/User is not found"
// @Failure 403 {string} string "Failed to create a new ad"
// @Failure 500 {string} string "Internal Server Error"
// @Router /ads [post]
func CreateAd(w http.ResponseWriter, r *http.Request) {
	var (
		locationEWKB []byte
		geom         orb.Geometry
		userInput    UserAdInput
		subcategory  models.Subcategory
		user         models.User
	)

	body, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(body, &userInput)

	validate = validator.New()

	err := validate.Struct(userInput)
	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Validation Error")
		return
	}

	if userInput.Location != nil {
		geom = userInput.Location.Geometry()
		locationEWKB, err = orbToEWKB(geom, 4326)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Location Validation Error")
		}
	}

	err = models.DB.
		Preload("Category").
		Where("name = ?", userInput.Subcategory).
		First(&subcategory).Error
	if subcategory.ID == 0 {
		utils.RespondWithError(w, http.StatusNotFound, "Subcategory is not found")
		return
	}

	err = models.DB.
		Where("email = ?", userInput.UserEmail).
		First(&user).Error

	if user.ID == 0 {
		utils.RespondWithError(w, http.StatusNotFound, "User is not found")
		return
	}

	ad := &models.Advertisement{
		Title:          userInput.Title,
		Price:          userInput.Price,
		Subcategory_id: subcategory.ID,
		Description:    userInput.Description,
		User_id:        user.ID,
		Datetime:       userInput.Datetime,
		Pictures:       pq.StringArray(userInput.Pictures),
		LocationEWKB:   locationEWKB,
	}

	if err := models.DB.Create(ad).Error; err != nil {
		utils.RespondWithError(w, http.StatusForbidden, "Failed to create a new ad")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": ad.ID})
}

// UpdateAd godoc
// @Summary Update an advertisement
// @Description Update an existing advertisement by its ID with new information
// @Tags advertisements
// @Accept json
// @Produce json
// @Param id path int true "Ad ID"
// @Param ad body models.AdInput true "Advertisement data"
// @Success 200 {object} models.AdResponse "Successfully updated advertisement"
// @Failure 400 {object} string "Validation Error"
// @Failure 403 {object} string "Failed to update the ad"
// @Failure 404 {object} string "Ad/Subcategory/User not found"
// @Router /ads/{id} [put]
func UpdateAd(w http.ResponseWriter, r *http.Request) {
	var (
		locationEWKB []byte
		geom         orb.Geometry
		ad           models.Advertisement
		userInput    UserAdInput
		user         models.User
	)

	id := mux.Vars(r)["id"]
	if err := models.DB.Where("id = ?", id).First(&ad).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Ad not found")
		return
	}

	body, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(body, &userInput)

	validate := validator.New()
	err := validate.Struct(userInput)

	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Validation Error")
		return
	}

	if userInput.Location != nil {
		geom = userInput.Location.Geometry()
		locationEWKB, err = orbToEWKB(geom, 4326)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Location Validation Error")
		}
	}

	var subcategory models.Subcategory
	err = models.DB.
		Preload("Category").
		Where("name = ?", userInput.Subcategory).
		First(&subcategory).Error
	if subcategory.ID == 0 {
		utils.RespondWithError(w, http.StatusNotFound, "Subcategory is not found")
		return
	}

	err = models.DB.
		Where("email = ?", userInput.UserEmail).
		First(&user).Error

	if user.ID == 0 {
		utils.RespondWithError(w, http.StatusNotFound, "User is not found")
		return
	}

	ad.Title = userInput.Title
	ad.Price = userInput.Price
	ad.Subcategory_id = subcategory.ID
	ad.Description = userInput.Description
	ad.User_id = user.ID
	ad.Datetime = userInput.Datetime
	ad.Pictures = userInput.Pictures
	ad.LocationEWKB = locationEWKB

	if err := models.DB.Save(&ad).Error; err != nil {
		utils.RespondWithError(w, http.StatusForbidden, "Failed to update the ad")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ad)
}

// DeleteAd godoc
// @Summary Delete an advertisement
// @Description Deletes an advertisement by its ID
// @Tags advertisements
// @Accept json
// @Produce json
// @Param id path int true "Ad ID"
// @Success 200 {string} string "Ad deleted successfully"
// @Failure 404 {object} string "Ad not found"
// @Router /ads/{id} [delete]
func DeleteAd(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var ad models.Advertisement

	if err := models.DB.Where("id = ?", id).First(&ad).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Ad not found")
		return
	}

	models.DB.Delete(&ad)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
