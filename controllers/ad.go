package controllers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"github.com/sciphilib/go-dacha/common"
	"github.com/sciphilib/go-dacha/models"
	"github.com/sciphilib/go-dacha/utils"
	"gorm.io/gorm"
)

// GetAllAds godoc
// @Summary Get all ads
// @Description Retrieves a list of all advertisements with detailed information
// @Tags ads
// @Accept json
// @Produce json
// @Success 200 {array} models.AdResponse "An array of advertisement objects"
// @Failure 500 {object} nil "Internal Server Error"
// @Router /ads [get]
func GetAllAds(w http.ResponseWriter, r *http.Request) {
	var result []struct {
		models.Ad
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

	ads := make([]models.Ad, len(result))
	for i, r := range result {
		r.Ad.LocationText = common.GeoJSONText{Data: json.RawMessage(r.LocationText)}
		r.Ad.Subcategory.ID = r.SubcategoryID
		r.Ad.Subcategory.Name = r.SubcategoryName
		r.Ad.Subcategory.Category = r.CategoryName
		ads[i] = r.Ad
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
			"id":    ad.ID,
			"title": ad.Title,
			"price": ad.Price,
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

func GetAd(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var result struct {
		models.Ad
		LocationText string         `json:"location_text"`
		Pictures     pq.StringArray `gorm:"column:pictures" json:"pictures"`
	}

	err := models.DB.Raw(`
        SELECT advertisements.*, ST_AsGeoJSON(advertisements.location::geometry) AS location_text FROM advertisements WHERE id = ?`, id).Scan(&result).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.RespondWithError(w, http.StatusNotFound, "Ad not found")
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Internal Server Error")
		}
		return
	}

	ad := result.Ad
	ad.LocationText = common.GeoJSONText{Data: json.RawMessage(result.LocationText)}
	ad.PicturesText = []string(result.Pictures)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ad); err != nil {
		log.Printf("Serialization error: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Error encoding response")
	}
}

func CreateAd(w http.ResponseWriter, r *http.Request) {

}

func UpdateAd(w http.ResponseWriter, r *http.Request) {

}

func DeleteAd(w http.ResponseWriter, r *http.Request) {

}
