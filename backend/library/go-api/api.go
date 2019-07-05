package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/teejays/clog"
)

// StartServer initializes and runs the HTTP server
func StartServer(addr string, port int, routes []Route, authMiddlewareFunc MiddlewareFunc, preMiddlewareFuncs, postMiddlewareFuncs []MiddlewareFunc) error {

	m, err := GetHandler(routes, authMiddlewareFunc, preMiddlewareFuncs, postMiddlewareFuncs)
	if err != nil {
		return fmt.Errorf("could not setup the http handler: %c", err)
	}

	http.Handle("/", m)

	// Start the server
	clog.Infof("HTTP Server listenining on: %s:%d", addr, port)

	err = http.ListenAndServe(fmt.Sprintf("%s:%d", addr, port), nil)
	if err != nil {
		return fmt.Errorf("HTTP Server failed to start or continue running: %v", err)
	}

	return nil

}

// GetHandler constructs a HTTP handler with all the routes and midlleware funcs configured
func GetHandler(routes []Route, authMiddlewareFunc MiddlewareFunc, preMiddlewareFuncs, postMiddlewareFuncs []MiddlewareFunc) (http.Handler, error) {

	// Initiate a router
	m := mux.NewRouter()

	// Register routes to the handler
	// Set up pre handler middlewares
	for _, mw := range preMiddlewareFuncs {
		m.Use(mux.MiddlewareFunc(mw))
	}

	// Create an authenticated subrouter
	a := m.PathPrefix("").Subrouter()
	a.Use(mux.MiddlewareFunc(authMiddlewareFunc))

	// Range over routes and register them
	for _, route := range routes {
		// If the route is supposed to be authenticated, use auth mux
		r := m
		if route.Authenticate {
			r = a
		}
		// Register the route
		r.HandleFunc(route.GetPattern(), route.HandlerFunc).
			Methods(route.Method)
	}

	// Set up pre handler middlewares
	for _, mw := range postMiddlewareFuncs {
		m.Use(mux.MiddlewareFunc(mw))
	}

	return m, nil
}

// MiddlewareFunc can be inserted in a server for processing
type MiddlewareFunc mux.MiddlewareFunc

// LoggerMiddleware is a http.Handler middleware function that logs any request received
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log the request
		clog.Debugf("Server: HTTP request received for %s %s", r.Method, r.URL.Path)
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// SetJSONHeaderMiddleware sets the header for the response
func SetJSONHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set the header
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
