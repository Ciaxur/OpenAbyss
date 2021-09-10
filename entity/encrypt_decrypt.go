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

// Attempts to encrypt given file to given destination returning the state of
//  the encryption
func Encrypt(filePath string, destPath string, sk *rsa.PrivateKey) bool {
	fileBuffer, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Println("couldn't read file")
		return false
	} else {
		eBuffer, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &sk.PublicKey, fileBuffer, []byte("OAEP Encrypted"))
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
	}

	return true
}

// Attempts to decrypt given destination returning the state of
//  the decryption
func Decrypt(filePath string, destPath string, sk *rsa.PrivateKey) bool {
	fileBuffer, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Println("couldn't read file")
		return false
	} else {
		fileBuffer, err = base64.StdEncoding.DecodeString(string(fileBuffer))
		if err != nil {
			log.Println("could not base64 decode file:", err.Error())
			return false
		}

		eBuffer, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, sk, fileBuffer, []byte("OAEP Encrypted"))
		if err != nil {
			log.Println("could not decrypt file:", err.Error())
			return false
		} else {
			err := ioutil.WriteFile(destPath, eBuffer, 0644)
			if !utils.HandleErr(err, "could not write decrypt data to '"+destPath) {
				return false
			}
			log.Printf("Decrypted file successfuly: '%s'\n", destPath)
		}
	}

	return true
}
