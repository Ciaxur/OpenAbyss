package utils

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
)

/**
 * Helper function that checks if a file exists
 *  given a file name returning the state of
 *  the existance of the file
 */
func FileExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	return false
}

// Exports the public/private key to a file given
//  the filename and entity to export
func ExportKeys(keyPair *rsa.PrivateKey, dir string, keyname string) error {
	// Attempt to create the directory (in case not avail)
	os.Mkdir(dir, 0777)

	// Open Files to write to
	privKeyFile, err := os.Create(filepath.Join(dir, keyname))
	if err != nil {
		return err
	}
	defer privKeyFile.Close()
	pubKeyFile, err := os.Create(filepath.Join(dir, keyname+".pub"))
	if err != nil {
		return err
	}
	defer pubKeyFile.Close()

	// Encode Private Key to file
	pem.Encode(privKeyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(keyPair),
	})

	// Encode Public Key to file
	pem.Encode(pubKeyFile, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&keyPair.PublicKey),
	})

	return nil
}
