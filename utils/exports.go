package utils

import (
	"os"
	"path/filepath"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
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
func ExportKeys(pk *packet.PublicKey, sk *packet.PrivateKey, dir string, keyname string) error {
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

	// Open Armor Encode Writers
	pWriter, err := armor.Encode(pubKeyFile, openpgp.PublicKeyType, make(map[string]string))
	if err != nil {
		return err
	}
	defer pWriter.Close()
	sWriter, err := armor.Encode(privKeyFile, openpgp.PrivateKeyType, make(map[string]string))
	if err != nil {
		return err
	}
	defer sWriter.Close()

	// Encode keys to writers
	pk.Serialize(pWriter)
	sk.Serialize(sWriter)

	return nil
}
