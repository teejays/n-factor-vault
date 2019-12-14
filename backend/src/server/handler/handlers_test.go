package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/teejays/clog"

	api "github.com/teejays/gopi/mux"
	apitest "github.com/teejays/gopi/mux/muxtest"
	"github.com/teejays/n-factor-vault/backend/library/env"
	"github.com/teejays/n-factor-vault/backend/library/orm"

	"github.com/teejays/n-factor-vault/backend/src/auth"
	"github.com/teejays/n-factor-vault/backend/src/server/handler"
	"github.com/teejays/n-factor-vault/backend/src/user"
	"github.com/teejays/n-factor-vault/backend/src/vault"
)

func init() {
	err := initError()
	if err != nil {
		clog.FatalErr(err)
	}
}

func initError() error {
	var err error

	// Set the log level
	clog.Infof("Application Environment: %s", env.GetAppEnv())
	clog.LogLevel = 8

	// Initialize the ORM package
	err = orm.Init()
	if err != nil {
		return err
	}

	// Initialize the services: the order should be important ideally, so dependent services are initialized later
	err = user.Init()
	if err != nil {
		return err
	}

	err = vault.Init()
	if err != nil {
		return err
	}

	/* Uncomment when we tests for secrets service
	err = secret.Init()
	if err != nil {
		return err
	}
	*/

	return nil
}

func helperCreateTestUsersT(t *testing.T) {
	err := helperCreateUsers("Jon", "Jane")
	if err != nil {
		t.Error(err)
	}
}

func helperLoginTestUsersT(t *testing.T) (string, string) {
	token1, err := helperLoginUser("Jon")
	if err != nil {
		t.Error(err)
	}
	token2, err := helperLoginUser("Jane")
	if err != nil {
		t.Error(err)
	}
	return token1, token2
}

func helperCreateTestVaultsT(t *testing.T, token string) {
	err := helperCreateVaults(token, "Facebook", "Twitter")
	if err != nil {
		t.Errorf("could not create a test vaults: %v", err)
	}
}

var mockUsers = map[string]string{
	"Jon":  `{"name":"Jon Doe", "email":"jon.doe@email.com","password":"jons_secret"}`,
	"Jane": `{"name":"Jane Does", "email":"jane.does@email.com","password":"janes_secret"}`,
}

var mockVaults = map[string]string{
	"Facebook": `{"name":"Facebook", "description":"Shared account for our org"}`,
	"Twitter":  `{"name":"Twitter", "description":"Shared account for friends"}`,
}

func helperCreateUsers(names ...string) error {
	// Define the Handler Request to signup a user
	p := apitest.HandlerReqParams{
		Route:       "/v1/signup",
		Method:      http.MethodPost,
		HandlerFunc: handler.HandleSignup,
	}

	// Loop over the users and create them
	for _, u := range names {
		if _, _, err := p.MakeHandlerRequest(
			mockUsers[u], // this is the HTTP request content
			[]int{http.StatusCreated, http.StatusOK},
		); err != nil {
			return err
		}
	}

	return nil

}

func helperLoginUser(name string) (string, error) {
	// Define the Handler Request to signup a user
	p := apitest.HandlerReqParams{
		Route:       "/v1/login",
		Method:      http.MethodPost,
		HandlerFunc: handler.HandleLogin,
	}

	// Make the Login request
	_, body, err := p.MakeHandlerRequest(mockUsers[name], []int{http.StatusOK})
	if err != nil {
		return "", err
	}

	// Get JWT Token from response
	var m = make(map[string]interface{})
	if err := json.Unmarshal(body, &m); err != nil {
		return "", err
	}
	if _, ok := m["jwt"]; !ok {
		return "", fmt.Errorf("couldn't get JWT token in response")
	}
	token, ok := m["jwt"].(string)
	if !ok {
		return "", fmt.Errorf("JWT token is not of type string")
	}
	if token == "" {
		return "", fmt.Errorf("JWT token is empty")
	}

	return token, nil

}

func helperCreateVaults(token string, vaults ...string) error {
	// Create a test handler request (this needs to be authenticated)
	p := apitest.HandlerReqParams{
		Route:           "/v1/vaults",
		Method:          http.MethodPost,
		HandlerFunc:     handler.HandleCreateVault,
		AuthBearerToken: token,
		Middlewares:     []api.MiddlewareFunc{auth.AuthenticateRequestMiddleware},
	}

	// Create Vaults
	for _, v := range vaults {
		_, _, err := p.MakeHandlerRequest(mockVaults[v], []int{http.StatusCreated})
		if err != nil {
			return err
		}
	}

	return nil
}
