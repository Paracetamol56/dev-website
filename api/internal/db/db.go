package db

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const defaultName = "dev"

var db *mongo.Database

func Init() error {
	clientOptions := options.Client().ApplyURI(os.Getenv("DB_URI"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalln(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatalln(err)
	}

	name := os.Getenv("DB_NAME")
	if name == "" {
		name = defaultName
	}
	db = client.Database(name)
	return err
}

func GetDB() *mongo.Database {
	return db
}

// Use swaps the database, which tests need to run against a disposable one
func Use(database *mongo.Database) {
	db = database
}
