package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Contact : Struct for Storing Contacts
type Contact struct {
	ID    primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name  string             `json:"name,omitempty" bson:"name,omitempty" validate:"required,alpha,min=3,max=20"`
	Phone string             `json:"phone,omitempty" bson:"phone,omitempty" validate:"required,numeric,len=10"`
}

// Define methods on the struct in form of Insert, Get, Delete
