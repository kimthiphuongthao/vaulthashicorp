package hash

import (
	"crypto/sha256"
	"crypto/sha512"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2"

	"github.com/kimthiphuongthao/vaulthashicorp/internal/legacycred/config"
)

func Hash(cfg config.Config, hashType, plaintext string) (string, error) {
	switch hashType {
	case "bcrypt":
		return hashBcrypt(cfg, plaintext)
	case "sha256":
		return hashSHA(cfg, plaintext, sha256.New)
	case "sha512":
		return hashSHA(cfg, plaintext, sha512.New)
	case "pbkdf2":
		return hashPBKDF2(cfg, plaintext)
	default:
		return "", fmt.Errorf("unsupported hash_type")
	}
}

func hashBcrypt(cfg config.Config, plaintext string) (string, error) {
	cost := cfg.BcryptCost
	if cost == 0 {
		cost = config.DefaultConfig().BcryptCost
	}
	out, err := bcrypt.GenerateFromPassword([]byte(plaintext), cost)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func hashSHA(cfg config.Config, plaintext string, newHash func() hash.Hash) (string, error) {
	salt := cfg.ShaSalt
	pos := cfg.ShaSaltPosition
	if pos == "" {
		pos = "prefix"
	}

	var input string
	switch pos {
	case "prefix":
		input = salt + plaintext
	case "suffix":
		input = plaintext + salt
	default:
		return "", fmt.Errorf("invalid sha_salt_position")
	}

	h := newHash()
	_, _ = h.Write([]byte(input))
	sum := h.Sum(nil)

	switch cfg.ShaOutputEncoding {
	case "", "hex":
		return hex.EncodeToString(sum), nil
	case "base64":
		return base64.StdEncoding.EncodeToString(sum), nil
	default:
		return "", fmt.Errorf("invalid sha_output_encoding")
	}
}

func hashPBKDF2(cfg config.Config, plaintext string) (string, error) {
	iters := cfg.PBKDF2Iterations
	if iters == 0 {
		iters = config.DefaultConfig().PBKDF2Iterations
	}
	keyLen := cfg.PBKDF2KeyLength
	if keyLen == 0 {
		keyLen = config.DefaultConfig().PBKDF2KeyLength
	}
	saltLen := cfg.PBKDF2SaltLength
	if saltLen == 0 {
		saltLen = config.DefaultConfig().PBKDF2SaltLength
	}
	if saltLen <= 0 {
		return "", fmt.Errorf("invalid pbkdf2_salt_length")
	}

	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	var dk []byte
	algo := cfg.PBKDF2Hash
	switch cfg.PBKDF2Hash {
	case "", "sha256":
		algo = "sha256"
		dk = pbkdf2.Key([]byte(plaintext), salt, iters, keyLen, sha256.New)
	case "sha512":
		algo = "sha512"
		dk = pbkdf2.Key([]byte(plaintext), salt, iters, keyLen, sha512.New)
	default:
		return "", fmt.Errorf("invalid pbkdf2_hash")
	}

	// Store as a self-describing string so legacy update code can parse if needed.
	return fmt.Sprintf("pbkdf2$%s$%d$%s$%s",
		algo,
		iters,
		base64.StdEncoding.EncodeToString(salt),
		base64.StdEncoding.EncodeToString(dk),
	), nil
}
