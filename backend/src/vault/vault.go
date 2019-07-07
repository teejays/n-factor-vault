// Package vault implements the main entity of this app - a vault. Any particular site/topic
// for which secure information is stored is called a vault. A vault is owned by one or more user
// and can have it's opening security customizable.
package vault

import (
	"context"

	"github.com/teejays/clog"
	"github.com/teejays/n-factor-vault/backend/src/auth"
	"github.com/teejays/n-factor-vault/backend/src/orm"
	"github.com/teejays/n-factor-vault/backend/src/user"
)

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* O R M   M O D E L S
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// Vault definers the properties and fields of a vault enity
// TODO: a vault should have hasMany relation with user.User
// DECISION: do we want to explicitly have hasMany relations across tables?
type Vault struct {
	orm.BaseModel `xorm:"extends"`
	Name          string `xorm:"notnull unique(name_adminuser)" json:"name"`
	Description   string `xorm:"notnull" json:"description"`
	AdminUserID   orm.ID `xorm:"notnull unique(name_adminuser)" json:"admin_user_id"`

	Users []*user.User `xorm:"-" json:"users"`
}

// vaultUser represents the mapping between vault and users that are a part of it. This is not exported
// since we want to add this data to the main Vault struct while returning a Vault type, and don't to expose
// by itself.
type vaultUser struct {
	orm.BaseModel `xorm:"extends"`
	VaultID       orm.ID `xorm:"'vault_id' notnull unique(vault_user)" json:"vault_id"`
	UserID        orm.ID `xorm:"'user_id' notnull unique(vault_user)" json:"user_id"`
}

func init() {
	err := orm.RegisterModel(&Vault{})
	if err != nil {
		clog.FatalErr(err)
	}

	err = orm.RegisterModel(&vaultUser{})
	if err != nil {
		clog.FatalErr(err)
	}
}

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* M E T H O D S
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// CreateVaultRequest are the parameters that are passed when creating a vault
type CreateVaultRequest struct {
	Name        string
	Description string
}

// CreateVault creates a new vault with the current authenticated user as the admin
func CreateVault(ctx context.Context, req CreateVaultRequest) (*Vault, error) {
	clog.Debugf("vault: creating vault %s", req.Name)
	var err error

	// Create a vault instance
	v := Vault{
		Name:        req.Name,
		Description: req.Description,
	}

	// In this case get and assign the ID now so we can use it the vault-user entities
	// Assigning it explicitly means that the ORm library doesn't assign it itself during insert
	v.ID = orm.GetNewID()

	// Get the ID of the user creating this vault, so we have the userID
	u, err := auth.GetUserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	v.AdminUserID = u.ID

	// Set the vault-user
	vu := []*vaultUser{
		&vaultUser{
			VaultID: v.ID,
			UserID:  u.ID,
		},
	}

	// Insert the vault and the new vault-user mapping
	err = orm.InsertTx(&v, vu[0])
	if err != nil {
		return nil, err
	}

	// v.Users field is ignored by ORM in any case, and we only need to populate it as part of the Vault instance in Go
	v.Users = []*user.User{u}

	return &v, nil
}
