package password

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

var defaultSaltSize = 16
var defaultNumIterations = 2000
var defaultHashLength = 256

type SecurePassword struct {
	IterationCount int    `json:"-"`
	Salt           []byte `json:"-"`
	Hash           []byte `json:"-"`
}

func NewSecurePassword(password string) (SecurePassword, error) {

	var secure SecurePassword
	var err error

	// Verify that password is not empty
	if strings.TrimSpace(password) == "" {
		return secure, fmt.Errorf("password is empty")
	}

	secure.IterationCount = defaultNumIterations

	// Get Salt
	secure.Salt, err = getSalt(defaultSaltSize)
	if err != nil {
		return secure, err
	}

	// Get Hash
	secure.Hash = getHash(password, secure.Salt, secure.IterationCount)

	return secure, nil

}

func (sp *SecurePassword) ValidatePassword(password string) bool {
	newHash := getHash(password, sp.Salt, sp.IterationCount)
	return areEqual(sp.Hash, newHash)
}

func areEqual(h1, h2 []byte) bool {
	if len(h1) != len(h2) {
		return false
	}
	for i := 0; i < len(h1); i++ {
		if h1[i] != h2[i] {
			return false
		}
	}
	return true
}

func getHash(password string, salt []byte, numIterations int) []byte {
	hash := pbkdf2.Key([]byte(password), salt, numIterations, defaultHashLength, sha256.New)
	return hash
}

func getSalt(size int) ([]byte, error) {
	if size < 1 {
		return nil, fmt.Errorf("invalid salt size %d", size)
	}
	salt := make([]byte, size)
	n, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		return nil, err
	}
	if n != size {
		return nil, fmt.Errorf("number of bytes read %d are not the same as expected size of salt %d", n, size)
	}
	return salt, nil
}
