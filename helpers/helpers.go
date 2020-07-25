package helpers

import (
	"context"
	"fmt"
	"log"

	"github.com/mediocregopher/radix/v3"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetCache connects to the Redis instance
func GetCache() *radix.Pool {
	c, err := radix.NewPool("tcp", "127.0.0.1:6379", 10)
	if err != nil {
		log.Printf("%v", err)
	}
	// PING the server and recover from  panic
	// TODO: enable auth
	return c
}

// GetDB connects to the MongoDB instance and returns a DB client.
func GetDB(uri string) *mongo.Client {
	clientOptions := options.Client().ApplyURI(uri)
	defer fmt.Println("INFO: Connected to MongoDB!")
	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	return c
}
