package entity

import (
	"io/ioutil"
	"log"
	"strings"
)

// Attempts to load keys into Entity object
func LoadKeys() error {
	files, err := ioutil.ReadDir("keys")
	e := Entity{}
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() {
			if strings.HasSuffix(file.Name(), "pub") {
				e.PublicKey = DecodePublicKey("keys", file.Name())
			} else {
				e.PrivateKey = DecodePrivateKey("keys", file.Name())
				e.Name = file.Name()
			}
		}

		// Check Entity is done
		if e.PrivateKey != nil && e.PublicKey != nil {
			if !ValidateKeyPair(e.PrivateKey) {
				log.Printf("Invalid Key Pair '%s' Public[%d] Private[%d]\n", e.Name, e.PublicKey.Size(), e.PrivateKey.Size())
			} else {
				Store.Add(e)
			}
			e = Entity{} // reset
		}
	}

	return nil
}
