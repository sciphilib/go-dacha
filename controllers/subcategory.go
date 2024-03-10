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

// GetAllSubcategories godoc
// @Summary Get all subcategories
// @Description Retrieves a list of all subcategories with their categories
// @Tags subcategories
// @Accept json
// @Produce json
// @Success 200 {array} models.SubcategoryResponse "List of subcategories"
// @Router /subcategories [get]
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

// GetSubcategory godoc
// @Summary Get a subcategory by ID
// @Description Retrieves a subcategory by its ID including category name
// @Tags subcategories
// @Accept json
// @Produce json
// @Param id path int true "Subcategory ID"
// @Success 200 {object} models.SubcategoryResponse "Subcategory found"
// @Failure 404 {object} string "Subcategory not found"
// @Router /subcategories/{id} [get]
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

// CreateSubcategory godoc
// @Summary Create a new subcategory
// @Description Creates a new subcategory within a category
// @Tags subcategories
// @Accept json
// @Produce json
// @Param subcategory body SubcategoryInput true "Subcategory creation data"
// @Success 200 {object} models.Subcategory "Subcategory created"
// @Failure 400 {object} string "Invalid JSON payload or validation error"
// @Failure 403 {object} string "Unknown category"
// @Router /subcategories [post]
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

// UpdateSubcategory godoc
// @Summary Update a subcategory
// @Description Updates an existing subcategory by ID
// @Tags subcategories
// @Accept json
// @Produce json
// @Param id path int true "Subcategory ID"
// @Param subcategory body SubcategoryInput true "Subcategory update data"
// @Success 200 {object} models.Subcategory "Subcategory updated"
// @Failure 400 {object} string "Invalid JSON payload or validation error"
// @Failure 403 {object} string "Unknown category"
// @Failure 404 {object} string "Subcategory not found"
// @Router /subcategories/{id} [put]
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

// DeleteSubcategory godoc
// @Summary Delete a subcategory
// @Description Deletes an existing subcategory by ID
// @Tags subcategories
// @Accept json
// @Produce json
// @Param id path int true "Subcategory ID"
// @Success 200 "Subcategory successfully deleted"
// @Failure 404 {object} string "Subcategory not found"
// @Router /subcategories/{id} [delete]
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
