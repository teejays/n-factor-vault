package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	api "github.com/teejays/n-factor-vault/backend/library/go-api"
	"github.com/teejays/n-factor-vault/backend/src/auth"
	"github.com/teejays/n-factor-vault/backend/src/user"
)

// HandleLogin handles the login API requests
func HandleLogin(w http.ResponseWriter, r *http.Request) {

	// Read the HTTP request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false)
		return
	}
	defer r.Body.Close()

	// Unmarshal JSON into Go type
	var creds auth.LoginCredentials
	err = json.Unmarshal(body, &creds)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false)
		return
	}

	// Attempt login and get the token
	resp, err := auth.Login(creds)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false)
		return
	}

	api.WriteResponse(w, http.StatusOK, resp)

}

// HandleSignup handles the Signup API requests
func HandleSignup(w http.ResponseWriter, r *http.Request) {

	// Read the HTTP request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false)
		return
	}
	defer r.Body.Close()

	// Unmarshal JSON into Go type
	var req user.CreateUserRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false)
		return
	}

	// Attempt login and get the token
	u, err := user.CreateUser(req)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false)
		return
	}

	api.WriteResponse(w, http.StatusOK, u)

}
