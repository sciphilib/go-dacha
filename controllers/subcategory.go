package controllers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/sciphilib/go-dacha/models"
	"github.com/sciphilib/go-dacha/utils"
	"gorm.io/gorm"
)

func GetAllSubcategories(w http.ResponseWriter, r *http.Request) {
	var subcategories []models.Subcategory
	err := models.DB.
		Preload("Category").
		Find(&subcategories).Error

	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Encoding error")
		return
	}

	var formattedSubcategories []map[string]interface{}
	for _, subcategory := range subcategories {
		formattedSubcategory := map[string]interface{}{
			"id":       subcategory.ID,
			"name":     subcategory.Name,
			"category": subcategory.Category.Name,
		}
		formattedSubcategories = append(formattedSubcategories, formattedSubcategory)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(formattedSubcategories); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Encoding error")
		return
	}
}

func GetSubcategory(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var subcategory models.Subcategory

	err := models.DB.
		Preload("Category").
		First(&subcategory, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.RespondWithError(w, http.StatusNotFound, "Subcategory not found")
			return
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Encoding error")
			return
		}
		return
	}

	response := map[string]interface{}{
		"id":       subcategory.ID,
		"name":     subcategory.Name,
		"category": subcategory.Category.Name,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Serialization error")
		return
	}
}

type SubcategoryInput struct {
	Category string `json:"category"`
	Name     string `json:"name"`
}

func CreateSubcategory(w http.ResponseWriter, r *http.Request) {
	var input SubcategoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if err := validator.New().Struct(input); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Validation Error")
		return
	}

	var category models.Category
	if err := models.DB.Where("name = ?", input.Category).First(&category).Error; err != nil {
		utils.RespondWithError(w, http.StatusForbidden, "Unknown category")
		return
	}

	subcategory := models.Subcategory{
		Name:       input.Name,
		CategoryID: category.ID,
	}

	if err := models.DB.Create(&subcategory).Error; err != nil {
		utils.RespondWithError(w, http.StatusForbidden, "Failed to create new subcategory")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(subcategory)
}

func UpdateSubcategory(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var subcategory models.Subcategory

	if err := models.DB.Where("id = ?", id).First(&subcategory).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Subcategory not found")
		return
	}

	var input SubcategoryInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if err := validator.New().Struct(input); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Validation Error")
		return
	}

	var category models.Category
	if err := models.DB.Where("name = ?", input.Category).First(&category).Error; err != nil {
		utils.RespondWithError(w, http.StatusForbidden, "Unknown category")
		return
	}

	subcategory.Name = input.Name
	subcategory.CategoryID = category.ID

	if err := models.DB.Save(&subcategory).Error; err != nil {
		utils.RespondWithError(w, http.StatusForbidden, "Failed to update subcategory")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(subcategory)
}

func DeleteSubcategory(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var Subcategory models.Subcategory

	if err := models.DB.Where("id = ?", id).First(&Subcategory).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Subcategory not found")
		return
	}

	models.DB.Delete(&Subcategory)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
