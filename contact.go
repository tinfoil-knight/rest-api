package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Contact : Struct for Storing Contacts
type Contact struct {
	ID    primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name  string             `json:"name" bson:"name,omitempty"`
	Phone string             `json:"phone,omitempty" bson:"phone,omitempty"`
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	collection := getClient()
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "POST":
		// TODO
	case "PUT":
		// TODO
	case "DELETE":
		// TODO
	default:
		filter := bson.D{{}}
		var results []Contact
		cur, err := collection.Find(context.TODO(), filter)
		fmt.Println(cur)
		checkErr(err)
		defer cur.Close(context.TODO())
		for cur.Next(context.TODO()) {
			var elem Contact
			err := cur.Decode(&elem)
			checkErr(err)
			fmt.Println(elem)
			results = append(results, elem)
		}

		if err := cur.Err(); err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(results)
	}

}

// HELPER FUNCTIONS

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

func getClient() *mongo.Collection {
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGODB_URI"))
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}
	collection := client.Database(os.Getenv("DBNAME")).Collection("contacts")
	return collection
}

func main() {
	setVariable()
	httpPort := os.Getenv("PORT")
	portString := fmt.Sprintf(":%s", httpPort)
	http.HandleFunc("/api", apiHandler)

	fmt.Printf("Server starting on http://localhost:%s\n", httpPort)

	log.Fatal(http.ListenAndServe(portString, logRequest(http.DefaultServeMux)))
}
