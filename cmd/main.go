package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"openabyss/entity"
	"strings"

	"github.com/manifoldco/promptui"
)

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func showKeysMenu() {
	prompt := promptui.Select{
		Label: "Keys",
		Items: []string{
			"Generate New Key Pairs",
			"List Keys",
			"Go Back",
		},
	}
	idx, _, err := prompt.Run()
	if utils.IsErrorSIGINT(err) {
		return
	}
	utils.HandleErr(err, "show keys prompt error")

	// GENERATE KEYS
	if idx == 0 {
		// Prompt for keyname
		p := promptui.Prompt{
			Label:     "Keyname",
			AllowEdit: true,
			Validate: func(s string) error {
				// Check valid keyname & keyname existance
				if len(s) == 0 {
					return errors.New("no empty name allowed")
				} else if entity.Store.Has(s) {
					return errors.New("keyname already exists")
				}
				return nil
			},
		}
		keyname, err := p.Run()
		if utils.IsErrorSIGINT(err) {
			return
		}
		utils.HandleErr(err, "keyname prompt error")

		e1 := entity.GenerateKeys("keys", keyname, 2048)
		e1.Name = keyname
		log.Println("Generated Key:", e1.Name)
		entity.Store.Add(e1)
	} else if idx == 1 { // LIST KEYS
		if entity.Store.Length == 0 {
			log.Println("No keys stored")
		} else {
			for k, v := range entity.Store.Keys {
				fmt.Printf("[%s]: %s\n", k, v.Name)
			}
		}
	}
}

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
			showKeysMenu()
		} else if result == "Encrypt/Decrypt" {
			log.Println("WIP")
		} else if result == "Exit" {
			isDone = true
		}
	}
}
