// Package vault implements the main entity of this app - a vault. Any particular site/topic
// for which secure information is stored is called a vault. A vault is owned by one or more user
// and can have it's opening security customizable.
package vault

import (
	"context"
	"fmt"
	"strings"

	"github.com/teejays/clog"
	"github.com/teejays/n-factor-vault/backend/library/id"
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
	orm.BaseModel `gorm:"embedded"`
	Name          string
	Description   string
	UserID        id.ID
	Users         []*user.User
}

// vaultUser represents the mapping between vault and users that are a part of it. This is not exported
// since we want to add this data to the main Vault struct while returning a Vault type, and don't to expose
// by itself.
type vaultUser struct {
	orm.BaseModel `gorm:"embedded"`
	VaultID       id.ID
	UserID        id.ID
	IsConfirmed   bool
}

func init() {
	err := orm.AutoMigrate(&Vault{})
	if err != nil {
		clog.FatalErr(err)
	}

	err = orm.AutoMigrate(&vaultUser{})
	if err != nil {
		clog.FatalErr(err)
	}
}

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* M E T H O D S
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// CreateVaultRequest are the parameters that are passed when creating a vault
type CreateVaultRequest struct {
	UserID      id.ID
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
	if req.UserID.IsEmpty() {
		return nil, fmt.Errorf("admin userID is empty")
	}

	// Create a vault instance
	v := Vault{
		Name:        req.Name,
		Description: req.Description,
		UserID:      req.UserID,
	}

	// v.Users field is ignored by ORM in any case, and we only need to populate it as part of the Vault instance in Go
	// For now, since this is a new vault, we have only one user (the admin)
	v.Users, err = user.GetUsers(v.UserID)
	if err != nil {
		return nil, err
	}

	// In this case get and assign the ID now so we can use it the vault-user entities
	// Assigning it explicitly means that the ORm library doesn't assign it itself during insert
	v.ID = id.GetNewID()

	//TODO: Figure out associationo
	err = orm.InsertOne(&v)
	if err != nil {
		return nil, err
	}

	// Set the vault-user for the user creating this vault. Since this user is the admin,
	// we can assume that their relation to the vault is 'confirmed'
	vu := vaultUser{
		VaultID:     v.ID,
		UserID:      v.UserID,
		IsConfirmed: true,
	}

	err = orm.InsertOne(&vu)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

// GetVault returns the vault object with the given id
func GetVault(ctx context.Context, id id.ID) (*Vault, error) {
	clog.Debugf("%s: GetVault(): id %v", gServiceName, id)

	var v Vault

	// Get the Vault fields that are stored in the main DB (this does not include v.Users)
	exists, err := orm.FindByID(id, &v)
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
	vaultUsers, err := GetVaultUsersByVaultID(ctx, id)
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
func GetVaultsByUser(ctx context.Context, userID id.ID) ([]*Vault, error) {
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

type AddUserToVaultRequest struct {
	UserID  id.ID `json:"user_id"`
	VaultID id.ID
}

func AddUserToVault(ctx context.Context, req AddUserToVaultRequest) (*Vault, error) {
	clog.Debugf("%s: AddUserToVault(ctx, req): req:\n%+v", gServiceName, req)

	// TODO: Make sure that the user making a request to add someone to a vault
	// actually has rights/access to add someone to this vault
	// i.e. at least check that the user is a part of the vault itself

	// Get the vault
	v, err := GetVault(ctx, req.VaultID)
	if err != nil {
		return nil, err
	}
	// Add user to the vault
	err = v.AddUser(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// AddUser adds a new user to the vault
func (v *Vault) AddUser(ctx context.Context, userID id.ID) error {
	// add a new vault user
	vu, err := addVaultUser(ctx, v.ID, userID)
	if err != nil {
		return err
	}
	// if successfully added vaultUser to the database, append it to this vault users
	u, err := user.GetUser(vu.UserID)
	if err != nil {
		return err
	}
	v.Users = append(v.Users, u)
	return nil
}

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* H E L P E R S
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

func GetVaultUsersByVaultID(ctx context.Context, vaultID id.ID) ([]*vaultUser, error) {
	clog.Debugf("%s: GetVaultUsersByVaultID(): vaultID %v", gServiceName, vaultID)

	var vaultUsers []*vaultUser
	_, err := orm.FindByColumn("vault_id", vaultID, &vaultUsers)
	if err != nil {
		return nil, err
	}
	return vaultUsers, nil
}

func getVaultUsersByUserID(ctx context.Context, userID id.ID) ([]*vaultUser, error) {
	clog.Debugf("%s: getVaultUsersByUserID(): userID %v", gServiceName, userID)

	var vaultUsers []*vaultUser
	var whereConds = map[string]interface{}{
		"user_id":      userID,
		"is_confirmed": true,
	}
	_, err := orm.Find(whereConds, &vaultUsers)
	if err != nil {
		return nil, err
	}
	return vaultUsers, nil
}

func addVaultUser(ctx context.Context, vaultID, userID id.ID) (*vaultUser, error) {
	clog.Debugf("%s: addUserToVault(): vaultID <%v> | userID <%v>", gServiceName, userID)

	if userID.IsEmpty() {
		return nil, fmt.Errorf("userID is empty")
	}

	if vaultID.IsEmpty() {
		return nil, fmt.Errorf("vaultID is empty")
	}

	var vu = vaultUser{
		VaultID: vaultID,
		UserID:  userID,
		// TODO: We shouldn't be confirming users into a vault by default. In reality, they should be invited
		// or should request to join, and once they accept or the request is approved, the should be confirmed.
		IsConfirmed: true,
	}

	err := orm.InsertOne(&vu)
	if err != nil {
		return nil, err
	}

	return &vu, nil
}
