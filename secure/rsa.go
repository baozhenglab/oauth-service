package secure

import (
	"crypto/rand"
	"crypto/rsa"
)

func GenerateRSAPrivateKey() *rsa.PrivateKey {
	key, _ := rsa.GenerateKey(rand.Reader, 1024)
	return key
}
