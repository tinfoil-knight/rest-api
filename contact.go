package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"go.mongodb.org/mongo-driver/bson"
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
		// TODO
	case "PUT":
		// TODO
	case "DELETE":
		// TODO
	case "GET":
		filter := bson.D{{}}
		var results []*Contact
		cur, err := collection.Find(context.TODO(), filter)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}
		defer cur.Close(context.TODO())
		for cur.Next(context.TODO()) {
			var elem Contact
			err := cur.Decode(&elem)
			checkErr(err)
			results = append(results, &elem)
		}

		if err := cur.Err(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}

		json.NewEncoder(w).Encode(results)

	default:
		fmt.Println("Illegal Method")

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
	setVariable()
	client = getClient()

	httpPort := os.Getenv("PORT")
	portString := fmt.Sprintf(":%s", httpPort)
	http.HandleFunc("/api", apiHandler)

	fmt.Printf("Server starting on http://localhost:%s\n", httpPort)

	log.Fatal(http.ListenAndServe(portString, logRequest(http.DefaultServeMux)))
}
