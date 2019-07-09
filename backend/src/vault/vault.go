// Package vault implements the main entity of this app - a vault. Any particular site/topic
// for which secure information is stored is called a vault. A vault is owned by one or more user
// and can have it's opening security customizable.
package vault

import (
	"context"
	"fmt"
	"strings"

	"github.com/teejays/clog"
	"github.com/teejays/n-factor-vault/backend/src/orm"
	"github.com/teejays/n-factor-vault/backend/src/user"
)

var gServiceName = "Vault Service"

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
	AdminUserID orm.ID
	Name        string
	Description string
}

// CreateVault creates a new vault with the current authenticated user as the admin
func CreateVault(ctx context.Context, req CreateVaultRequest) (*Vault, error) {
	clog.Debugf("vault: creating vault %s", req.Name)
	var err error

	// TODO: Validate the request
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("name is empty")
	}
	if strings.TrimSpace(req.Description) == "" {
		return nil, fmt.Errorf("description is empty")
	}
	if req.AdminUserID.IsEmpty() {
		return nil, fmt.Errorf("admin userID is empty")
	}

	// Create a vault instance
	v := Vault{
		Name:        req.Name,
		Description: req.Description,
		AdminUserID: req.AdminUserID,
	}

	// v.Users field is ignored by ORM in any case, and we only need to populate it as part of the Vault instance in Go
	// For now, since this is a new vault, we have only one user (the admin)
	v.Users, err = user.GetUsers(v.AdminUserID)
	if err != nil {
		return nil, err
	}

	// In this case get and assign the ID now so we can use it the vault-user entities
	// Assigning it explicitly means that the ORm library doesn't assign it itself during insert
	v.ID = orm.GetNewID()

	// Set the vault-user
	vu := []*vaultUser{
		&vaultUser{
			VaultID: v.ID,
			UserID:  v.AdminUserID,
		},
	}

	// Insert the vault and the new vault-user mapping
	err = orm.InsertTx(&v, vu[0])
	if err != nil {
		return nil, err
	}

	return &v, nil
}

// GetVault returns the vault object with the given id
func GetVault(ctx context.Context, id orm.ID) (*Vault, error) {
	clog.Debugf("%s: GetVault(): id %v", gServiceName, id)

	var v Vault

	// Get the Vault fields that are stored in the main DB (this does not include v.Users)
	exists, err := orm.GetByID(id, &v)
	if err != nil {
		return nil, err
	}
	if !exists {
		clog.Warnf("%s: no vault found with id %v", gServiceName, id)
		return nil, nil
	}
	if v.ID != id {
		panic(fmt.Sprintf("vault fetched by id (%v) has a different id (%v)", id, v.ID))
	}

	// Populate v.Users: Get vaultUsers first and then get user objects for those userIDs
	vaultUsers, err := getVaultUsersByVaultID(ctx, id)
	if err != nil {
		return nil, err
	}
	for _, vu := range vaultUsers {
		u, err := user.GetUser(vu.UserID)
		if err != nil {
			return nil, err
		}
		v.Users = append(v.Users, u)
	}

	return &v, nil
}

// GetVaultsByUser fetches all vaults that the given user is a part of (even if the user did not create that vault)
func GetVaultsByUser(ctx context.Context, userID orm.ID) ([]*Vault, error) {
	clog.Debugf("%s: GetVaultsByUsers(): user %v", gServiceName, userID)

	// Get all the vaultIDs associated with the user
	vaultUsers, err := getVaultUsersByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("could not get vaultUsers by userID: %v", err)
	}

	// Get all the associated vaults for those vaultIDs
	var vaults = []*Vault{}
	for _, vu := range vaultUsers {
		v, err := GetVault(ctx, vu.VaultID)
		if err != nil {
			return nil, err
		}
		vaults = append(vaults, v)
	}
	clog.Debugf("%s: GetVaultsByUsers(): user %v: returning:\n%+v", gServiceName, userID, vaults)
	return vaults, nil
}

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* H E L P E R S
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

func getVaultUsersByVaultID(ctx context.Context, vaultID orm.ID) ([]*vaultUser, error) {
	clog.Debugf("%s: getVaultUsersByVaultID(): vauldID %v", gServiceName, vaultID)

	var vaultUsers []*vaultUser
	err := orm.FindByColumn("vault_id", vaultID, &vaultUsers)
	if err != nil {
		return nil, err
	}
	return vaultUsers, nil
}

func getVaultUsersByUserID(ctx context.Context, userID orm.ID) ([]*vaultUser, error) {
	clog.Debugf("%s: getVaultUsersByUserID(): userID %v", gServiceName, userID)

	var vaultUsers []*vaultUser
	err := orm.FindByColumn("user_id", userID, &vaultUsers)
	if err != nil {
		return nil, err
	}
	return vaultUsers, nil
}
