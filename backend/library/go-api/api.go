package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/teejays/clog"
)

// StartServer initializes and runs the HTTP server
func StartServer(addr string, port int, routes []Route, preMiddlewareFuncs, postMiddlewareFuncs []MiddlewareFunc) error {

	// Start the router
	m := mux.NewRouter()

	// Register routes to the handler
	// Set up pre handler middlewares
	for _, mw := range preMiddlewareFuncs {
		m.Use(mux.MiddlewareFunc(mw))
	}

	// Range over routes and set them up
	for _, r := range routes {
		m.HandleFunc(r.GetPattern(), r.HandlerFunc).
			Methods(r.Method)
	}

	// Set up pre handler middlewares
	for _, mw := range postMiddlewareFuncs {
		m.Use(mux.MiddlewareFunc(mw))
	}

	http.Handle("/", m)

	// Start the server
	clog.Infof("Listenining on: %s:%d", addr, port)

	return http.ListenAndServe(fmt.Sprintf("%s:%d", addr, port), nil)

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
