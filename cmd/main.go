package main

import (
	"log"
	"openabyss/entity"
	"openabyss/utils"
	"os"

	"github.com/manifoldco/promptui"
)

func main() {
	// Initiate Entity Store
	entity.Init()

	for isDone := false; !isDone; {
		prompt := promptui.Select{
			Label: "Actions",
			Items: []string{
				"Keys",
				"Encrypt/Decrypt",
				"Exit",
			},
		}

		_, result, err := prompt.Run()
		if utils.IsErrorSIGINT(err) {
			log.Println("Exiting...")
			os.Exit(0)
		}
		utils.HandleErr(err, "main.prompt wth")

		if result == "Keys" {
			ShowKeysMenu()
		} else if result == "Encrypt/Decrypt" {
			if entity.Store.Length == 0 {
				log.Println("No available keys. Generate one to encrypt/decrypt")
			} else {
				ShowEncDecryptMenu()
			}
		} else if result == "Exit" {
			isDone = true
		}
	}
}
