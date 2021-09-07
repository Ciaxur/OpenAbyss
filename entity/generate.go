package entity

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"openabyss/utils"
	"os"
	"path"
	"time"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

/**
 * Validates the keypair being valid returning the state of validity
 */
func ValidateKeyPair(pk *packet.PublicKey, sk *packet.PrivateKey) bool {
	return pk.KeyId == sk.KeyId
}

func GenerateKeys(dir string, keyname string, bits int) Entity {
	// Generate & Create RSA Keys
	sk, err := rsa.GenerateKey(rand.Reader, bits)
	utils.HandleErr(err, "error generating Private key")

	gpgSkKey := packet.NewRSAPrivateKey(time.Now(), sk)
	gpgPbKey := packet.NewRSAPublicKey(time.Now(), &sk.PublicKey)

	// Export keys to file
	err = utils.ExportKeys(gpgPbKey, gpgSkKey, dir, keyname)
	utils.HandleErr(err, "could no export keys to file")

	return Entity{
		PrivateKey: gpgSkKey,
		PublicKey:  gpgPbKey,
	}
}

func DecodePublicKey(dir string, keyname string) *packet.PublicKey {
	// Open the file
	keyFile, err := os.Open(path.Join(dir, keyname))
	utils.HandleErr(err, "could not read key file")
	defer keyFile.Close()

	// Decode the file
	block, err := armor.Decode(keyFile)
	utils.HandleErr(err, "couldn't decode keyfile")
	if block.Type != openpgp.PublicKeyType {
		utils.HandleErr(errors.New("not public key type"), "")
	}

	// Read & Parse the Key
	pktReader := packet.NewReader(block.Body)
	pkt, err := pktReader.Next()
	utils.HandleErr(err, "could not read packet")
	key, ok := pkt.(*packet.PublicKey)
	if !ok {
		utils.HandleErr(errors.New("failed to convert packet to public key type"), "")
	}

	return key
}

func DecodePrivateKey(dir string, keyname string) *packet.PrivateKey {
	// Open the file
	keyFile, err := os.Open(path.Join(dir, keyname))
	utils.HandleErr(err, "could not read key file")
	defer keyFile.Close()

	// Decode the file
	block, err := armor.Decode(keyFile)
	utils.HandleErr(err, "couldn't decode keyfile")
	if block.Type != openpgp.PrivateKeyType {
		utils.HandleErr(errors.New("not private key type"), "")
	}

	// Read & Parse the Key
	pktReader := packet.NewReader(block.Body)
	pkt, err := pktReader.Next()
	utils.HandleErr(err, "could not read packet")
	key, ok := pkt.(*packet.PrivateKey)
	if !ok {
		utils.HandleErr(errors.New("failed to convert packet to private key type"), "")
	}

	return key
}
