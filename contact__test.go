package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func runServer(fn func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(fn))
}

func Test__HTTPStatus(t *testing.T) {
	ts := runServer(apiHandler)
	url := ts.URL + "/api/"
	res, _ := http.Get(url)

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected StatusCode: %v, Received StatusCode: %v", http.StatusOK, res.StatusCode)
	}

	ts.Close()
}
