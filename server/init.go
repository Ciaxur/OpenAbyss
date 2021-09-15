package main

import (
	"openabyss/entity"
	"openabyss/server/storage"
)

func Init() {
	// Load Entity Store
	entity.Init()

	// Load Storage
	storage.Init()
}
