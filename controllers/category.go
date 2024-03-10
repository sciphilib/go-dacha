package controllers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/sciphilib/go-dacha/models"
	"github.com/sciphilib/go-dacha/utils"
)

type CategoryInput struct {
	Name string `json:"name" validate:"required"`
}

// GetAllCategories godoc
// @Summary Get all categories
// @Description Retrieves a list of all categories
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {array} models.Category "List of categories"
// @Router /categories [get]
func GetAllCategories(w http.ResponseWriter, r *http.Request) {
	var categories []models.Category
	models.DB.Find(&categories)

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(categories); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Encoding error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
}

// GetCategory godoc
// @Summary Get a category by ID
// @Description Retrieves a category by its ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} models.Category "Category found"
// @Failure 404 {object} string "Category not found"
// @Router /categories/{id} [get]
func GetCategory(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var category models.Category

	if err := models.DB.Where("id = ?", id).First(&category).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Category not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

// CreateCategory godoc
// @Summary Create a new category
// @Description Creates a new category with the provided name
// @Tags categories
// @Accept json
// @Produce json
// @Param category body CategoryInput true "Category data"
// @Success 200 {object} models.Category "Category created"
// @Failure 400 {object} string "Validation Error"
// @Router /categories [post]
func CreateCategory(w http.ResponseWriter, r *http.Request) {
	var input CategoryInput

	body, _ := ioutil.ReadAll(r.Body)
	_ = json.Unmarshal(body, &input)

	validate = validator.New()
	err := validate.Struct(input)

	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Validation Error")
		return
	}

	category := &models.Category{
		Name: input.Name,
	}

	models.DB.Create(category)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

// UpdateCategory godoc
// @Summary Update a category
// @Description Updates the name of an existing category by ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param category body CategoryInput true "Updated category data"
// @Success 200 {object} models.Category "Category updated"
// @Failure 400 {object} string "Validation Error"
// @Failure 404 {object} string "Category not found"
// @Router /categories/{id} [put]
func UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var category models.Category

	if err := models.DB.Where("id = ?", id).First(&category).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Category not found")
		return
	}

	var input CategoryInput

	body, _ := ioutil.ReadAll(r.Body)
	_ = json.Unmarshal(body, &input)

	validate = validator.New()
	err := validate.Struct(input)

	if err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Validation Error")
		return
	}

	category.Name = input.Name

	if err := models.DB.Save(&category).Error; err != nil {
		utils.RespondWithError(w, http.StatusForbidden, "Failed to update category")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(category)
}

// DeleteCategory godoc
// @Summary Delete a category
// @Description Deletes an existing category by ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 "Category successfully deleted"
// @Failure 404 {object} string "Category not found"
// @Router /categories/{id} [delete]
func DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var category models.Category

	if err := models.DB.Where("id = ?", id).First(&category).Error; err != nil {
		utils.RespondWithError(w, http.StatusNotFound, "Category not found")
		return
	}

	models.DB.Delete(&category)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
