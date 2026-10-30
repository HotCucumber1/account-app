package secret

import (
	"crypto/sha256"
)

func Digest(value string) []byte {
	sum := sha256.Sum256([]byte(value))
	return sum[:]
}
