package entity

import "log"

func Init() {
	// Load in Keys if available
	err := LoadKeys()
	if err == nil {
		log.Printf("Loaded in %d keys\n", Store.Length)
	} else {
		log.Println("No keys loaded")
	}
}
