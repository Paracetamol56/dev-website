package main

import (
	"log"
	"main/backend"
	"main/backend/api"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

var validate *validator.Validate

func init() {
	godotenv.Load()
	validate = validator.New()
	if err := api.InitDatabase(); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	backend.Run()
}
