package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/teejays/clog"
	jwt "github.com/teejays/go-jwt"
	"github.com/teejays/n-factor-vault/backend/library/go-api"
)

const sampleSecretKey = "I am a secret key"
const authExpiryDuration = 48 * time.Hour

// Keys for storing auth information in http.Request context
type contextKey string

const ctxKeyToken = contextKey("jwt_token")
const ctxKeyUserID = contextKey("jwt_userid")
const ctxKeyIsAuthenticated = contextKey("is_authenticated")

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

// JWTClaim is the data that will be stored in the JWT token
type JWTClaim struct {
	UserID string `json:"uid"`
	jwt.BasicPayload
}

// generateToken creates and returns an authentication token for the user
func generateToken(user interface{}) (string, error) {

	// Get the JWT client and create a token
	cl, err := jwt.GetClient()
	if err != nil {
		return "", err
	}

	payloadData := JWTClaim{
		UserID: "1",
	}

	payload := jwt.NewBasicPayload(payloadData)

	token, err := cl.CreateToken(payload)
	if err != nil {
		return "", fmt.Errorf("error creating JWT token: %v", err)
	}

	return token, err

}

// AuthenticateRequestMiddleware should implement the authentication logic. It should should at the auth token
// and figure out what user context. Currently, this is not implemented and it only relies on
// and explicitly passed userID in the route.
func AuthenticateRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		clog.Debug("AuthenticateRequest() called...")

		// Get the token
		token, err := getBearerHeaderToken(r)
		if err != nil {
			api.WriteError(w, http.StatusUnauthorized, err, false)
			return
		}

		// Get the claim from the token (this verifies the token as well)
		claim, err := getJWTClaimFromToken(token)
		if err != nil {
			api.WriteError(w, http.StatusUnauthorized, err, false)
		}

		// Authentication succesful
		// Add the authentication payload to the context
		ctx := r.Context()
		ctx = context.WithValue(ctx, ctxKeyIsAuthenticated, true)
		ctx = context.WithValue(ctx, ctxKeyToken, token)
		ctx = context.WithValue(ctx, ctxKeyUserID, claim.UserID)

		// Add the updated context to http.Request
		r = r.WithContext(ctx)

		clog.Debug("Authentication process finished...")
		next.ServeHTTP(w, r)
	})
}

func getBearerHeaderToken(r *http.Request) (string, error) {

	// Get the authentication header
	val := r.Header.Get("Authorization")
	clog.Debugf("Authenticate Header: %v", val)
	// In JWT, we're looking for the Bearer type token
	// This means that the val should be like: Bearer <token>
	if strings.TrimSpace(val) == "" {
		return "", fmt.Errorf("Authorization header not found")
	}
	// - split by the space
	valParts := strings.Split(val, " ")
	if len(valParts) != 2 {
		return "", fmt.Errorf("Authorization header has an invalid form: it's not 'Authorization:Bearer TOKEN'")
	}
	if valParts[0] != "Bearer" {
		return "", fmt.Errorf("Authorization header has an invalid form: it's not `Authorization:Bearer TOKEN'")
	}

	return valParts[1], nil
}

// getJWTClaimFromToken authenticates that the token is valid.
func getJWTClaimFromToken(token string) (JWTClaim, error) {
	var claim JWTClaim

	// Get the JWT client and create a token
	cl, err := jwt.GetClient()
	if err != nil {
		return claim, err
	}

	err = cl.VerifyAndDecode(token, &claim)
	if err != nil {
		return claim, err
	}

	return claim, nil
}
