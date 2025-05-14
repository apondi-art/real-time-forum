package main

import (
	"log"

	"real-time-forum/internals/database"
)

func main() {
	if _, err := database.New("./internals/database/real_time.db"); err != nil {
		log.Fatal(err)
	}
}
