package helpers

import (
	"log"

	"github.com/mediocregopher/radix/v3"
)

// GetCache connects to the Redis instance
func GetCache() *radix.Pool {
	c, err := radix.NewPool("tcp", "127.0.0.1:6379", 10)
	if err != nil {
		log.Printf("%v", err)
	}
	// PING the server and recover from  panic
	// TODO: enable auth
	return c
}
