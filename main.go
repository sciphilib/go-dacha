package main

import (
	"github.com/joho/godotenv"
	"github.com/sciphilib/go-dacha/controllers"
	"github.com/sciphilib/go-dacha/models"
	"net/http"
)

func main() {
	godotenv.Load()

	handler := controllers.New()

	server := &http.Server{
		Addr:    "0.0.0.0:8008",
		Handler: handler,
	}

	models.ConnectDatabase()

	server.ListenAndServe()
}
