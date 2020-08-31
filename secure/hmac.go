package secure

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"math/rand"
	"time"
)

func GenerateSalt() string {
	rand.Seed(time.Now().UTC().UnixNano())
	random := string(rand.Intn(1000 * 1000))
	salt := sha256.Sum256([]byte(random))
	return hex.EncodeToString(salt[12:])
}

func ComputeHmac256(password, salt, secretKey string) string {
	key := []byte(secretKey)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(password + salt))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
