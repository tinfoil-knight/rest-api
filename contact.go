package main

import (
	"fmt"
	"log"
	"net/http"
)

// Contact : Struct for Storing Contacts
type Contact struct {
	Name  string
	Phone string
}

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

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s\n", r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func main() {
	httpPort := 8080
	http.HandleFunc("/api", apiHandler)
	fmt.Println("Server starting on http://localhost:8080")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", httpPort), logRequest(http.DefaultServeMux)))
}
