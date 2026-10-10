package main

import (
	"context"
	"dev/internal/db"
	"dev/internal/models"
	"dev/internal/server"
	"dev/internal/utils"
	"log"

	_ "dev/docs"

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
	if migrated, err := models.MigrateUserIdentities(context.Background()); err != nil {
		log.Fatalln("Failed to migrate user identities:", err)
	} else if migrated > 0 {
		log.Printf("Migrated %d users to identities", migrated)
	}
	utils.ScheduleIconRefresh()
	server.Run()
}
