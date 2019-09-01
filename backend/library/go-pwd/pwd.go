package password

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

// SecurePassword represents a one-way encypted form of a password. It includes the hash, and
// other information that can be used to regenerate the hash and validate a password.
type SecurePassword struct {
	IterationCount int    `json:"-"`
	Salt           []byte `json:"-"`
	Hash           []byte `json:"-"`
}

// Config contains configuration fields that are required inorder to generate a new SecurePassword.
type Config struct {
	saltSize, numIterations, hashLength int
}

var defaultConfig = Config{
	saltSize:      16,
	numIterations: 2000,
	hashLength:    256,
}

// NewSecurePassword uses the default config to generate a new SecurePassword
func NewSecurePassword(password string) (SecurePassword, error) {
	config := defaultConfig
	return config.NewSecurePassword(password)
}

// NewSecurePassword uses the receiver config to generate a new SecurePassword
func (c Config) NewSecurePassword(password string) (SecurePassword, error) {
	var secure SecurePassword
	var err error

	// Verify that password is not empty
	if strings.TrimSpace(password) == "" {
		return secure, fmt.Errorf("password is empty")
	}

	secure.IterationCount = c.numIterations

	// Get Salt
	secure.Salt, err = getSalt(c.saltSize)
	if err != nil {
		return secure, err
	}

	// Get Hash
	secure.Hash = getHash(password, secure.Salt, secure.IterationCount, c.hashLength)

	return secure, nil
}

// ValidatePassword takes a password and validates that the password produces the same hash i.e. password is valid
func ValidatePassword(sp SecurePassword, password string) bool {
	newHash := getHash(password, sp.Salt, sp.IterationCount, len(sp.Hash))
	return areEqual(sp.Hash, newHash)
}

// GetHash generates a hash of specified hashLength given the password, the salt and number of iterations to use
// to produce the hash
func GetHash(password string, salt []byte, numIterations int, hashLength int) []byte {
	return getHash(password, salt, numIterations, hashLength)
}

// getHash generates a hash of specified hashLength given the password, the salt and number of iterations to use
// to produce the hash
func getHash(password string, salt []byte, numIterations int, hashLength int) []byte {
	hash := pbkdf2.Key([]byte(password), salt, numIterations, hashLength, sha256.New)
	return hash
}

// GetSalt generates a randomly generated salt of length size
func GetSalt(size int) ([]byte, error) {
	return getSalt(size)
}

// getSalt generates a randomly generated salt of length size
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

// areEqual takes two slice of bytes (i.e. two hashes) and returns true if the both hash are equal
// and false if they are not equal.
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
