package main

import (
	"log"
	"real-time-forum/internals/database"
)




func main(){

	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
}