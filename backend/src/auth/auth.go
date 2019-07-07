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
	"github.com/teejays/n-factor-vault/backend/src/orm"
	"github.com/teejays/n-factor-vault/backend/src/user"
)

const sampleSecretKey = "I am a secret key"
const authExpiryDuration = 48 * time.Hour

// Keys for storing auth information in http.Request context
type contextKey string

const gCtxKeyToken = contextKey("jwt_token")
const gCtxKeyUserID = contextKey("jwt_userid")
const gCtxKeyIsAuthenticated = contextKey("is_authenticated")

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
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse is the structure of how a successful login request repoonse will look like
type LoginResponse struct {
	JWT string
}

// ErrNotAuthenticated is returned when a request or context is not authenticated
var ErrNotAuthenticated = fmt.Errorf("not authenticated")

// ErrInvalidCredentails means that login credentials are invalid
var ErrInvalidCredentails = fmt.Errorf("login credentials are invalid")

// Login authenticates user login credentials and returns an auth token if login is successful
func Login(creds LoginCredentials) (LoginResponse, error) {
	var resp LoginResponse

	if strings.TrimSpace(creds.Email) == "" {
		return resp, fmt.Errorf("no email provided")
	}
	if strings.TrimSpace(creds.Password) == "" {
		return resp, fmt.Errorf("no password provided")
	}

	// Get user by email
	u, err := user.GetSecureUserByEmail(creds.Email)
	if err != nil {
		return resp, err
	}
	if u == nil {
		clog.Warnf("auth: no user found with email %s", creds.Email)
		return resp, ErrInvalidCredentails
	}

	// Validate password
	isValid := u.SecurePassword.ValidatePassword(creds.Password)
	if !isValid {
		return resp, ErrInvalidCredentails
	}

	// Generate the token
	token, err := generateToken(u.User)
	if err != nil {
		return resp, err
	}

	resp.JWT = token
	return resp, nil

}

// JWTClaim is the data that will be stored in the JWT token
type JWTClaim struct {
	jwt.BaseClaim
	UserID orm.ID `json:"uid"`
}

// generateToken creates and returns an authentication token for the user
func generateToken(u user.User) (string, error) {

	// Get the JWT client and create a token
	cl, err := jwt.GetClient()
	if err != nil {
		return "", err
	}

	claim := JWTClaim{UserID: u.ID}

	token, err := cl.CreateToken(&claim)
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
			api.WriteError(w, http.StatusUnauthorized, err, false, nil)
			return
		}

		// Get the claim from the token (this verifies the token as well)
		claim, err := getJWTClaimFromToken(token)
		if err != nil {
			api.WriteError(w, http.StatusUnauthorized, err, false, nil)
			return
		}

		if claim.UserID.IsEmpty() {
			clog.Errorf("auth: middleware: got an empty userID from a verified jwt token")
			api.WriteError(w, http.StatusInternalServerError, fmt.Errorf(api.ErrMessageClean), false, nil)
			return
		}
		// Authentication succesful
		// Add the authentication payload to the context
		ctx := r.Context()
		ctx = context.WithValue(ctx, gCtxKeyIsAuthenticated, true)
		ctx = context.WithValue(ctx, gCtxKeyToken, token)
		ctx = context.WithValue(ctx, gCtxKeyUserID, claim.UserID)

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

	clog.Debugf("auth: verified claim from token:\n%+v", claim)

	return claim, nil
}

// GetUserFromContext returns the user instance of the authenticated user using the information in the context
func GetUserFromContext(ctx context.Context) (*user.User, error) {
	if !IsContextAuthenticated(ctx) {
		return nil, ErrNotAuthenticated
	}

	clog.Debug("auth: Getting User from Context")

	// Check the value of userID in context
	v := ctx.Value(gCtxKeyUserID)
	if v == nil {
		return nil, ErrNotAuthenticated
	}
	userID, ok := v.(orm.ID)
	if !ok {
		clog.Errorf("auth: gCtxKeyUserID value in context cannot be converted to orm.ID: %v", v)
		return nil, ErrNotAuthenticated
	}
	if userID == "" {
		return nil, ErrNotAuthenticated
	}

	// Get the User
	u, err := user.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	return u, nil

}

// IsContextAuthenticated takes a context and returns true if it is authenticated
func IsContextAuthenticated(ctx context.Context) bool {

	clog.Debug("auth: IsContextAuthenticated")

	// Not authenticated if the request itself is nil
	if ctx == nil {
		return false
	}

	// Check the value isAuthenticated in context
	clog.Debug("auth: IsContextAuthenticated: getting 'isAuthenticated' value from context")
	v1 := ctx.Value(gCtxKeyIsAuthenticated)
	if v1 == nil {
		return false
	}
	isAuthenticated, ok := v1.(bool)
	if !ok {
		clog.Errorf("auth: gCtxKeyIsAuthenticated value in context cannot be converted to bool: %v", v1)
		return false
	}
	if !isAuthenticated {
		return false
	}

	// Check the value of userID in context
	clog.Debug("auth: IsContextAuthenticated: getting 'userID' value from context")
	v2 := ctx.Value(gCtxKeyUserID)
	if v2 == nil {
		clog.Warn("auth: IsContextAuthenticated: gCtxKeyUserID is nil in context")
		return false
	}
	userID, ok := v2.(orm.ID)
	if !ok {
		clog.Errorf("auth: gCtxKeyUserID value in context cannot be converted to orm.ID: %v", v2)
		return false
	}
	if userID == "" {
		clog.Warnf("auth: IsContextAuthenticated: gCtxKeyUserID is empty in context, %v", v2)
		return false
	}

	// Make sure we have a JWT token
	clog.Debug("auth: IsContextAuthenticated: getting 'token' value from context")
	v3 := ctx.Value(gCtxKeyToken)
	if v2 == nil {
		return false
	}
	token, ok := v3.(string)
	if !ok {
		clog.Errorf("auth: gCtxKeyToken value in context cannot be converted to string: %v", v3)
		return false
	}
	if token == "" {
		return false
	}

	return true

}
