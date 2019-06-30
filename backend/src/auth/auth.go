package auth

import (
	"fmt"
	"time"

	"github.com/teejays/clog"
	jwt "github.com/teejays/go-jwt"
)

const sampleSecretKey = "I am a secret key"
const authExpiryDuration = 48 * time.Hour

// init initializes the JWT client
func init() {
	// Get the JWT client and create a token
	if jwt.IsClientInitialized() {
		return
	}

	err := jwt.InitClient(sampleSecretKey, authExpiryDuration)
	if err != nil {
		clog.Fatalf("Could not initialize the JWT Client: %v", err)
	}
}

// LoginCredentials represents user creds for logging in
type LoginCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse is the structure of how a successful login request repoonse will look like
type LoginResponse struct {
	JWT string
}

// Login authenticates user login credentials and returns an auth token if login is successful
func Login(creds LoginCredentials) (LoginResponse, error) {
	var resp LoginResponse

	// Get user by username
	// TODO: Call the user service to actually get the user object
	var user interface{}

	// Validate password
	err := validateCredentails(user, creds)
	if err != nil {
		return resp, err
	}

	// Generate the token
	token, err := generateToken(user)
	if err != nil {
		return resp, err
	}

	resp.JWT = token
	return resp, nil

}

// validateCredentails takes user credentials and validate the credentails
func validateCredentails(user interface{}, creds LoginCredentials) error {

	// TODO: We do not have the database or the user service setup yet, so let's just validate by default
	return nil

}

// generateToken creates and returns an authentication token for the user
func generateToken(user interface{}) (string, error) {

	// Get the JWT client and create a token
	cl, err := jwt.GetClient()
	if err != nil {
		return "", err
	}

	payloadData := map[string]interface{}{
		"UserID": "1",
		"Email":  "user@email.com",
	}

	payload := jwt.NewBasicPayload(payloadData)

	token, err := cl.CreateToken(payload)
	if err != nil {
		return "", fmt.Errorf("error creating JWT token: %v", err)
	}

	return token, err

}
