package main

import (
	"log"
	"net/http"
)

func apiHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// TODO
	case "PUT":
		// TODO
	case "DELETE":
		// TODO
	default:
		// TODO for GET
	}

}

func main() {
	http.HandleFunc("/api", apiHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
