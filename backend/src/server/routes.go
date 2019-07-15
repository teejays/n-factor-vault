package server

import (
	"net/http"

	api "github.com/teejays/n-factor-vault/backend/library/go-api"
	"github.com/teejays/n-factor-vault/backend/src/server/handler"
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
			HandlerFunc: handler.HandleSignup,
		},
		// Login Handler
		{
			Method:      http.MethodPost,
			Version:     ver1,
			Path:        "login",
			HandlerFunc: handler.HandleLogin,
		},
		// Vault Create Handler
		{
			Method:       http.MethodPost,
			Version:      ver1,
			Path:         "vault",
			HandlerFunc:  handler.HandleCreateVault,
			Authenticate: true,
		},
		// Vault Get Vaults For User
		{
			Method:       http.MethodGet,
			Version:      ver1,
			Path:         "vaults",
			HandlerFunc:  handler.HandleGetVaults,
			Authenticate: true,
		},
		{
			Method:       http.MethodPost,
			Version:      ver1,
			Path:         "vault/{vault_id}/user",
			HandlerFunc:  handler.HandleAddVaultUser,
			Authenticate: true,
		},
		{
			Method:       http.MethodPost,
			Version:      ver1,
			Path:         "vault/{vault_id}/secret",
			HandlerFunc:  handler.HandleRequestSecret,
			Authenticate: true,
		},
		{
			Method:       http.MethodPatch,
			Version:      ver1,
			Path:         "vault/secret/{secret_request_id}",
			HandlerFunc:  handler.HandleUpdateSecretStatus,
			Authenticate: true,
		},
		{
			Method:       http.MethodGet,
			Version:      ver1,
			Path:         "vault/secret/{secret_request_id}/status",
			HandlerFunc:  handler.HandleGetSecretStatus,
			Authenticate: true,
		},
		{
			Method:       http.MethodGet,
			Version:      ver1,
			Path:         "vault/secret/{secret_request_id}",
			HandlerFunc:  handler.HandleGetSecret,
			Authenticate: true,
		},
	}

	return routes
}

// HandlePingRequest responds with pong
func HandlePingRequest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`Pong!`))
}
