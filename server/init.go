package main

import (
	"log"
	"openabyss/entity"
)

func Init() {
	// Load in Entity Keys
	err := entity.LoadKeys()
	if err != nil {
		log.Fatalln("[init]: keys could not be loading into entity:", err)
	} else {
		log.Printf("[init]: loaded '%d' keys\n", entity.Store.Length)
	}
}
