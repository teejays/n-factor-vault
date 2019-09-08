package totp

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/go-playground/validator.v9"

	"github.com/teejays/clog"
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

	TOTP

	Anyone with access to the database should not be able to generate the TOTP code. TOTP Code is generated
	using a SecretKey provided by the external website. However, we shouldn't store this SecretKey directly
	in the database or anyone with access to database will be able to generate the TOTP code. We should therefore
	encrypt it before storing i.e. store EncryptedSecretKey.
		Setting up TOTP Vault process:
			SecretKey + *EncryptionKey* --> EncryptedSecretKey (DB)

	When a TOTP code is needed, we can decrypt EncryptedSecretKey into SecretKey and generate the token.
			EncryptedSecretKey (DB) + *EncryptionKey* --> SecretKey --> TOTP Code

	As you can see, this is a symmetric encryption. However, the process of decrypting should only be possible
	when certain criteria is met i.e. n out of m users associated with a vault approve it. Otherwise, there should be no
	way of decrypting the EncryptedSecretKey.

	This mean, *EncryptionKey*, cannot be stored in the database, or in the code, otherwise anyone with access to code/db
	can figure it out and hence get to the SecretKey and produce a TOTP Token. EncryptionKey needs to be generated on the
	fly!

	But how do we do this?
	Approach 1:
	Generate EncryptionKey by feeding a hash of the password of the team members, so we will need the team members
	to enter their password before the EncryptionKey could be generated. But how do we then make sure that we allow
	access a pre-determined subset of members have approved the access?

	If a team has m members in total, and minimum n members need to approve for access, we should generate mCn (m choose n)
	combinations of EncryptionKeys (based on passwords from all combinations of n-members), therefore store mCn EncryptedSecretKeys
	in the database. For example: if a Discord vault has 6 members in total (A,B,C,D,E,F),
	and we need at least 3 to approvals (1 requested + 2 more approvals), we will need to store EncryptedSecretKeys using encryption keys
	from all of the following combinations ABC, ABD, ABE, ABF, BCD, BCE, BCF, CDE, CDF and DEF, where ABC represent an encryption key
	genearted using some input (e.g. passwords) from user A, B and C.

	When a vault is first created, we will only have 1 member (i.e. the creator, let's say A). We will take his/her password,
	create a secure hash from it and use that hash to encrypt the SecretKey, as both n=m=1 in this case. But then user B is confirmed
	as a member, and now m=2 and n=2. Now, we can remove/decrypt the EncryptedSecretKey using A's EncryptionKey (generate the encryption key
	after asking A from an input prompt). Then, we will encrypt the SecretKey again using an EncryptionKey from both A and B
	(they'll enter input/password again). Now, in order to descrypt EncryptedSecretKey, both A and B will have to participate.

	Now, let's a third member is added, C but we still need to minimum of 2 approvals to grant access i.e. m=3, n=2.
	What we can do is decrypt the old EncryptedSecretKey from when m was 2, and encrypt it using EncryptionKeys with inputs from
	user AB, AC, BC. If anyone asks (let's say B) for access, we can send approval prompt to both  A and C, and if any one of them
	approve (i.e. provide their input password), we can generate one of the
	encryption keys (AB, or BC) and be able to decrypt the EncryptedSecretKey, and hence generate the TOTP code.

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

	err := orm.RegisterModels(&Account{})
	if err != nil {
		return err
	}

	return err
}

// Account represents one TOTP setup for a particular website/service
type Account struct {
	orm.BaseModel       `gorm:"EMBEDDED"`
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
	Name       string `validate:"required"`
	PrivateKey []byte `validate:"required,min=1"`
}

// CreateAccount creates a TOTP instance
func CreateAccount(req CreateAccountRequest) (Account, error) {
	var a Account

	// Validate the request
	err := validate.Struct(req)
	if err != nil {
		return a, err
	}

	a.Name = req.Name

	// Encrypt the private key
	// TODO: Make the encryption safer, but for POC this is fine
	// In order to encrypt, we will need an encryption key. Let's generate this encryption key
	// using a combination of a stored common key + service name + salt

	// Populate the TOTP instance
	key := getEncryptionKey(req.Name)
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
		return c, fmt.Errorf("decrypting with key: %v", err)
	}

	// generate a TOTP code
	now := time.Now().Unix()
	code, err := getTOTPValue(privateKey, a.StartUnixTime, now, a.IntervalSeconds)
	if err != nil {
		return c, fmt.Errorf("generating TOTP code: %v", err)
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
func getEncryptionKey(accountName string) []byte {
	h := sha256.New()
	h.Write([]byte(accountName))
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

// decryptWithKey takes a key and encrypted data, and decrypts it
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

	counter := int64(math.Floor(float64(endUnixTime-startUnixTime) / float64(intervalSeconds)))
	// we will use counter as the 'message' in our hash function, however counter is an int
	// so we need to make it into []byte. We can use an ASCII representation of the int - but that
	// wouldn't work as the 'authenticating' service will probably use the  binary representation of
	// the number itself to genrate the code, and we need to match those codes
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(counter))

	// Private Key should be of UPPER CASE
	privateKey = bytes.ToUpper(privateKey)

	clog.Debugf("%s: getting code using private key: '%s'", "totp", privateKey)
	secretBytes, err := base32.StdEncoding.DecodeString(string(privateKey))
	if err != nil {
		return "", fmt.Errorf("encoding private key to base64: %v", err)
	}

	// Get the HMAC
	clog.Debugf("%s: getting code using secret: '%s'", "totp", secretBytes)
	mac := hmac.New(sha1.New, secretBytes)
	clog.Debugf("%s: getting code of counter: '%d'", "totp", counter)
	mac.Write(buf)
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
