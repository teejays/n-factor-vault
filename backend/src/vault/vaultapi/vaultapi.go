package vaultapi

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/teejays/clog"
	"github.com/teejays/n-factor-vault/backend/library/go-api"
	"github.com/teejays/n-factor-vault/backend/src/auth"
	"github.com/teejays/n-factor-vault/backend/src/vault"
)

func init() {
	clog.LogLevel = 8
}

// HandleCreateVault creates a new vault for the authenticated user
func HandleCreateVault(w http.ResponseWriter, r *http.Request) {

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
	var req vault.CreateVaultRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, true, api.ErrInvalidJSON)
		return
	}

	// Populate the AdminUserID field of req using the authneticated userID
	u, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err, true, nil)
		return
	}
	req.AdminUserID = u.ID

	// Attempt login and get the token
	v, err := vault.CreateVault(r.Context(), req)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false, nil)
		return
	}

	api.WriteResponse(w, http.StatusCreated, v)

}

// HandleGetVaults (GET) returns the vaults that the authenticated user is a part of
func HandleGetVaults(w http.ResponseWriter, r *http.Request) {

	// Populate the AdminUserID field of req using the authneticated userID
	u, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err, true, nil)
		return
	}

	// Attempt login and get the token
	vaults, err := vault.GetVaultsByUser(r.Context(), u.ID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err, true, nil)
		return
	}

	api.WriteResponse(w, http.StatusOK, vaults)

}
