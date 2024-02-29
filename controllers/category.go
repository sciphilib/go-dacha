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
