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
	"github.com/teejays/n-factor-vault/backend/library/orm"
	"github.com/teejays/n-factor-vault/backend/library/util"

	"github.com/teejays/n-factor-vault/backend/src/user"
)

var gServiceName = "Vault Service"

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* O R M   M O D E L S
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// Vault definers the properties and fields of a vault entity
// TODO: a vault should have hasMany relation with user.User
// DECISION: do we want to explicitly have hasMany relations across tables?
type Vault struct {
	orm.BaseModel `gorm:"embedded"`
	Name          string      `gorm:"unique_index:idx_name_admin" json:"name"`
	Description   string      `json:"description"`
	AdminUserID   id.ID       `gorm:"unique_index:idx_name_admin" json:"admin_user_id"`
	VaultUsers    []VaultUser `json:"vault_users"`
}

// VaultUser represents the mapping between vault and users that are a part of it. This is not exported
// since we want to add this data to the main Vault struct while returning a Vault type, and don't to expose
// by itself.
type VaultUser struct {
	orm.BaseModel `gorm:"embedded"`
	VaultID       id.ID     `gorm:"unique_index:idx_vault_user" json:"vault_id"`
	UserID        id.ID     `gorm:"unique_index:idx_vault_user" json:"user_id"`
	User          user.User `json:"user"`
}

// ShamirsVault represents the encryption structure of a vault
type ShamirsVault struct {
	orm.BaseModel `gorm:"embedded"`
	VaultID       id.ID `gorm:"unique_index:idx_vault" json:"vault_id"`
	N             int   `json:"n"` // total number of people who share the secret
	K             int   `json:"k"` // minimum number required to decrypt the secret
}

// Init initializes the service so it can connect with the ORM
func Init() error {
	return orm.RegisterModels(&Vault{}, &VaultUser{}, &ShamirsVault{})
}

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* M E T H O D S
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

type CreateAndInitializeVaultRequest struct {
	CreateVaultRequest
	CreateShamirVaultRequest
	AddMemberByEmailsToVaultRequest
}

type AddMemberByEmailsToVaultRequest struct {
	VaultID      id.ID
	MemberEmails []string
}

// CreateVaultRequest are the parameters that are passed when creating a vault
type CreateVaultRequest struct {
	AdminUserID id.ID
	Name        string
	Description string
}

type CreateShamirVaultRequest struct {
	VaultID id.ID `json:"vault_id"`
	N       int   `json:"n"`
	K       int   `json:"k"`
}

type AddUserToVaultRequest struct {
	VaultID id.ID `json:"vault_id"`
	UserID  id.ID `json:"user_id"`
}

// CreateVault creates a new vault with the current authenticated user as the admin
func CreateVault(ctx context.Context, req CreateVaultRequest) (*Vault, error) {
	clog.Debugf("vault: creating vault %s", req.Name)
	var err error

	// Validate: Validate the request
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

	// Set the vault-user for the user creating this vault.
	vu := VaultUser{
		UserID: v.AdminUserID,
	}
	v.VaultUsers = []VaultUser{vu}

	// Save the Vault, which will be generate the VaultID
	err = orm.InsertOne(&v)
	if err != nil {
		return nil, err
	}

	return &v, nil
}

// CreateAndInitializeVault creates a new vault with the current authenticated user as the admin
func CreateAndInitializeVault(ctx context.Context, req CreateAndInitializeVaultRequest) (*Vault, error) {
	clog.Debugf("vault: creating vault %s", req.Name)
	var err error

	// Create Vault instance
	v, err := CreateVault(ctx, req.CreateVaultRequest)
	if err != nil {
		return nil, fmt.Errorf("creating vault: %v", err)
	}

	// Validate: Minimum Number of Approvals should be greater than 1
	if req.K < 2 {
		return nil, fmt.Errorf("minimum number of approvals required should be greater than 1")
	}

	// Validate: Member emails should be unique
	errs := util.ValidateUniqueStrings(req.MemberEmails)
	if len(errs) > 0 {
		return nil, fmt.Errorf("%v", errs)
	}
	// Validate: Number of members should not be less than K
	if len(req.MemberEmails) < req.K {
		return nil, fmt.Errorf("number of members should be less than or equal to the minimum number of approvals required")
	}

	// TODO: Everything in here should happen in a single transaction

	// Create vault users for the existing users on this vault
	for _, email := range req.MemberEmails {
		// Get the User object corresponding to the email and make sure that the user exists
		// If it does, add it to the users for the vault
		u, err := user.GetUserByEmail(email)
		if err != nil {
			return nil, err
		}
		if u.ID.IsEmpty() {
			return nil, fmt.Errorf("no user with email %s found", email)
		}
		vu := VaultUser{UserID: u.ID}
		v.VaultUsers = append(v.VaultUsers, vu)
	}

	// Save the Vault, which will add the members
	err = orm.Save(v)
	if err != nil {
		return nil, err
	}

	// Create the instance to store the Shamir's config fot this vault
	var sc = ShamirsVault{
		N:       len(v.VaultUsers),
		K:       req.K,
		VaultID: v.ID,
	}

	err = orm.InsertOne(&sc)
	if err != nil {
		return nil, err
	}

	return v, nil
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

	return &v, nil
}

// GetVaultsByUser fetches all vaults that the given user is a part of (even if the user did not create that vault)
func GetVaultsByUser(ctx context.Context, userID id.ID) ([]*Vault, error) {
	clog.Debugf("%s: GetVaultsByUsers(): user %v", gServiceName, userID)

	// Get all the vaultIDs associated with the user
	VaultUsers, err := getVaultUsersByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("could not get VaultUsers by userID: %v", err)
	}

	// Get all the associated vaults for those vaultIDs
	var vaults = []*Vault{}
	for _, vu := range VaultUsers {
		v, err := GetVault(ctx, vu.VaultID)
		if err != nil {
			return nil, err
		}
		vaults = append(vaults, v)
	}
	clog.Debugf("%s: GetVaultsByUsers(): user %v: returning:\n%+v", gServiceName, userID, vaults)
	return vaults, nil
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
	// TODO: Check if user doesn't exist currently

	// Create a new VaultUser and add to Vault
	var vu = VaultUser{
		UserID:  userID,
		VaultID: v.ID,
	}
	v.VaultUsers = append(v.VaultUsers, vu)

	// // add a new vault user
	// vu, err := addVaultUser(ctx, v.ID, userID)
	// if err != nil {
	// 	return err
	// }
	// // if successfully added VaultUser to the database, append it to this vault users
	// u, err := user.GetUser(vu.UserID)
	// if err != nil {
	// 	return err
	// }
	// v.Users = append(v.Users, u)

	err := orm.Save(v)
	if err != nil {
		return fmt.Errorf("could not save: %v", err)
	}
	return nil
}

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* H E L P E R S
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

func GetVaultUsersByVaultID(ctx context.Context, vaultID id.ID) ([]*VaultUser, error) {
	clog.Debugf("%s: GetVaultUsersByVaultID(): vaultID %v", gServiceName, vaultID)

	var VaultUsers []*VaultUser
	_, err := orm.FindByColumn("vault_id", vaultID, &VaultUsers)
	if err != nil {
		return nil, err
	}
	return VaultUsers, nil
}

func getVaultUsersByUserID(ctx context.Context, userID id.ID) ([]*VaultUser, error) {
	clog.Debugf("%s: getVaultUsersByUserID(): userID %v", gServiceName, userID)

	var VaultUsers []*VaultUser
	var whereConds = map[string]interface{}{
		"user_id": userID,
	}
	_, err := orm.Find(whereConds, &VaultUsers)
	if err != nil {
		return nil, err
	}
	return VaultUsers, nil
}

func addVaultUser(ctx context.Context, vaultID, userID id.ID) (*VaultUser, error) {
	clog.Debugf("%s: addUserToVault(): vaultID <%v> | userID <%v>", gServiceName, userID)

	if userID.IsEmpty() {
		return nil, fmt.Errorf("userID is empty")
	}

	if vaultID.IsEmpty() {
		return nil, fmt.Errorf("vaultID is empty")
	}

	var vu = VaultUser{
		VaultID: vaultID,
		UserID:  userID,
		// TODO: There should be some concept of confirming users into a vault. One they are invited
		// or request to join, and once they accept or the request is approved, the should be confirmed.
	}

	err := orm.InsertOne(&vu)
	if err != nil {
		return nil, err
	}

	return &vu, nil
}
