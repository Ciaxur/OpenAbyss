package entity

import "golang.org/x/crypto/openpgp/packet"

var (
	Store EntityStore = EntityStore{
		Keys:   make(map[string]Entity),
		Length: 0,
	}
)

type Entity struct {
	PublicKey  *packet.PublicKey
	PrivateKey *packet.PrivateKey
	Name       string
}
