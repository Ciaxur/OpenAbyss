package entity

import (
	"crypto/rsa"
)

var (
	Store EntityStore = EntityStore{
		Keys:   make(map[string]Entity),
		Length: 0,
	}
)

type Entity struct {
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
	Name       string
}
