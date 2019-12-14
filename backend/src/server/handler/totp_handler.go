package handler

import (
	"net/http"

	"github.com/teejays/clog"

	api "github.com/teejays/gopi/mux"
	"github.com/teejays/n-factor-vault/backend/library/id"

	"github.com/teejays/n-factor-vault/backend/src/totp"
)

type CreateAccountRequest struct {
	Name       string
	PrivateKey string
}

// HandleCreateTOTPAccount creates a new vault for the authenticated user
func HandleCreateTOTPAccount(w http.ResponseWriter, r *http.Request) {

	var body CreateAccountRequest
	err := api.UnmarshalJSONFromRequest(r, &body)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false, nil)
		return
	}

	var req = totp.CreateAccountRequest{
		Name:       body.Name,
		PrivateKey: []byte(body.PrivateKey),
	}

	// Attempt login and get the token
	a, err := totp.CreateAccount(req)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false, nil)
		return
	}

	api.WriteResponse(w, http.StatusCreated, a.ID)

}

// HandleTOTPGetCode (GET) returns the vaults that the authenticated user is a part of
func HandleTOTPGetCode(w http.ResponseWriter, r *http.Request) {

	var req totp.GetCodeRequest

	// TODO: Verify that the requesting user has access to this TOTP account
	// AND the user has been approved by peers to access the code

	// Get the ID of the TOTP account for which we need the code
	accountID, err := api.GetMuxParamStr(r, "totp_account_id")
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false, nil)
		return
	}
	req.AccountID, err = id.StrToID(accountID)
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, err, false, nil)
		return
	}

	// Attempt login and get the token
	code, err := totp.GetCode(req)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, err, true, nil)
		return
	}
	clog.Debugf("%s: HandleGetCode(): returning:\n%+v", "HandleTOTPGetCode", code)
	api.WriteResponse(w, http.StatusOK, code)

}
