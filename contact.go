package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"github.com/tinfoil-knight/rest-api/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var validate *validator.Validate

// Contact : Struct for Storing Contacts
type Contact struct {
	ID    primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name  string             `json:"name,omitempty" bson:"name,omitempty" validate:"required,alpha,min=3,max=20"`
	Phone string             `json:"phone,omitempty" bson:"phone,omitempty" validate:"required,numeric,len=10"`
}

// Check if unique works

func apiHandler(w http.ResponseWriter, r *http.Request) {
	validate = validator.New()
	param := r.URL.Path[len("/api/"):]

	w.Header().Set("Content-Type", "application/json")
	collection := client.Database(config.Get("DB")).Collection(config.Get("COLLECTION"))
	switch r.Method {
	case "POST":
		var contact Contact
		// Read Request
		json.NewDecoder(r.Body).Decode(&contact)
		// Validate Request Data
		vErr := validate.Struct(contact)
		if vErr != nil {
			for _, err := range vErr.(validator.ValidationErrors) {
				newErr := fmt.Errorf("%v has validation error in %v", err.Namespace(), err.Type())
				sendErr(w, http.StatusBadRequest, newErr)
			}
			return
		}
		// Process Request in DB
		result, err := collection.InsertOne(context.TODO(), contact)
		if err != nil {
			sendErr(w, http.StatusInternalServerError, err)
			return
		}
		// Send Response
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(result)
	case "PUT":
		var contact Contact
		// Read Request
		json.NewDecoder(r.Body).Decode(&contact)
		// Validate Request Data
		vErr := validate.Var(contact.Phone, "required,alphanum,len=10")
		if vErr != nil {
			sendErr(w, http.StatusBadRequest, vErr)
			return
		}

		id, _ := primitive.ObjectIDFromHex(param)
		filter := bson.M{"_id": id}
		update := bson.D{primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "phone", Value: contact.Phone},
		}},
		}
		// Process Request in DB
		updateResult, err := collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			sendErr(w, http.StatusInternalServerError, err)
			return
		}
		// Send Response
		json.NewEncoder(w).Encode(updateResult)
	case "DELETE":
		id, _ := primitive.ObjectIDFromHex(param)
		filter := bson.M{"_id": id}
		// Process Request in DB
		deleteResult, err := collection.DeleteOne(context.TODO(), filter)
		if err != nil {
			sendErr(w, http.StatusInternalServerError, err)
			return
		}
		// Send Response
		json.NewEncoder(w).Encode(deleteResult)
	case "GET":
		if len(param) == 0 {
			filter := bson.D{{}}
			// Fetch all documents
			cur, err := collection.Find(context.TODO(), filter)
			if err != nil {
				sendErr(w, http.StatusInternalServerError, err)
				return
			}
			// Write the documents to a splice of struct Contact
			var contacts []*Contact
			defer cur.Close(context.TODO())
			for cur.Next(context.TODO()) {
				var contact Contact
				err := cur.Decode(&contact)
				checkErr(err)
				contacts = append(contacts, &contact)
			}
			if err := cur.Err(); err != nil {
				sendErr(w, http.StatusInternalServerError, err)
				return
			}
			// Send Response
			json.NewEncoder(w).Encode(contacts)
		} else {
			id, _ := primitive.ObjectIDFromHex(param)
			filter := bson.M{"_id": id}
			var contact Contact
			// Find document w/ matching id
			err := collection.FindOne(context.TODO(), filter).Decode(&contact)
			if err != nil {
				sendErr(w, http.StatusInternalServerError, err)
				return
			}
			// Send Response
			json.NewEncoder(w).Encode(contact)
		}
	default:
		err := errors.New("Illegal Method")
		sendErr(w, http.StatusMethodNotAllowed, err)
		return
	}

}

// HELPER FUNCTIONS
func sendErr(w http.ResponseWriter, StatusCode int, err error) {
	w.WriteHeader(StatusCode)
	log.Println(err.Error())
	w.Write([]byte(`{ "error": "` + err.Error() + `" }`))
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		handler.ServeHTTP(w, r)
		log.Printf("%s %s %s\n", r.Method, r.URL, time.Since(start).String())
	})
}

func getClient(uri string) *mongo.Client {
	clientOptions := options.Client().ApplyURI(uri)
	defer fmt.Println("Connected to MongoDB!")
	c, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	return c
}

func main() {
	fmt.Printf("Connecting to %v\n", config.Get("MONGODB_URI"))
	client = getClient(config.Get("MONGODB_URI"))
	httpPort := config.Get("PORT")
	portString := fmt.Sprintf(":%s", httpPort)
	http.HandleFunc("/api/", apiHandler)

	fmt.Printf("Server starting on http://localhost:%s\n", httpPort)

	log.Fatal(http.ListenAndServe(portString, logRequest(http.DefaultServeMux)))
}
