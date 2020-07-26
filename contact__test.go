package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tinfoil-knight/rest-api/config"
	"github.com/tinfoil-knight/rest-api/helpers"
	"github.com/tinfoil-knight/rest-api/models"
	"go.mongodb.org/mongo-driver/bson"
)

var results []*models.Contact
var dbInitialized = false

type serverResponse struct {
	MatchedCount  int8
	ModifiedCount int8
	UpsertedCount int8
	DeletedCount  int8
	UpsertedID    string
}

func runServer(fn func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(fn))
}

func initDB() {
	client = helpers.InitDB(config.Get("MONGODB_URI"))
	err := client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	collection := client.Database(config.Get("TESTDB")).Collection(config.Get("COLLECTION"))

	// mongo db generates new IDS everytime its adds new entries
	// so results[0] contains the ID from the first run initDB() function call

	if !dbInitialized {
		collection.DeleteMany(context.TODO(), bson.M{})
		contact1 := models.Contact{Name: "Jay Randall", Phone: "9087453245"}
		contact2 := models.Contact{Name: "Reinne Parsley", Phone: "8904576732"}
		contacts := []interface{}{contact1, contact2}
		collection.InsertMany(context.TODO(), contacts)
		dbInitialized = true
	}
	cur, _ := collection.Find(context.TODO(), bson.D{{}})
	for cur.Next(context.TODO()) {
		var elem models.Contact
		cur.Decode(&elem)
		results = append(results, &elem)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	cur.Close(context.TODO())
	// Cache Client is initiated here temporarily
	pool = helpers.InitCache()
}

func Test__GetAll(t *testing.T) {
	initDB()
	ts := runServer(apiHandler)
	url := ts.URL + "/api/"
	res, err := http.Get(url)
	if err != nil {
		t.Errorf("%s", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("HTTPStatusCode | Expected: %v, Received: %v", http.StatusOK, res.StatusCode)
	}
	var contacts *[]models.Contact
	json.NewDecoder(res.Body).Decode(&contacts)
	res.Body.Close()
	ts.Close()
}

func Test__GetOneByID(t *testing.T) {
	initDB()
	ts := runServer(apiHandler)
	contact := results[0]
	id := (contact.ID).Hex()
	url := ts.URL + "/api/" + id
	res, err := http.Get(url)
	if err != nil {
		t.Errorf("%s", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("HTTPStatusCode | Expected: %v, Received: %v", http.StatusOK, res.StatusCode)
	}
	var resContact models.Contact
	json.NewDecoder(res.Body).Decode(&resContact)
	if contact.Name != resContact.Name {
		t.Errorf("Field: Name of Contact | Expected: %s, Received: %s", contact.Name, resContact.Name)
	}
	if contact.Phone != resContact.Phone {
		t.Errorf("Field: Phone of Contact | Expected: %s, Received: %s", contact.Phone, resContact.Phone)
	}
	res.Body.Close()
	ts.Close()

}

func Test__PostOne(t *testing.T) {
	// Test Config
	initDB()
	ts := runServer(apiHandler)
	url := ts.URL + "/api/"
	// Test Run
	reqBody, err := json.Marshal(map[string]string{"name": "Ryder", "phone": "9022457831"})
	if err != nil {
		fmt.Println(err.Error())
	}
	res, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Errorf("%s\n", err)
	}
	if res.StatusCode != http.StatusCreated {
		t.Errorf("HTTPStatusCode | Expected: %v, Received: %v", http.StatusCreated, res.StatusCode)
	}
	res.Body.Close()
	ts.Close()
}
func Test__ChangeOneByID(t *testing.T) {
	initDB()
	ts := runServer(apiHandler)
	contact := results[0]
	id := (contact.ID).Hex()
	url := ts.URL + "/api/" + id
	// Test Run
	client := &http.Client{}
	reqBody, err := json.Marshal(map[string]string{"phone": "9145636789"})
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(reqBody))
	if err != nil {
		t.Errorf("%s", err)
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		t.Errorf("%s", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("HTTPStatusCode | Expected: %v, Received: %v", http.StatusOK, res.StatusCode)
	}
	var putResult serverResponse
	json.NewDecoder(res.Body).Decode(&putResult)
	if putResult.ModifiedCount != 1 {
		t.Errorf("ModifiedCount | Expected: %v, Received: %v", 1, putResult.ModifiedCount)
	}
	ts.Close()

}

func Test__DeleteOneByID(t *testing.T) {
	initDB()
	ts := runServer(apiHandler)
	contact := results[0]
	id := (contact.ID).Hex()
	url := ts.URL + "/api/" + id
	// Test Run
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		t.Errorf("%s", err)
	}
	res, err := client.Do(req)
	if err != nil {
		t.Errorf("%s", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("HTTPStatusCode | Expected: %v, Received: %v", http.StatusOK, res.StatusCode)
	}
	var delResult serverResponse
	json.NewDecoder(res.Body).Decode(&delResult)
	if delResult.DeletedCount != 1 {
		t.Errorf("ModifiedCount | Expected: %v, Received: %v", 1, delResult.DeletedCount)
	}
	ts.Close()

}

/**
Test for the following things:
- Is the Cache Working?
- Does the server start if the Cache doesn't work?
- Does the server reject and unresolved request properly?
- Kind of error is sent when a resource is not found.
- Does the configuration work properly.
**/
