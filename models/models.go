package models

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Contact : Struct for Storing Contacts
type Contact struct {
	ID    primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name  string             `json:"name,omitempty" bson:"name,omitempty" validate:"required,alpha,min=3,max=20"`
	Phone string             `json:"phone,omitempty" bson:"phone,omitempty" validate:"required,numeric,len=10"`
}

// GetClient : Returns a MongoDB client to be used for the DB.
func GetClient(uri string) *mongo.Client {
	clientOptions := options.Client().ApplyURI(uri)
	defer fmt.Println("Connected to MongoDB!")
	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	return c
}
