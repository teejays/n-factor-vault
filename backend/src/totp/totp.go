package totp

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/go-playground/validator.v9"

	"github.com/teejays/n-factor-vault/backend/library/id"
	"github.com/teejays/n-factor-vault/backend/library/orm"
)

func init() {}

/* Notes

[website-totp connection]
	Add a new website for totp
	- requires exchanging of secret keys
	- name of the website

	Get the code for a given website
	- use the secret key of the website to generate the
	code and


*/

var validate *validator.Validate

var gDefaultStartUnixTime int64 // defaults to 0
var gDefaultIntervalInSeconds int64 = 30
var gDefaultCodeLength = 6

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* O R M   M O D E L S
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// Init initializes the service so it can connect with the ORM
func Init() error {
	validate = validator.New()

	err := orm.RegisterModel(&Account{})
	if err != nil {
		return err
	}

	return err
}

// Account represents one TOTP setup for a particular website/service
type Account struct {
	orm.BaseModel       `gorm:"extended"`
	Name                string `gorm:"INDEX"`
	EncryptedPrivateKey []byte `gorm:"NOT NULL"`
	StartUnixTime       int64
	IntervalSeconds     int64
}

// TableName overrides the SQL table name of Account struct
func (a Account) TableName() string {
	return "totp_accounts"
}

// Code is the TOTP authentication code that is passed to the service for auth verification
type Code struct {
	Code     string
	ValidAt  time.Time
	ExpireAt time.Time
}

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* M E T H O D S
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// CreateAccountRequest is the data required to create a new Account
type CreateAccountRequest struct {
	AccountName string `validate:"required"`
	PrivateKey  []byte `validate:"required,min=1"`
}

// CreateAccount creates a TOTP instance
func CreateAccount(req CreateAccountRequest) (Account, error) {
	var a Account

	// Validate the request
	err := validate.Struct(req)
	if err != nil {
		return a, err
	}

	a.Name = req.AccountName

	// Encrypt the private key
	// TODO: Make the encryption safer, but for POC this is fine
	// In order to encrypt, we will need an encryption key. Let's generate this encryption key
	// using a combination of a stored common key + service name + salt

	// Populate the TOTP instance
	key := getEncryptionKey(req.AccountName)
	encryptedPrivateKey, err := encryptWithKey(key, req.PrivateKey)
	if err != nil {
		return a, err
	}
	a.EncryptedPrivateKey = encryptedPrivateKey

	a.StartUnixTime = gDefaultStartUnixTime // Assume that we start from timestamp zero, otherwise find what the standard is, or take it from the request

	a.IntervalSeconds = gDefaultIntervalInSeconds // use this as default interval seconds for now, or take it from request

	// Save it in the database
	err = orm.InsertOne(&a)
	if err != nil {
		return a, err
	}

	return a, nil

}

// GetCodeRequest is the data required to get a code for an account
type GetCodeRequest struct {
	AccountID id.ID `validate:"required"`
}

// GetCode generates a TOTP code for the given TOTP connection
func GetCode(req GetCodeRequest) (Code, error) {

	var c Code

	// Get the TOTP instance for this ID
	var a Account
	found, err := orm.FindByID(req.AccountID, &a)
	if err != nil {
		return c, err
	}
	if !found {
		return c, errors.New("no account found")
	}

	// decrypt the private key of this connection
	key := getEncryptionKey(a.Name)
	privateKey, err := decryptWithKey(key, a.EncryptedPrivateKey)
	if err != nil {
		return c, err
	}

	// generate a TOTP code
	now := time.Now().Unix()
	code, err := getTOTPValue(privateKey, a.StartUnixTime, now, a.IntervalSeconds)
	if err != nil {
		return c, err
	}
	// Get expiry timestamp
	c.Code = code
	numIntervals := int64(math.Ceil(float64(now-a.StartUnixTime) / float64(a.IntervalSeconds)))
	expireAtInUnix := a.StartUnixTime + numIntervals*a.IntervalSeconds
	c.ExpireAt = time.Unix(expireAtInUnix, 0)

	// Get start validity timestamp
	c.ValidAt = c.ExpireAt.Add(-time.Second * time.Duration(a.IntervalSeconds))

	return c, nil
}

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* H E L P E R S
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */
func getEncryptionKey(AccountName string) []byte {
	h := sha256.New()
	h.Write([]byte(AccountName))
	return h.Sum(nil)
}

// encryptWithKey takes a key and data, and encrypts it
// https://tutorialedge.net/golang/go-encrypt-decrypt-aes-tutorial/
func encryptWithKey(key []byte, data []byte) ([]byte, error) {

	// In order to encrypt, we need to do two things:
	// 1. Create a cipher using a key
	// 2. Operate that cipher on the data, using some mode of opeeration

	// 1. Generate the cipher
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 2. Operate that cipher using one of possible modes (GCM)
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	// we need a nonce to operate a GCM, so initialize it and populate it with random data
	// READ: Understand why 'nonce' is important
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		fmt.Println(err)
	}

	// Use seal function to get the encrypted text
	encrypted := gcm.Seal(nonce, nonce, data, nil)

	return encrypted, nil

}

// encryptWithKey takes a key and data, and decrypts it
// https://tutorialedge.net/golang/go-encrypt-decrypt-aes-tutorial/
func decryptWithKey(key []byte, encryptedData []byte) ([]byte, error) {

	// In order to encrypt, we need to do two things:
	// 1. Create a cipher using a key
	// 2. Operate that cipher on the data, using some mode of opeeration

	// 1. Generate the cipher
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 2. Operate that cipher using one of possible modes (GCM)
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	// the length of the encrypted data should be more than length of nonce since encrypted data contains nonce
	if len(encryptedData) < gcm.NonceSize() {
		return nil, errors.Errorf("decrypting: length of encrypted data %d is not equal to length of nonce %d", len(encryptedData), gcm.NonceSize())
	}

	// encrypted data contains nonce + encrypted data, so split it
	nonce, encryptedData := encryptedData[:gcm.NonceSize()], encryptedData[gcm.NonceSize():]

	// decrypt the data, finally
	data, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return nil, err
	}

	return data, nil

}

// getTOTPValue generates a TOTP code for the given key
// How to generate a TOTP value: https://en.wikipedia.org/wiki/Time-based_One-time_Password_algorithm
// TODO: Use a better written library like https://github.com/pquerna/otp
func getTOTPValue(privateKey []byte, startUnixTime int64, endUnixTime int64, intervalSeconds int64) (string, error) {

	if endUnixTime <= startUnixTime {
		return "", errors.New("could not calculate counter time as end time is less than or equal to start time")
	}
	if intervalSeconds < 1 {
		return "", errors.New("could not calculate counter time as interval is less than 1 second")
	}

	counterTime := (endUnixTime - startUnixTime) / intervalSeconds
	// we will use counterTime as the 'message' in our hash function, however counterTime is an int
	// so we need to make it into []byte. We can use an ASCII representation of the int - but that
	// wouldn't work as the 'authenticating' service will probably use the  binary representation of
	// the number itself to genrate the code, and we need to match those codes
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, counterTime)
	if err != nil {
		return "", err
	}
	counterTimeInBytes := buf.Bytes()

	// Get the HMAC
	mac := hmac.New(sha512.New, privateKey)
	mac.Write(counterTimeInBytes)
	messageMAC := mac.Sum(nil)

	// HOTP is the truncated version of HMAC
	// Code from: https://github.com/pquerna/otp/blob/master/hotp/hotp.go
	// "Dynamic truncation" in RFC 4226
	// http://tools.ietf.org/html/rfc4226#section-5.4
	offset := messageMAC[len(messageMAC)-1] & 0xf
	value := int64(((int(messageMAC[offset]) & 0x7f) << 24) |
		((int(messageMAC[offset+1] & 0xff)) << 16) |
		((int(messageMAC[offset+2] & 0xff)) << 8) |
		(int(messageMAC[offset+3]) & 0xff))

	// TODO: Remove this hard-coding
	lenCode := gDefaultCodeLength
	mod := int32(value % int64(math.Pow10(lenCode)))

	// format the mod so it always has lenCode digits
	f := fmt.Sprintf("%%0%dd", lenCode)
	modF := fmt.Sprintf(f, mod)

	return modF, nil
}
