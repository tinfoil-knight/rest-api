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
	"go.mongodb.org/mongo-driver/bson"
)

var results []*Contact

func runServer(fn func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(fn))
}

func initDB() {
	client = getClient(config.Get("MONGODB_URI"))
	err := client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	collection := client.Database(config.Get("TESTDB")).Collection(config.Get("COLLECTION"))
	collection.DeleteMany(context.TODO(), bson.M{})
	contact1 := Contact{Name: "Jay Randall", Phone: "9087453245"}
	contact2 := Contact{Name: "Reinne Parsley", Phone: "8904576732"}
	contacts := []interface{}{contact1, contact2}
	collection.InsertMany(context.TODO(), contacts)
	cur, _ := collection.Find(context.TODO(), bson.D{{}})
	for cur.Next(context.TODO()) {
		var elem Contact
		cur.Decode(&elem)
		results = append(results, &elem)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	cur.Close(context.TODO())
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
		t.Errorf("StatusCode | Expected: %v, Received: %v", http.StatusOK, res.StatusCode)
	}
	var contacts *[]Contact
	json.NewDecoder(res.Body).Decode(contacts)
	log.Println(contacts)
	res.Body.Close()
	ts.Close()
}

func Test__GetOneByID(t *testing.T) {
	initDB()
	ts := runServer(apiHandler)
	contact := results[0]
	id := (contact.ID).Hex()
	fmt.Println(id)
	url := ts.URL + "/api/" + id
	res, err := http.Get(url)
	if err != nil {
		t.Errorf("%s", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("HTTPStatusCode | Expected: %v, Received: %v", http.StatusOK, res.StatusCode)
	}
	var resContact Contact
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
		fmt.Printf("%s", err.Error())
	}
	res, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Errorf("%s", err)
	}
	if res.StatusCode != http.StatusCreated {
		t.Errorf("Expected StatusCode: %v, Received StatusCode: %v", http.StatusCreated, res.StatusCode)
	}
	res.Body.Close()
	ts.Close()
}

func Test__ChangeOneByID(t *testing.T) {
	initDB()

}

func Test__DeleteOneByID(t *testing.T) {
	initDB()
}
