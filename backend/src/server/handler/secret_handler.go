package handler

import (
	"net/http"

	api "github.com/teejays/gopi/mux"
	"github.com/teejays/n-factor-vault/backend/library/id"

	"github.com/teejays/n-factor-vault/backend/src/auth"
	"github.com/teejays/n-factor-vault/backend/src/secret"
)

// HandleRequestSecret handles request to reveal a secret
func HandleRequestSecret(w http.ResponseWriter, r *http.Request) {

	var req secret.RequestParams
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

	// Populate the UserID field of req using the authenticated userID
	u, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err, true, nil)
		return
	}
	req.UserID = u.ID

	// Send the secret request and get the status
	s, err := secret.Request(r.Context(), req)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false, nil)
		return
	}

	api.WriteResponse(w, http.StatusCreated, s)
}

// HandleUpdateSecretStatus handles request to update a secret approval
func HandleUpdateSecretStatus(w http.ResponseWriter, r *http.Request) {

	var req secret.UpdateParams
	err := api.UnmarshalJSONFromRequest(r, &req)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false, nil)
		return
	}
	// Get the secretRequestID from URL params
	secretRequestIDStr, err := api.GetMuxParamStr(r, "secret_request_id")
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false, nil)
		return
	}
	req.SecretRequestID, err = id.StrToID(secretRequestIDStr)
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
	req.UserID = u.ID

	// Update the secret status
	s, err := secret.UpdateStatus(r.Context(), req)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false, nil)
		return
	}

	api.WriteResponse(w, http.StatusCreated, s)
}

// HandleGetSecretStatus handles request to get secret status
func HandleGetSecretStatus(w http.ResponseWriter, r *http.Request) {

	// Get the secretRequestID from URL params
	secretRequestIDStr, err := api.GetMuxParamStr(r, "secret_request_id")
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false, nil)
		return
	}
	secretRequestID, err := id.StrToID(secretRequestIDStr)
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

	// Get status
	vaults, err := secret.GetStatus(r.Context(), secret.GetParams{secretRequestID, u.ID})
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err, true, nil)
		return
	}
	api.WriteResponse(w, http.StatusOK, vaults)
}

// HandleGetSecret handles request to get secret
func HandleGetSecret(w http.ResponseWriter, r *http.Request) {

	// Get the secretRequestID from URL params
	secretRequestIDStr, err := api.GetMuxParamStr(r, "secret_request_id")
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false, nil)
		return
	}
	secretRequestID, err := id.StrToID(secretRequestIDStr)
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

	// Get status
	vaults, err := secret.Get(r.Context(), secret.GetParams{secretRequestID, u.ID})
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err, false, nil)
		return
	}
	api.WriteResponse(w, http.StatusOK, vaults)
}
