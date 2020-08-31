package secure

import (
	"encoding/base64"
	"github.com/gtank/cryptopasta"
	"github.com/pkg/errors"
)

type AES struct {
	Key []byte
}

func NewEAS(key []byte) *AES {
	return &AES{Key: key}
}

func (c *AES) Encrypt(plaintext []byte) (string, error) {
	if len(c.Key) < 32 {
		return "", errors.Errorf("Key must be 32 bytes, got %d bytes", len(c.Key))
	}

	var key [32]byte
	copy(key[:], c.Key[:32])

	ciphertext, err := cryptopasta.Encrypt(plaintext, &key)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func (c *AES) Decrypt(ciphertext string) ([]byte, error) {
	if len(c.Key) < 32 {
		return []byte{}, errors.Errorf("Key must be longer 32 bytes, got %d bytes", len(c.Key))
	}

	var key [32]byte
	copy(key[:], c.Key[:32])

	raw, err := base64.URLEncoding.DecodeString(ciphertext)
	if err != nil {
		return []byte{}, errors.WithStack(err)
	}

	plaintext, err := cryptopasta.Decrypt(raw, &key)
	if err != nil {
		return []byte{}, errors.WithStack(err)
	}
	return plaintext, nil
}
