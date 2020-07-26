package helpers

import (
	"fmt"
	"log"

	"github.com/mediocregopher/radix/v3"
)

// InitCache connects to the Redis instance
func InitCache() *radix.Pool {
	c, err := radix.NewPool("tcp", "127.0.0.1:6379", 10)
	if err != nil {
		log.Printf("%v", err)
	}
	fmt.Println("INFO: PINGing Redis")
	err = c.Do(radix.Cmd(nil, "PING"))
	if err != nil {
		log.Printf("%v", err)
	}
	fmt.Println("INFO: Connected to Redis")
	return c
}
