package utils

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
)

// Marshalls ED25519 Public Key to pem format
func ED25519_to_pem(pk ed25519.PublicKey) []byte {
	// ED25519 Public Key -> PKIX
	b, _ := x509.MarshalPKIXPublicKey(pk)

	// PKIX -> PEM
	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: b,
	}
	return pem.EncodeToMemory(block)
}

// Marshalls ED25519 Private Key to pem format
func ED25519_to_pem_sk(sk ed25519.PrivateKey) []byte {
	// ED25519 Private Key -> PKIX
	b, _ := x509.MarshalPKCS8PrivateKey(sk)

	// PKIX -> PEM
	block := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: b,
	}
	return pem.EncodeToMemory(block)
}

// Unmarshalls ED25519 Public Key from pem format
func PEM_to_ed25519(pem_data []byte) ed25519.PublicKey {
	// PEM -> PKIX
	decoded_block, _ := pem.Decode(pem_data)
	decoded_pkix := decoded_block.Bytes

	// PKIX -> ED25519 Public Key
	pk_interface, _ := x509.ParsePKIXPublicKey(decoded_pkix)
	return pk_interface.(ed25519.PublicKey)
}

// Unmarshalls ED25519 Private Key from pem format
func PEM_to_ed25519_sk(pem_data []byte) ed25519.PrivateKey {
	// PEM -> PKIX
	decoded_block, _ := pem.Decode(pem_data)
	decoded_pkix := decoded_block.Bytes

	// PKIX -> ED25519 Private Key
	pk_interface, _ := x509.ParsePKCS8PrivateKey(decoded_pkix)
	return pk_interface.(ed25519.PrivateKey)
}
