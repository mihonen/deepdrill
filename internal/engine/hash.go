package engine




import (
	"crypto/sha256"
	"encoding/hex"
)



func GenerateHash(tree string) string {
	h := sha256.Sum256([]byte(tree))
	return hex.EncodeToString(h[:])
}


