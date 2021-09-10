package main

import (
	"io/ioutil"
	"log"
	"openabyss/entity"
	"openabyss/utils"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
)

func loadKeys() error {
	files, err := ioutil.ReadDir("keys")
	e := entity.Entity{}
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() {
			if strings.HasSuffix(file.Name(), "pub") {
				e.PublicKey = entity.DecodePublicKey("keys", file.Name())
			} else {
				e.PrivateKey = entity.DecodePrivateKey("keys", file.Name())
				e.Name = file.Name()
			}
		}

		// Check Entity is done
		if e.PrivateKey != nil && e.PublicKey != nil {
			if !entity.ValidateKeyPair(e.PrivateKey) {
				log.Printf("Invalid Key Pair '%s' Public[%d] Private[%d]\n", e.Name, e.PublicKey.Size(), e.PrivateKey.Size())
			} else {
				entity.Store.Add(e)
			}
			e = entity.Entity{} // reset
		}
	}

	return nil
}

func main() {
	// Load in Keys if available
	err := loadKeys()
	if err == nil {
		log.Printf("Loaded in %d keys\n", entity.Store.Length)
	} else {
		log.Println("No keys loaded")
	}

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
