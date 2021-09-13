package entity

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"openabyss/utils"
	"os"
	"path"
)

/**
 * Validates the keypair being valid returning the state of validity
 */
func ValidateKeyPair(sk *rsa.PrivateKey) bool {
	if err := sk.Validate(); err != nil {
		return false
	}
	return true
}

func GenerateKeys(dir string, keyname string, bits int) Entity {
	// Generate & Create RSA Keys
	rsaKeyPair, err := rsa.GenerateKey(rand.Reader, bits)
	utils.HandleErr(err, "error generating RSA Keypair")

	// Export keys to file
	err = utils.ExportKeys(rsaKeyPair, dir, keyname)
	utils.HandleErr(err, "could no export keys to file")

	return Entity{
		PrivateKey: rsaKeyPair,
		PublicKey:  &rsaKeyPair.PublicKey,
		Name:       keyname,
	}
}

func DecodePublicKey(dir string, keyname string) *rsa.PublicKey {
	// Open the file
	keyFile, err := os.Open(path.Join(dir, keyname))
	utils.HandleErr(err, "could not read key file")
	defer keyFile.Close()

	// Decode the file
	rawFileBytes, err := ioutil.ReadAll(keyFile)
	utils.HandleErr(err, "could not read public key file")

	decodedKey, _ := pem.Decode(rawFileBytes)
	pk, err := x509.ParsePKCS1PublicKey(decodedKey.Bytes)
	utils.HandleErr(err, "couldn't parse public key")

	return pk
}

func DecodePrivateKey(dir string, keyname string) *rsa.PrivateKey {
	// Open the file
	keyFile, err := os.Open(path.Join(dir, keyname))
	utils.HandleErr(err, "could not read key file")
	defer keyFile.Close()

	// Decode the file
	rawFileBytes, err := ioutil.ReadAll(keyFile)
	utils.HandleErr(err, "could not read private key file")

	decodedKey, _ := pem.Decode(rawFileBytes)
	sk, err := x509.ParsePKCS1PrivateKey(decodedKey.Bytes)
	utils.HandleErr(err, "couldn't parse private key")

	return sk
}
