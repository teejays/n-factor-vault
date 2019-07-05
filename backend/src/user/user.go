package user

import (
	"fmt"

	"github.com/teejays/clog"
	"github.com/teejays/n-factor-vault/backend/src/orm"
	"github.com/teejays/n-factor-vault/backend/src/user/password"
)

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* O R M   M O D E L
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

type User struct {
	orm.BaseModel `xorm:"extends"`
	Name          string `xorm:"notnull" json:"name"`
	Email         string `xorm:"unique notnull" json:"email"`
}

type UserSecure struct {
	User                    `xorm:"extends"`
	password.SecurePassword `xorm:"extends"`
}

func init() {
	// 1. Setup User ORM Model
	err := orm.SyncModelSchema(&UserSecure{})
	if err != nil {
		clog.FatalErr(err)
	}
}

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* M E T H O D S
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

type CreateUserRequest struct {
	Name     string
	Email    string
	Password string
}

func CreateUser(req CreateUserRequest) (*User, error) {
	var err error

	// TODO: Validate the request is good?
	var u UserSecure
	u.Name = req.Name
	u.Email = req.Email

	// Get the password hash
	u.SecurePassword, err = password.NewSecurePassword(req.Password)
	if err != nil {
		return nil, err
	}

	// TODO: Before we create the user, we should check to
	// make sure that a user with same Email, ID etc. does
	// not exist.
	existingUser, err := getUserByEmail(u.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, fmt.Errorf("an account with this email already exists")
	}

	// Save to DB
	err = orm.InsertOne(&u)
	if err != nil {
		return nil, err
	}

	return &u.User, nil
}

func GetSecureUserByEmail(email string) (*UserSecure, error) {
	return getSecureUserByEmail(email)
}

func getUserByEmail(email string) (*User, error) {
	su, err := getSecureUserByEmail(email)
	if err != nil {
		return nil, err
	}
	if su == nil {
		return nil, nil
	}
	return &su.User, nil
}

func getSecureUserByEmail(email string) (*UserSecure, error) {
	var su UserSecure
	exists, err := orm.GetByColumn("email", email, &su)
	if err != nil {
		return nil, err
	}
	if !exists {
		clog.Warnf("user: no user found with email %s", email)
		return nil, nil
	}
	return &su, nil
}
