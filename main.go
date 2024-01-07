package main

import (
	"github.com/Paracetamol56/dev-website/api/db"
	"github.com/Paracetamol56/dev-website/api/server"
	"github.com/joho/godotenv"
	"github.com/swaggo/swag/example/basic/docs"
)

func main() {
	docs.SwaggerInfo.Title = "Dev Website API"
	docs.SwaggerInfo.Description = "API for the dev website"

	godotenv.Load()
	db.Init()
	server.Run()
}
