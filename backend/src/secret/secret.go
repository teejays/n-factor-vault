package secret

import (
	"context"
	"fmt"

	"github.com/teejays/clog"
	"github.com/teejays/n-factor-vault/backend/src/orm"
	"github.com/teejays/n-factor-vault/backend/src/vault"
)

var gServiceName = "Secret Service" //LOL

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* O R M   M O D E L S
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// Secret stores the secrets of vaults
type Secret struct {
	orm.BaseModel `xorm:"extends"`
	VaultID       orm.ID `xorm:"notnull unique(secret)" json:"vault_id"`
	Secret        string `xorm:"notnull" json:"secret"`
}

// SecretRequest stores the requests users make to reveal vault secrets
type SecretRequest struct {
	orm.BaseModel `xorm:"extends"`
	UserID        orm.ID `xorm:"notnull" json:"user_id"`
	VaultID       orm.ID `xorm:"notnull" json:"vault_id"`
	Approved      bool   `xorm:"notnull default false" json:"approved"`
}

// SecretApproval stores the approvals for reveal requests
type SecretApproval struct {
	orm.BaseModel   `xorm:"extends"`
	SecretRequestID orm.ID `xorm:"notnull" json:"secret_request_id"`
	UserID          orm.ID `xorm:"notnull" json:"user_id"`
	Approved        bool   `xorm:"default null" json:"approved"`
}

func init() {
	err := orm.RegisterModel(&Secret{})
	if err != nil {
		clog.FatalErr(err)
	}

	err = orm.RegisterModel(&SecretRequest{})
	if err != nil {
		clog.FatalErr(err)
	}

	err = orm.RegisterModel(&SecretApproval{})
	if err != nil {
		clog.FatalErr(err)
	}
}

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* M E T H O D S
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// RequestParams are the parameters for a request to reveal a vault's secret
type RequestParams struct {
	VaultID orm.ID
	UserID  orm.ID
}

// UpdateParams are the parameters to update the reveal secret status (approve/reject)
type UpdateParams struct {
	SecretRequestID orm.ID
	UserID          orm.ID
	Approval        bool
}

// GetParams are the parameters to get the secret/secret status
type GetParams struct {
	SecretRequestID orm.ID
	UserID          orm.ID
}

// Status stores the information of the current approval status for the reveal secret request
type Status struct {
	SecretRequestID orm.ID
	Approved        bool
	Approvals       map[orm.ID]bool
}

// Request creates a request to reveal a vault's secret for the current authenticated user
func Request(ctx context.Context, req RequestParams) (*Status, error) {
	clog.Debugf("%s: creating a request to reveal secret of vault %s", gServiceName, req.VaultID)
	//Find all other users of the vault
	users, err := vault.GetVaultUsersByVaultID(ctx, req.VaultID)
	if err != nil {
		return nil, err
	}

	// Create a new request
	rr := SecretRequest{
		UserID:   req.UserID,
		VaultID:  req.VaultID,
		Approved: false,
	}
	rr.ID = orm.GetNewID()
	err = orm.InsertOne(&rr)
	if err != nil {
		return nil, err
	}

	var ras []SecretApproval

	// Create new approvals
	for _, user := range users {
		if user == nil {
			continue
		}
		ra := SecretApproval{
			SecretRequestID: rr.ID,
			UserID:          user.UserID,
			Approved:        false,
		}
		if user.UserID == req.UserID {
			ra.Approved = true
		}
		//TODO: fix this. Not sure why the BeforeInsert() is not triggering for InsertMulti
		ra.ID = orm.GetNewID()
		ras = append(ras, ra)
	}
	err = orm.InsertMulti(&ras)
	if err != nil {
		return nil, err
	}

	return GetStatus(ctx, GetParams{rr.ID, req.UserID})
}

// UpdateStatus updates the approval status of the specified request for the current authenticated user
func UpdateStatus(ctx context.Context, req UpdateParams) (*Status, error) {
	clog.Debugf("%s: updating the approval of secret of request %s", gServiceName, req.SecretRequestID)
	//Update the approval of the secret status of this user
	userID, err := orm.IDToStr(req.UserID)
	if err != nil {
		return nil, err
	}
	reqID, err := orm.IDToStr(req.SecretRequestID)
	if err != nil {
		return nil, err
	}

	requestConditions := map[string]string{
		"user_id":           userID,
		"secret_request_id": reqID,
	}

	//TODO: FindByColumn() and FindByColumns() both need to take in slices as the result interface
	// Might make sense to make FindOneByColumn() and FindOneByColumns()
	var sas []SecretApproval
	err = orm.FindByColumns(map[string]interface{}{
		"secret_request_id": req.SecretRequestID,
		"user_id":           req.UserID,
	}, &sas)
	if err != nil {
		return nil, err
	}

	if len(sas) != 1 {
		return nil, fmt.Errorf("%s: expected %d secret approvals but got %d", gServiceName, 1, len(sas))
	}
	sa := sas[0]
	sa.Approved = true

	err = orm.Update(requestConditions, sa)
	if err != nil {
		return nil, err
	}
	//Check if the overall request has been approved with this approval
	if req.Approval {
		s, err := GetStatus(ctx, GetParams{req.SecretRequestID, req.UserID})
		if err != nil {
			return nil, err
		}

		for _, approvals := range s.Approvals {
			if !approvals {
				return GetStatus(ctx, GetParams{req.SecretRequestID, req.UserID})
			}
		}
		var srs []SecretRequest
		err = orm.FindByColumn("id", req.SecretRequestID, &srs)
		if err != nil {
			return nil, err
		}

		if len(srs) != 1 {
			return nil, fmt.Errorf("%s: expected %d secret approvals but got %d", gServiceName, 1, len(srs))
		}
		sr := srs[0]
		sr.Approved = true

		err = orm.Update(map[string]string{"id": reqID}, sr)
		if err != nil {
			return nil, err
		}

	}

	return GetStatus(ctx, GetParams{req.SecretRequestID, req.UserID})
}

// GetStatus gets the current status of the given SecretRequest id
func GetStatus(ctx context.Context, req GetParams) (*Status, error) {
	clog.Debugf("%s: getting secret status of request %s", gServiceName, req.SecretRequestID)
	//TODO: confirm user has access to the request to retrieve status

	var s Status
	var srs []SecretRequest
	var sas []SecretApproval
	//Get the secret request
	err := orm.FindByColumn("id", req.SecretRequestID, &srs)
	if err != nil {
		return nil, err
	}
	if len(srs) != 1 {
		return &s, fmt.Errorf("%s: expected %d secret request but got %d", gServiceName, 1, len(srs))
	}
	s.SecretRequestID = srs[0].ID
	s.Approved = srs[0].Approved

	//Get the secret approvals
	err = orm.FindByColumn("secret_request_id", req.SecretRequestID, &sas)
	if err != nil {
		return nil, err
	}

	s.Approvals = make(map[orm.ID]bool)
	for _, sa := range sas {
		s.Approvals[sa.UserID] = sa.Approved
	}

	return &s, nil
}

// Get returns the secret for the specified vault
func Get(ctx context.Context, req GetParams) (*Secret, error) {
	clog.Debugf("%s: revealing secret of vault %s", gServiceName, req.SecretRequestID)

	//Check the approval status
	status, err := GetStatus(ctx, GetParams{req.SecretRequestID, req.UserID})
	if err != nil {
		return nil, err
	}
	if !status.Approved {
		return nil, fmt.Errorf("%s: secret %s not approved", gServiceName, req.SecretRequestID)
	}

	//Get the secret
	var srs []SecretRequest
	//Get the secret request
	err = orm.FindByColumn("id", req.SecretRequestID, &srs)
	if err != nil {
		return nil, err
	}
	if len(srs) != 1 {
		return nil, fmt.Errorf("%s: expected %d secret request but got %d", gServiceName, 1, len(srs))
	}

	var ss []Secret
	err = orm.FindByColumn("vault_id", srs[0].VaultID, &ss)
	if err != nil {
		return nil, err
	}
	if len(ss) != 1 {
		//TODO: figure out how secrets get created, for now return a dummy
		return &Secret{Secret: "Here is your Secret"}, nil
		return nil, fmt.Errorf("%s: expected %d secrets but got %d", gServiceName, 1, len(ss))
	}
	return &ss[0], err
}
