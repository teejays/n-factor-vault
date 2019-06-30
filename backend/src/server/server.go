package server

import (
	api "github.com/teejays/n-factor-vault/backend/library/go-api"
	"github.com/teejays/n-factor-vault/backend/src/auth"
)

// StartServer initializes and startes the HTTP server
func StartServer(addr string, port int) error {

	// Get the Routes
	routes := GetRoutes()

	// Middlewares
	preMiddlewareFuncs := []api.MiddlewareFunc{api.MiddlewareFunc(api.LoggerMiddleware)}
	postMiddlewareFuncs := []api.MiddlewareFunc{api.SetJSONHeaderMiddleware}

	return api.StartServer(addr, port, routes, auth.AuthenticateRequestMiddleware, preMiddlewareFuncs, postMiddlewareFuncs)

}
