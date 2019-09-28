package user

import (
	"fmt"

	"github.com/teejays/clog"

	pwd "github.com/teejays/n-factor-vault/backend/library/go-pwd"
	"github.com/teejays/n-factor-vault/backend/library/id"
	"github.com/teejays/n-factor-vault/backend/library/orm"
)

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* O R M   M O D E L S
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// User is the basic user object
type User struct {
	orm.BaseModel
	Name  string
	Email string
}

// Password is separate struct for storing user hashed passwords
type Password struct {
	orm.BaseModel
	UserID id.ID `gorm:"unique_index:idx_user"`
	pwd.SecurePassword
}

// Init initializes the service so it can connect with the ORM
func Init() (err error) {
	return orm.RegisterModels(&User{}, &Password{})
}

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* M E T H O D S
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

type CreateUserRequest struct {
	Name     string `validate:"required,notblank"`
	Email    string `validate:"required,email,notblank"`
	Password string `validate:"required,notblank"`
}

// CreateUser creates a new user
func CreateUser(req CreateUserRequest) (*User, error) {
	var err error
	var u User

	u.Name = req.Name
	u.Email = req.Email

	// Get the password hash
	var pass Password
	pass.SecurePassword, err = pwd.NewSecurePassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Before we create the user, we should check to make sure that a user with same Email, ID etc. does not exist.
	existingUser, err := getUserByEmail(u.Email)
	if err != nil {
		return nil, err
	}
	if !existingUser.IsEmpty() {
		return nil, fmt.Errorf("an account with this email already exists")
	}

	// Save to DB (the ID will auto-populated)
	err = orm.InsertOne(&u)
	if err != nil {
		return nil, err
	}

	pass.UserID = u.ID
	err = orm.InsertOne(&pass)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

// GetUsers returns an slice of users given the userIDs passed
func GetUsers(ids ...id.ID) ([]User, error) {
	var users []User
	for _, id := range ids {
		u, err := getUser(id)
		if err != nil {
			return nil, fmt.Errorf("userID %v: %v", id, err)
		}
		users = append(users, u)
	}
	return users, nil
}

// GetUser provides the single user with the ID id
func GetUser(id id.ID) (User, error) {
	return getUser(id)
}

func getUser(id id.ID) (User, error) {
	var u User
	exists, err := orm.FindByID(id, &u)
	if err != nil {
		return u, err
	}
	if !exists {
		clog.Warnf("user: no user found with id %v", id)
		return u, nil
	}
	if u.ID != id {
		panic(fmt.Sprintf("user fetched by id (%v) has a different id (%v)", id, u.ID))
	}
	return u, nil
}

func GetUserByEmail(email string) (User, error) {
	return getUserByEmail(email)
}

func getUserByEmail(email string) (User, error) {
	var u User
	exists, err := orm.FindByColumn("email", email, &u)
	if err != nil {
		return u, err
	}
	if !exists {
		clog.Warnf("user: no user found with email %s", email)
		return u, nil
	}
	return u, nil
}

// GetPasswordForUser returns the full User object, including the password info (iteration count, hash, salt etc.)
// This function is exported only because the auth service needs this info to validate the password when a user is logging in.
// DISCUSS: Ideally, this info shouldn't travel between services - so should we do the password validation here? Or should we
// actually store these secure credentials as part of the auth service/database?
func GetPasswordForUser(u User) (Password, error) {
	return getPasswordForUser(u)
}

func getPasswordForUser(u User) (Password, error) {

	var pass Password
	exists, err := orm.FindByColumn("user_id", u.ID, &pass)
	if err != nil {
		return pass, err
	}
	if !exists {
		// This means a user exists but no corresponding password hash, which should never happen
		panic(fmt.Sprintf("no password entry for user %s found", u.ID))
	}

	return pass, nil

}
