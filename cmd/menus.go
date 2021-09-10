package main

import (
	"errors"
	"fmt"
	"log"
	"openabyss/entity"
	"openabyss/utils"

	"github.com/manifoldco/promptui"
)

func ShowKeysMenu() {
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
		entity.Store.Length += 1
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

func ShowEncDecryptMenu() {
	prompt := promptui.Select{
		Label: "Encrypt/Decrypt Menu",
		Items: []string{
			"Encrypt file",
			"Decrypt file",
			"Go back",
		},
	}
	_, val, err := prompt.Run()
	if utils.IsErrorSIGINT(err) {
		return
	}

	// Get path to file
	var filePath string
	if val != "Go back" {
		p := promptui.Prompt{
			Label:     "Path to file",
			AllowEdit: true,
			Validate: func(s string) error {
				if !utils.FileExists(s) {
					return errors.New("file not found")
				}
				return nil
			},
		}
		filePath, _ = p.Run()
	}

	if val == "Encrypt file" {
		sk := SelectPrivateKey()
		if sk != nil {
			entity.Encrypt(filePath, sk)
		} else {
			log.Fatalln("no private key selected")
		}
	} else if val == "Decrypt file" {
		sk := SelectPrivateKey()
		if sk != nil {
			entity.Decrypt(filePath, sk)
		} else {
			log.Fatalln("no private key selected")
		}
	}
}
