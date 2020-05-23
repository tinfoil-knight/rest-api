package main

import "os"

func setVariable() {
	os.Setenv("PORT", "8080")
	os.Setenv("MONGODB_URI", "mongodb://localhost:27017")
	os.Setenv("DBNAME", "phonebook")
}
