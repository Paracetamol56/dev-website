package main

import (
	"github.com/Paracetamol56/dev-website/api/db"
	"github.com/Paracetamol56/dev-website/api/server"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	db.Init()
	server.Run()
}
