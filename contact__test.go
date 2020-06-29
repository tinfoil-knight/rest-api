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
)

func runServer(fn func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(fn))
}

func initDB() {
	client = getClient(config.Get("MONGODB_URI"))
	err := client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
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
		t.Errorf("Expected StatusCode: %v, Received StatusCode: %v", http.StatusOK, res.StatusCode)
	}
	res.Body.Close()
	ts.Close()
}

func Test__GetOne(t *testing.T) {
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

func Test__DeleteOne(t *testing.T) {

}

func Test__ChangeOne(t *testing.T) {

}
