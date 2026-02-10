package legacycred

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	charsetAlnum = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charsetASCII = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}:,.?"
)

func generatePassword(length int, charset string) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("invalid password length")
	}

	alphabet := charsetAlnum
	switch charset {
	case "alnum", "":
		alphabet = charsetAlnum
	case "ascii":
		alphabet = charsetASCII
	default:
		alphabet = charset
	}

	out := make([]byte, length)
	max := big.NewInt(int64(len(alphabet)))
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		out[i] = alphabet[n.Int64()]
	}
	return string(out), nil
}
