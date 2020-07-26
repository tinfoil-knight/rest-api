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
	"github.com/tinfoil-knight/rest-api/helpers"
	"github.com/tinfoil-knight/rest-api/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/mediocregopher/radix/v3"
)

var client *mongo.Client
var pool *radix.Pool

func apiHandler(w http.ResponseWriter, r *http.Request) {
	var validate *validator.Validate
	validate = validator.New()
	param := r.URL.Path[len("/api/"):]

	w.Header().Set("Content-Type", "application/json")
	collection := client.Database(config.Get("DB")).Collection(config.Get("COLLECTION"))
	switch r.Method {
	case "POST":
		var contact models.Contact
		// Read Request
		json.NewDecoder(r.Body).Decode(&contact)
		// Validate Request Data
		vErr := validate.Struct(contact)
		if vErr != nil {
			var errors []string
			for _, err := range vErr.(validator.ValidationErrors) {
				newErr := fmt.Sprintf("%v has validation error in %v", err.Namespace(), err.Type())
				errors = append(errors, newErr)
			}
			sendErr(w, http.StatusBadRequest, fmt.Errorf("%v", errors))
			return
		}
		// Process Request in DB
		result, err := collection.InsertOne(context.TODO(), contact)
		if err != nil {
			sendErr(w, http.StatusInternalServerError, err)
			return
		}
		// Delete inavlid keys
		if err := pool.Do(radix.Cmd(nil, "DEL", "ALL")); err != nil {
			log.Println(err)
		}
		// Send Response
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(result)
	case "PUT":
		var contact models.Contact
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
		// Delete invalid keys
		if err := pool.Do(radix.Cmd(nil, "DEL", "ALL", param)); err != nil {
			log.Println(err)
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
		// Delete invalid keys
		if deleteResult.DeletedCount > 0 {
			if err := pool.Do(radix.Cmd(nil, "DEL", "ALL", param)); err != nil {
				log.Println(err)
			}
		}
		// Send Response
		json.NewEncoder(w).Encode(deleteResult)
	case "GET":
		if len(param) == 0 {
			// Utilising Cache if Available
			var cachedValue string
			if err := pool.Do(radix.Cmd(&cachedValue, "GET", "ALL")); err != nil {
				log.Println(err)
			}
			if len(cachedValue) != 0 {
				// Send Cached Response
				var cachedContacts []*models.Contact

				bytes := []byte(cachedValue)
				if err := json.Unmarshal(bytes, &cachedContacts); err != nil {
					log.Println(err)
				}
				if err := json.NewEncoder(w).Encode(cachedContacts); err != nil {
					log.Println(err)
				} else {
					return
				}
			}
			// Requesting Response from MongoDB when cache is unavailable
			var contacts []*models.Contact
			filter := bson.D{{}}
			// Fetch all documents from DB
			cur, err := collection.Find(context.TODO(), filter)
			if err != nil {
				sendErr(w, http.StatusInternalServerError, err)
				return
			}
			// Write the documents to a splice of struct Contact
			defer cur.Close(context.TODO())
			for cur.Next(context.TODO()) {
				var contact models.Contact
				err := cur.Decode(&contact)
				checkErr(err)
				contacts = append(contacts, &contact)
			}
			if err := cur.Err(); err != nil {
				sendErr(w, http.StatusInternalServerError, err)
				return
			}
			// Store Response in Cache
			jsonArray, err := json.Marshal(contacts)
			if err != nil {
				log.Println(err)
			}
			if err := pool.Do(radix.Cmd(nil, "SET", "ALL", string(jsonArray))); err != nil {
				log.Printf("%v", err)
			}
			// Send Response
			json.NewEncoder(w).Encode(contacts)
		} else {
			// Utilising Cache if Available
			var cachedValue string
			if err := pool.Do(radix.Cmd(&cachedValue, "GET", param)); err != nil {
				log.Println(err)
			}
			if len(cachedValue) != 0 {
				// Send Cached Response
				var cachedContact *models.Contact

				bytes := []byte(cachedValue)
				if err := json.Unmarshal(bytes, &cachedContact); err != nil {
					log.Println(err)
				}
				if err := json.NewEncoder(w).Encode(cachedContact); err != nil {
					log.Println(err)
				} else {
					return
				}
			}
			// Requesting Response from MongoDB when cache is unavailable
			id, _ := primitive.ObjectIDFromHex(param)
			filter := bson.M{"_id": id}
			var contact models.Contact
			// Find document w/ matching id
			err := collection.FindOne(context.TODO(), filter).Decode(&contact)
			if err != nil {
				sendErr(w, http.StatusInternalServerError, err)
				return
			}
			// Store Response in Cache
			jsonStruct, err := json.Marshal(contact)
			if err != nil {
				log.Println(err)
			}
			if err := pool.Do(radix.Cmd(nil, "SET", param, string(jsonStruct))); err != nil {
				log.Println(err)
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
func sendErr(w http.ResponseWriter, statusCode int, err error) {
	w.WriteHeader(statusCode)
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
		log.Printf("%s %s %s\n", r.Method, r.URL, time.Since(start).Round(time.Microsecond).String())
	})
}

func main() {
	fmt.Printf("INFO: Connecting to %v\n", config.Get("MONGODB_URI"))
	client = helpers.GetDB(config.Get("MONGODB_URI"))
	pool = helpers.GetCache()
	httpPort := config.Get("PORT")
	portString := fmt.Sprintf(":%s", httpPort)
	http.HandleFunc("/api/", apiHandler)

	fmt.Printf("INFO: Server starting on http://localhost:%s/api/\n", httpPort)
	log.Fatal(http.ListenAndServe(portString, logRequest(http.DefaultServeMux)))
}

/**
TODO:
Add stack trace to error handling.
Increase test coverage to 80%
Refactor DB actions if possible.
Add timeout to server.
Use http.Error() everywhere.
Move error handling to helpers package.
Use a Makefile.
Log into text files.
Connect to a slack-bot/telegram-bot and send server error messages. (Separate this things entirely from this module)
Add monitoring???
Make all functions in packages isolated and resusable.
Reduce the number of gloabl variables used.
Make logs colourful.
USe contxts properly w/ timeout.
Create a dockerfile.
Make this API compliant w/ Swagger.
Create a separate branch w/ Postgres.
Make handlers responsible for only writing responses and
validating stuff and leave the rest over to resusable database methods.
Add a photo-field and integrate Amaxon Cloudfront or some other CDN(netlify,digital ocean etc.)
The cache will run out of memory if the database gets too big and all requests are cached.
See if Redis prevents this automatically or does this needs to be configured manually.
Expand PUT functionality to change names too.
Add a production build flag.
See MongoDB Starter doc to get a hint on what parts to move in db helper methods. Insert Methods on struct Contact.
**/
