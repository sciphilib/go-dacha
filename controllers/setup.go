package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
)

func New() http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/users", GetAllUsers).Methods("GET")
	router.HandleFunc("/users/{id}", GetUser).Methods("GET")
	router.HandleFunc("/users/registration", RegisterUser).Methods("POST")
	router.HandleFunc("/users/authentication", AuthenticateUser).Methods("POST")

	return router
}
