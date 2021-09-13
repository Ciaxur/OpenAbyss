package entity

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"io/ioutil"
	"log"
	"openabyss/utils"
)

// Attempts to encrypt given data buffer to given destination returning the state of
//  the encryption
func Encrypt(data []byte, destPath string, sk *rsa.PrivateKey) bool {
	eBuffer, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &sk.PublicKey, data, []byte("OAEP Encrypted"))
	if err != nil {
		log.Println("could not encrypt file:", err.Error())
		return false
	} else {
		err := ioutil.WriteFile(destPath, []byte(base64.StdEncoding.EncodeToString(eBuffer)), 0644)
		if !utils.HandleErr(err, "could not write encrypted data to '"+destPath) {
			return false
		}
		log.Printf("Encrypted file successfuly: '%s'\n", destPath)
	}

	return true
}

// Attempts to decrypt given encrypted data to destination returning the state of
//  the decryption
func Decrypt(data []byte, destPath string, sk *rsa.PrivateKey) error {
	fileBuffer, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		log.Println("could not base64 decode file:", err.Error())
		return err
	}

	eBuffer, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, sk, fileBuffer, []byte("OAEP Encrypted"))
	if err != nil {
		log.Println("could not decrypt file:", err.Error())
		return err
	} else {
		err := ioutil.WriteFile(destPath, eBuffer, 0644)
		if !utils.HandleErr(err, "could not write decrypt data to '"+destPath) {
			return err
		}
		log.Printf("Decrypted file successfuly: '%s'\n", destPath)
	}
	return nil
}
