package totp

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teejays/clog"
	"github.com/teejays/n-factor-vault/backend/library/id"
	"github.com/teejays/n-factor-vault/backend/library/orm"
)

func init() {

	err := orm.Init()
	if err != nil {
		clog.FatalErr(err)
	}

	err = Init()
	if err != nil {
		clog.FatalErr(err)
	}

}

func TestCreateAccount(t *testing.T) {
	// Make sure that we empty any table that these tests might populate once the test is over
	orm.EmptyTestTables(t, &Account{})
	defer orm.EmptyTestTables(t, &Account{})

	tests := []struct {
		name    string
		req     CreateAccountRequest
		want    Account
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "error if empty account name",
			req: CreateAccountRequest{
				Name:       "",
				PrivateKey: []byte("some secret key"),
			},
			wantErr: true,
		},
		{
			name: "error if empty private key",
			req: CreateAccountRequest{
				Name:       "Facebook",
				PrivateKey: []byte(""),
			},
			wantErr: true,
		},
		{
			name: "error if nil private key",
			req: CreateAccountRequest{
				Name:       "Facebook",
				PrivateKey: nil,
			},
			wantErr: true,
		},
		{
			name: "success if good request",
			req: CreateAccountRequest{
				Name:       "Facebook",
				PrivateKey: []byte("ORUGKIDQOJUXMYLUMUQGWZLZ"), // base32 for "the private key"
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := CreateAccount(tt.req)
			if tt.wantErr {
				fmt.Println(err)
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.req.Name, got.Name)
			assert.Equal(t, int64(30), got.IntervalSeconds)
			assert.Equal(t, int64(0), got.StartUnixTime)
			assert.NotEqual(t, 0, len(got.EncryptedPrivateKey))
		})
	}
}

func TestEncryptDecrypt(t *testing.T) {
	type args struct {
		key     []byte
		message []byte
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "successful descryption for normal key and message",
			args: args{
				key:     []byte("a secret key use to encrypt info"),
				message: []byte("this is a private key for an account used to generate TOTP codes"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			encrypted, err := encryptWithKey(tt.args.key, tt.args.message)
			assert.NoError(t, err)

			decrypted, err := decryptWithKey(tt.args.key, encrypted)
			assert.NoError(t, err)

			assert.Equal(t, tt.args.message, decrypted)

		})
	}
}

func TestGetCode(t *testing.T) {

	orm.EmptyTestTables(t, &Account{})
	defer orm.EmptyTestTables(t, &Account{})
	// Create a new TOTP account so we testgetting it's code
	a := createTestAccount(t)

	type args struct {
		req GetCodeRequest
	}
	tests := []struct {
		name    string
		args    args
		want    Code
		wantErr bool
	}{
		{
			name: "successfully get a code for a valid account",
			args: args{
				req: GetCodeRequest{AccountID: a.ID},
			},
		},
		{
			name: "error if accountID is invalid",
			args: args{
				req: GetCodeRequest{AccountID: id.GetNewID()},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now()
			got, err := GetCode(tt.args.req)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			fmt.Printf("%+v", got)
			assert.NotEmpty(t, got.Code)
			assert.Len(t, got.Code, gDefaultCodeLength)
			assert.True(t, got.ValidAt.Before(time.Now()))
			assert.True(t, got.ExpireAt.After(now))
			assert.True(t, got.ExpireAt.Sub(got.ValidAt) == time.Duration(gDefaultIntervalInSeconds)*time.Second)
		})
	}
}

func createTestAccount(t *testing.T) Account {
	req := CreateAccountRequest{
		Name:       "Facebook",
		PrivateKey: []byte("ORUGKIDQOJUXMYLUMUQGWZLZ"), // base32 for "the private key"
	}
	a, err := CreateAccount(req)
	if err != nil {
		t.Fatal(err)
	}
	clog.Debugf("created a test TOTP account with ID %s", a.ID)
	return a
}
