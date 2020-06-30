package config

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Get : Gets config variables from the file
func Get(key string) string {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()

	if err != nil {
		log.Fatalf("Error while reading config file %s", err)
	}
	// To check if environement is test
	if strings.HasSuffix(os.Args[0], ".test") && key == "DB" {
		key = "TEST" + key
	}
	value, _ := viper.Get(key).(string)
	return value
}
