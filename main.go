package main

import (
	"github.com/Paracetamol56/dev-website/api/db"
	"github.com/Paracetamol56/dev-website/api/server"
	_ "github.com/Paracetamol56/dev-website/docs"
	"github.com/joho/godotenv"
)

//	@title			Dev Website API
//	@version		1.0
//	@description	This is the API for my personal dev website
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	Matheo Galuba
//	@contact.url	https://dev.matheo-galuba.com/contact
//	@contact.email	matheo.galu56@gmail.com

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:8080
//	@BasePath	/api

//	@securityDefinitions.apikey	Bearer
//	@in							header
//	@name						Authorization
//	@description				Bearer token

func main() {
	godotenv.Load()
	db.Init()
	server.Run()
}
