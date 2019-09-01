package password

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSecurePassword(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name    string
		config  Config
		args    args
		wantErr bool
	}{
		{
			"a simple password should produce a SecurePassword",
			defaultConfig,
			args{"password123"},
			false,
		},
		{
			"empty password should error",
			defaultConfig,
			args{""},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSecurePassword(tt.args.password)
			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}
			assert.Equal(t, tt.config.numIterations, got.IterationCount)
			assert.Equal(t, tt.config.saltSize, len(got.Salt))
			assert.Equal(t, tt.config.hashLength, len(got.Hash))
		})
	}
}

func TestValidatePassword(t *testing.T) {

	type args struct {
		password string
	}
	tests := []struct {
		name             string
		config           Config
		originalPassword string
		args             args
		want             bool
	}{
		{
			"return true if password is the same - 1",
			defaultConfig,
			"password123",
			args{"password123"},
			true,
		},
		{
			"return true if password is the same - 2",
			defaultConfig,
			"a1b2c3d4a1b2c3d4",
			args{"a1b2c3d4a1b2c3d4"},
			true,
		},
		{
			"return false if password is not the same",
			defaultConfig,
			"password123",
			args{"password1234"},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Get the SecurePassword
			sp, err := NewSecurePassword(tt.originalPassword)
			assert.NoError(t, err)

			if got := ValidatePassword(sp, tt.args.password); got != tt.want {
				t.Errorf("SecurePassword.ValidatePassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_areEqual(t *testing.T) {
	hash0 := []byte{}
	hash1 := []byte{97, 92, 43, 0, 1, 4}
	hash2 := []byte{97, 92, 43, 0, 1, 4}
	hash3 := []byte{51, 83, 42, 7, 31, 89}
	hash4 := []byte{97, 92, 43, 0, 1}

	type args struct {
		h1 []byte
		h2 []byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"return true if two slices are the same",
			args{hash1, hash2},
			true,
		},
		{
			"return false if one slice is smaller than other",
			args{hash1, hash4},
			false,
		},
		{
			"return false if two slices are different",
			args{hash1, hash3},
			false,
		},
		{
			"return false if one slice is empty",
			args{hash1, hash0},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := areEqual(tt.args.h1, tt.args.h2); got != tt.want {
				t.Errorf("areEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}
