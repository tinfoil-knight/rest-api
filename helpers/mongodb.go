package helpers

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// InitDB connects to the MongoDB instance and returns a DB client.
func InitDB(uri string) *mongo.Client {
	clientOptions := options.Client().ApplyURI(uri)
	defer fmt.Println("INFO: Connected to MongoDB")
	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("INFO: PINGing MongoDB")
	err = c.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	return c
}
