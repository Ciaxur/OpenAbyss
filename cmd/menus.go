package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"openabyss/entity"
	"openabyss/utils"
	"os"
	"path"
	"regexp"

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

		e1, err := entity.GenerateKeys(entity.KeyStorePath, keyname, 2048)
		if err == nil {
			log.Println("Generated Key:", e1.Name)
			entity.Store.Add(e1)
		}
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
	var destPath string
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

		p2 := promptui.Prompt{
			Label:     "Path to destination",
			AllowEdit: true,
			Validate: func(s string) error {
				if utils.FileExists(s) {
					return errors.New("path already exists")
				}
				return nil
			},
		}
		destPath, _ = p2.Run()

		// Handle Path and destination creation
		// Create directory path if not available
		//  - Case1: Create directory path for dest path being a directory
		//  - Case2: Create directory path for dest path's parent
		if isDirPath, _ := regexp.MatchString("/$", destPath); isDirPath {
			// Create path if doesn't exist
			if !utils.DirExists(destPath) {
				os.MkdirAll(destPath, 0755)
			}

			// Adjust Destination path with file end
			destPath += "_output_" + path.Base(filePath)
		} else if !utils.DirExists(path.Dir(destPath)) {
			os.MkdirAll(path.Dir(destPath), 0755)
		}
	}

	if val == "Encrypt file" {
		sk := SelectPrivateKey()
		if sk != nil {
			if data, err := ioutil.ReadFile(filePath); err != nil {
				log.Println("could not read given file:", err)
			} else {
				if destWriter, err := os.Create(destPath); err != nil {
					utils.HandleErr(err, "failed to create path")
				} else {
					entity.Encrypt(data, destWriter, sk)
				}
			}
		} else {
			log.Fatalln("no private key selected")
		}
	} else if val == "Decrypt file" {
		sk := SelectPrivateKey()
		if sk != nil {
			if data, err := ioutil.ReadFile(filePath); err != nil {
				log.Println("could not read given file:", err)
			} else {
				if destWriter, err := os.Create(destPath); err != nil {
					utils.HandleErr(err, "failed to create path")
				} else {
					entity.Decrypt(data, destWriter, sk)
				}
			}
		} else {
			log.Fatalln("no private key selected")
		}
	}
}
