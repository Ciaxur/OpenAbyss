package main

import (
	"crypto/rsa"
	"openabyss/entity"
	"openabyss/utils"
	"reflect"

	"github.com/manifoldco/promptui"
)

// Displays a prompt tui to select a private key to use
//  returns nil if no private key selected or none found
func SelectPrivateKey() *rsa.PrivateKey {
	p := promptui.Select{
		Label: "Select Key",
		Items: reflect.ValueOf(entity.Store.Keys).MapKeys(),
	}
	_, keyname, err := p.Run()
	if utils.IsErrorSIGINT(err) {
		return nil
	}

	return entity.Store.Keys[keyname].PrivateKey
}
