package api

import (
	"fmt"
	"net/http"
)

// Route represents a standard route object
type Route struct {
	Method      string
	Version     int
	Path        string
	HandlerFunc http.HandlerFunc
}

// GetPattern returns the url match pattern for the route
func (r Route) GetPattern() string {
	return fmt.Sprintf("/v%d/%s", r.Version, r.Path)
}
