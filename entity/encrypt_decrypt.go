package entity

import (
	"bytes"
	"crypto/aes"
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

// Helper funciton that decrypts the given Base64 Encoded and Decrypted CipherKey
//  returning the cipher block from the encrypted key
func decryptAesCipherBlock(pk *rsa.PrivateKey, encCipherKey []byte) (cipher.Block, error) {
	// Decrypt Cipher
	if cipherKey, err := base64.StdEncoding.DecodeString(string(encCipherKey)); err != nil {
		return nil, err
	} else {
		// Use private key to decrypt AES Key
		cipherKeyBuffer := bytes.NewBufferString("")
		if err := Decrypt(cipherKey, cipherKeyBuffer, pk); err != nil {
			return nil, err
		}

		// Create Cipher Block from obtained key
		if cipher, err := aes.NewCipher(cipherKeyBuffer.Bytes()); err != nil {
			return nil, err
		} else {
			return cipher, nil
		}
	}
}

// Attempts to encrypt given data writer using cipher to given destination returning the state of
//  the encryption. Obtaining Cipher from encrypted key in Enity object.
func CipherEncrypt(data []byte, destWriter io.Writer, entity *Entity) error {
	// Decrypt Cipher
	c, err := decryptAesCipherBlock(entity.PrivateKey, entity.AesEncryptedKey)
	if err != nil {
		return err
	}

	// Encrypt the data
	// Create iv & prepend to ciphertext
	cipherText := make([]byte, c.BlockSize()+len(data))
	iv := cipherText[:c.BlockSize()]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	// Start Encrypting
	stream := cipher.NewCFBEncrypter(c, iv)
	stream.XORKeyStream(cipherText[c.BlockSize():], data)

	if len(cipherText) == 0 {
		return errors.New("internal error: failed to encrypt data using cipher")
	}

	// Write Encrypted data -> Base64 -> IO Writer
	_, err = destWriter.Write([]byte(base64.StdEncoding.EncodeToString(cipherText)))
	if !utils.HandleErr(err, "could not write encrypted data to writer") {
		return err
	}

	return nil
}

// Attempts to decrypt given encrypted data using cipher to destination returning the state of
//  the decryption. Obtaining Cipher from encrypted key in Enity object.
func CipherDecrypt(data []byte, destWriter io.Writer, entity *Entity) error {
	// Decrypt Cipher
	c, err := decryptAesCipherBlock(entity.PrivateKey, entity.AesEncryptedKey)
	if err != nil {
		return err
	}

	// Convert base64 -> ciphertext
	cipherText, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		log.Println("could not base64 decode file:", err.Error())
		return err
	}

	// Decrypt the ciphertext
	if len(cipherText) < c.BlockSize() {
		return errors.New("ciphertext too short")
	}
	plainText := make([]byte, len(cipherText)-c.BlockSize())

	// Extract iv from prepended ciphertext
	iv := cipherText[:c.BlockSize()]

	// Start Decrypting
	stream := cipher.NewCFBDecrypter(c, iv)
	stream.XORKeyStream(plainText, cipherText[c.BlockSize():])

	if len(plainText) == 0 {
		return errors.New("internal error: failed to decrypt data using cipher")
	}

	// Write Decrypter Data -> IO Writer
	_, err = destWriter.Write(plainText)
	if !utils.HandleErr(err, "could not write decrypted data to writer") {
		return err
	}
	return nil
}
