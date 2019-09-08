package handler

import (
	"fmt"
	"net/http"

	"github.com/teejays/clog"

	"github.com/teejays/n-factor-vault/backend/library/go-api"
	"github.com/teejays/n-factor-vault/backend/library/id"

	"github.com/teejays/n-factor-vault/backend/src/auth"
	"github.com/teejays/n-factor-vault/backend/src/vault"
)

func init() {

}

// HandleCreateVault creates a new vault for the authenticated user
func HandleCreateVault(w http.ResponseWriter, r *http.Request) {

	var req vault.CreateVaultRequest
	err := api.UnmarshalJSONFromRequest(r, &req)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false, nil)
		return
	}

	// Populate the UserID field of req using the authenticated userID
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

	// Populate the UserID field of req using the authenticated userID
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
	clog.Debugf("%s: HandleGetVaults(): returning:\n%+v", "Vault Handler", vaults)
	api.WriteResponse(w, http.StatusOK, vaults)

}

// HandleAddVaultUser is the HTTP handler for adding a new user to a vault
func HandleAddVaultUser(w http.ResponseWriter, r *http.Request) {

	// In the HTTP request body, we only expect the userID of the user
	// to be added. The vaultID of the vault will be in the URL

	// Get the content of the request
	var req vault.AddUserToVaultRequest
	err := api.UnmarshalJSONFromRequest(r, &req)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false, nil)
		return
	}

	if req.UserID.IsEmpty() {
		api.WriteError(w, http.StatusBadRequest, fmt.Errorf("empty user_id"), false, nil)
		return
	}

	// Get the vaultID from URL params
	vaultID, err := api.GetMuxParamStr(r, "vault_id")
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false, nil)
		return
	}
	req.VaultID, err = id.StrToID(vaultID)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false, nil)
		return
	}

	// Call the vault package function to add user to the vault
	v, err := vault.AddUserToVault(r.Context(), req)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err, true, nil)
		return
	}

	api.WriteResponse(w, http.StatusOK, v)

}
