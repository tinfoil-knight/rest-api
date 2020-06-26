package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"./config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

// Contact : Struct for Storing Contacts
type Contact struct {
	// ID    primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name  string `json:"name" bson:"name,omitempty"`
	Phone string `json:"phone,omitempty" bson:"phone,omitempty"`
}

func apiHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	collection := client.Database(os.Getenv("DB")).Collection(os.Getenv("COLLECTION"))
	switch r.Method {
	case "POST":
		var contact *Contact
		json.NewDecoder(r.Body).Decode(&contact)
		result, err := collection.InsertOne(context.TODO(), contact)
		if err != nil {
			sendErr(w, err, "Couldn't create a new contact. Please try again.")
			return
		}
		json.NewEncoder(w).Encode(result)
	case "PUT":
		var contact *Contact
		json.NewDecoder(r.Body).Decode(&contact)
		filter := bson.D{primitive.E{Key: "name", Value: contact.Name}}
		update := bson.D{primitive.E{Key: "$set", Value: bson.D{
			primitive.E{Key: "phone", Value: contact.Phone},
		}},
		}
		updateResult, err := collection.UpdateOne(context.TODO(), filter, update)
		if err != nil {
			sendErr(w, err, "Changing the record failed. Try again.")
			return
		}
		json.NewEncoder(w).Encode(updateResult)

	case "DELETE":
		// TODO
	case "GET":
		filter := bson.D{{}}
		var contacts []*Contact
		cur, err := collection.Find(context.TODO(), filter)
		if err != nil {
			sendErr(w, err, "Couldn't fetch the records. Try again.")
			return
		}
		defer cur.Close(context.TODO())
		for cur.Next(context.TODO()) {
			var elem Contact
			err := cur.Decode(&elem)
			checkErr(err)
			contacts = append(contacts, &elem)
		}

		if err := cur.Err(); err != nil {
			sendErr(w, err, "Couldn't fetch the records. Try again.")
			return
		}

		json.NewEncoder(w).Encode(contacts)
	default:
		fmt.Println("Illegal Method")

	}

}

// HELPER FUNCTIONS
func sendErr(w http.ResponseWriter, err error, message string) {
	w.WriteHeader(http.StatusInternalServerError)
	log.Println(err.Error())
	w.Write([]byte(`{ "error": "` + message + `" }`))
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s\n", r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func getClient() *mongo.Client {
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGODB_URI"))
	defer fmt.Println("Connected to MongoDB!")
	c, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	return c
}

func main() {
	config.SetVariable()
	fmt.Printf("Connecting to %v ...\n", os.Getenv("MONGODB_URI"))
	client = getClient()

	httpPort := os.Getenv("PORT")
	portString := fmt.Sprintf(":%s", httpPort)
	http.HandleFunc("/api", apiHandler)

	fmt.Printf("Server starting on http://localhost:%s\n", httpPort)

	log.Fatal(http.ListenAndServe(portString, logRequest(http.DefaultServeMux)))
}
