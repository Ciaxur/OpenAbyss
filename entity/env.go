package entity

import (
	"crypto/cipher"
	"crypto/rsa"
	"openabyss/server/storage"
	"path"
)

var (
	Store EntityStore = EntityStore{
		Keys:   make(map[string]Entity),
		Length: 0,
	}
	KeyStorePath = path.Join(storage.InternalStoragePath, "keys")
)

type Entity struct {
	PublicKey       *rsa.PublicKey
	PrivateKey      *rsa.PrivateKey
	Cipher          *cipher.Block
	AesEncryptedKey []byte
	Name            string
}
