package config

import "os"

// SetVariable : Sets Environment Variables
func SetVariable() {
	os.Setenv("PORT", "8080")
	os.Setenv("MONGODB_URI", "mongodb://localhost:27017")
	os.Setenv("DB", "phonebook")
	os.Setenv("COLLECTION", "contacts")
}
