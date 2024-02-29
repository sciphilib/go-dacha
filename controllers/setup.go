package controllers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func New() http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/users", GetAllUsers).Methods("GET")
	router.HandleFunc("/users/{id}", GetUser).Methods("GET")
	router.HandleFunc("/users/registration", RegisterUser).Methods("POST")
	router.HandleFunc("/users/authentication", AuthenticateUser).Methods("POST")

	router.HandleFunc("/categories", GetAllCategories).Methods("GET")
	router.HandleFunc("/categories/{id}", GetCategory).Methods("GET")
	router.HandleFunc("/categories", CreateCategory).Methods("POST")
	router.HandleFunc("/categories/{id}", UpdateCategory).Methods("PUT")
	router.HandleFunc("/categories/{id}", DeleteCategory).Methods("DELETE")

	router.HandleFunc("/subcategories", GetAllSubcategories).Methods("GET")
	router.HandleFunc("/subcategories/{id}", GetSubcategory).Methods("GET")
	router.HandleFunc("/subcategories", CreateSubcategory).Methods("POST")
	router.HandleFunc("/subcategories/{id}", UpdateSubcategory).Methods("PUT")
	router.HandleFunc("/subcategories/{id}", DeleteSubcategory).Methods("DELETE")

	loggedRouter := Logger(router)

	return loggedRouter
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log.Printf("Started %s %s", r.Method, r.URL.Path)
		log.Println("Headers:")
		for name, values := range r.Header {
			valueString := fmt.Sprintf("%s: %s", name, values[0])
			log.Println(valueString)
		}

		next.ServeHTTP(w, r)

		log.Printf("Completed %s in %v", r.URL.Path, time.Since(start))
	})
}
