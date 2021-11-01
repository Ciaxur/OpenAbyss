package entity

import (
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"openabyss/utils"
)

// Attempts to encrypt given data writer to given destination returning the state of
//  the encryption
func Encrypt(data []byte, destWriter io.Writer, sk *rsa.PrivateKey) error {
	eBuffer, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &sk.PublicKey, data, []byte("OAEP Encrypted"))
	if err != nil {
		log.Println("could not encrypt file:", err.Error())
		return err
	} else {
		_, err := destWriter.Write([]byte(base64.StdEncoding.EncodeToString(eBuffer)))
		if !utils.HandleErr(err, "could not write encrypted data to writer") {
			return err
		}
	}

	return nil
}

// Attempts to decrypt given encrypted data to destination returning the state of
//  the decryption
func Decrypt(data []byte, destWriter io.Writer, sk *rsa.PrivateKey) error {
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
		_, err := destWriter.Write(eBuffer)
		if !utils.HandleErr(err, "could not write decrypted data to writer") {
			return err
		}
	}
	return nil
}

// Attempts to encrypt given data writer using cipher to given destination returning the state of
//  the encryption
func CipherEncrypt(data []byte, destWriter io.Writer, cipher *cipher.Block) error {
	log.Println(cipher)
	c := *cipher

	// Encrypt the data
	encBuffer := make([]byte, len(data))
	c.Encrypt(data, encBuffer)

	if len(encBuffer) == 0 {
		return errors.New("internal error: failed to encrypt data using cipher")
	}

	// Write Encrypted data -> Base64 -> IO Writer
	_, err := destWriter.Write([]byte(base64.StdEncoding.EncodeToString(encBuffer)))
	if !utils.HandleErr(err, "could not write encrypted data to writer") {
		return err
	}

	return nil
}

// Attempts to decrypt given encrypted data using cipher to destination returning the state of
//  the decryption
func CipherDecrypt(data []byte, destWriter io.Writer, cipher *cipher.Block) error {
	fileBuffer, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		log.Println("could not base64 decode file:", err.Error())
		return err
	}
	c := *cipher

	// Decrypt the data
	dataBuffer := make([]byte, len(data))
	c.Decrypt(dataBuffer, fileBuffer)
	log.Println(string(dataBuffer))

	if len(dataBuffer) == 0 {
		return errors.New("internal error: failed to decrypt data using cipher")
	}

	// Write Decrypter Data -> IO Writer
	_, err = destWriter.Write(dataBuffer)
	if !utils.HandleErr(err, "could not write decrypted data to writer") {
		return err
	}
	return nil
}
