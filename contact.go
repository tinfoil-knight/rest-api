package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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
			http.Error(w, fmt.Sprintf("%v", errors), http.StatusBadRequest)
			return
		}
		// Process Request in DB
		result, err := collection.InsertOne(context.TODO(), contact)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
			http.Error(w, vErr.Error(), http.StatusBadRequest)
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			// Write the documents to a splice of struct Contact
			defer cur.Close(context.TODO())
			for cur.Next(context.TODO()) {
				var contact models.Contact
				err := cur.Decode(&contact)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				contacts = append(contacts, &contact)
			}
			if err := cur.Err(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
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
				http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusMethodNotAllowed)
		return
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
	mode := config.InitFlags()
	fmt.Printf("INFO: Starting application in mode: %s\n", *mode)
	fmt.Printf("INFO: Connecting to %v\n", config.Get("MONGODB_URI"))
	client = helpers.InitDB(config.Get("MONGODB-URI"))
	pool = helpers.InitCache()
	httpPort := config.Get("PORT")
	portString := fmt.Sprintf(":%s", httpPort)
	http.HandleFunc("/api/", apiHandler)

	fmt.Printf("INFO: Server starting on http://localhost:%s/api/\n", httpPort)
	if *mode == "PROD" {
		reqBody, _ := json.Marshal(map[string]string{"text": "Heya"})
		res, err := http.Post(config.Get("SLACK-HOOK"), "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			log.Println(err)
		}
		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		resString := string(body)
		if resString != "ok" {
			fmt.Printf("ERROR: Message not delivered to Slack Workspace")
		} else {
			fmt.Printf("INFO: Message has been delivered to Slack Workspace")
		}
	}

	log.Fatal(http.ListenAndServe(portString, logRequest(http.DefaultServeMux)))
}

/**
TODO:
Logging
	Add stack trace to error handling.
	Make logs colourful.
	Log into text files.
	Add monitoring through an external service.
	Connect to a slack-bot/telegram-bot and send server error messages.
Testing, Builds & Docs
	Make this API compliant w/ Swagger.
	Increase test coverage to 80%
	Use a Makefile.
	Create a dockerfile.
	Check if Redis is a LRU Cache or not.
	Fuzzy Testing?
	Test for Errors using Fault Injection, eg: https://github.com/github/go-fault
Refactoring
	Refactor DB code entirely s.t. DB can be changed without editing this file.
	Make all functions in packages pure.
	Remove all global variables.
Core
	Expand PUT functionality to change names too.
	Add a photo-field and integrate Amazon Cloudfront or some other CDN(netlify,digital ocean etc.)
	Use contexts properly w/ timeout.
	Add timeout to server.
**/
