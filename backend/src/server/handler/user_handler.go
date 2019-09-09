package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/teejays/n-factor-vault/backend/library/go-api"

	"github.com/teejays/n-factor-vault/backend/src/auth"
	"github.com/teejays/n-factor-vault/backend/src/user"
)

// HandleSignup handles the Signup API requests
func HandleSignup(w http.ResponseWriter, r *http.Request) {

	var req user.CreateUserRequest
	err := api.UnmarshalJSONFromRequest(r, &req)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false, nil)
		return
	}

	// Attempt login and get the token
	u, err := user.CreateUser(req)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false, nil)
		return
	}

	api.WriteResponse(w, http.StatusCreated, u)

}

// HandleLogin handles the login API requests
func HandleLogin(w http.ResponseWriter, r *http.Request) {

	// Read the HTTP request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false, nil)
		return
	}
	defer r.Body.Close()

	if len(body) < 1 {
		api.WriteError(w, http.StatusBadRequest, api.ErrEmptyBody, false, nil)
		return
	}

	// Unmarshal JSON into Go type
	var creds auth.LoginCredentials
	err = json.Unmarshal(body, &creds)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, true, api.ErrInvalidJSON)
		return
	}

	// Attempt login and get the token
	resp, err := auth.Login(creds)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false, nil)
		return
	}

	api.WriteResponse(w, http.StatusOK, resp)

}
