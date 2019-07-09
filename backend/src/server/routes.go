package server

import (
	"net/http"

	api "github.com/teejays/n-factor-vault/backend/library/go-api"
	userhandler "github.com/teejays/n-factor-vault/backend/src/user/handler"
	vaulthandler "github.com/teejays/n-factor-vault/backend/src/vault/handler"
)

const (
	ver1 = 1
	ver2 = 1
)

// GetRoutes returns all the routes for this service
func GetRoutes() []api.Route {

	routes := []api.Route{
		// Ping Handler
		{
			Method:      http.MethodGet,
			Version:     ver1,
			Path:        "ping",
			HandlerFunc: HandlePingRequest,
		},
		// Ping Handler (Authenticated)
		{
			Method:       http.MethodGet,
			Version:      ver1,
			Path:         "secure/ping",
			HandlerFunc:  HandlePingRequest,
			Authenticate: true,
		},
		// Signup Handler
		{
			Method:      http.MethodPost,
			Version:     ver1,
			Path:        "signup",
			HandlerFunc: userhandler.HandleSignup,
		},
		// Login Handler
		{
			Method:      http.MethodPost,
			Version:     ver1,
			Path:        "login",
			HandlerFunc: userhandler.HandleLogin,
		},
		// Vault Create Handler
		{
			Method:       http.MethodPost,
			Version:      ver1,
			Path:         "vault",
			HandlerFunc:  vaulthandler.HandleCreateVault,
			Authenticate: true,
		},
		// Vault Get Vaults For User
		{
			Method:       http.MethodGet,
			Version:      ver1,
			Path:         "vault",
			HandlerFunc:  vaulthandler.HandleGetVaults,
			Authenticate: true,
		},
	}

	return routes
}

// HandlePingRequest reponds with pong
func HandlePingRequest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`Pong!`))
}
