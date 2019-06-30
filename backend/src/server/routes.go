package server

import (
	"net/http"

	api "github.com/teejays/n-factor-vault/backend/library/go-api"
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
		// Login Handler
		{
			Method:      http.MethodPost,
			Version:     ver1,
			Path:        "login",
			HandlerFunc: HandleLogin,
		},
	}

	return routes
}

// HandlePingRequest reponds with pong
func HandlePingRequest(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`Pong!`))
}
