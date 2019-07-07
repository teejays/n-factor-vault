package server

import (
	"net/http"

	api "github.com/teejays/n-factor-vault/backend/library/go-api"
	"github.com/teejays/n-factor-vault/backend/src/user/userapi"
	"github.com/teejays/n-factor-vault/backend/src/vault/vaultapi"
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
			HandlerFunc: userapi.HandleSignup,
		},
		// Login Handler
		{
			Method:      http.MethodPost,
			Version:     ver1,
			Path:        "login",
			HandlerFunc: userapi.HandleLogin,
		},
		// Vault Create Handler
		{
			Method:       http.MethodPost,
			Version:      ver1,
			Path:         "vault",
			HandlerFunc:  vaultapi.HandleCreateVault,
			Authenticate: true,
		},
	}

	return routes
}

// HandlePingRequest reponds with pong
func HandlePingRequest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`Pong!`))
}
