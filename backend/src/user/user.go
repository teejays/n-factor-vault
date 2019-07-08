package user

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/teejays/clog"
	pwd "github.com/teejays/n-factor-vault/backend/library/go-pwd"
	"github.com/teejays/n-factor-vault/backend/src/orm"
)

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
* O R M   M O D E L S
* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

type User struct {
	orm.BaseModel `xorm:"extends"`
	Name          string `xorm:"notnull" json:"name"`
	Email         string `xorm:"unique notnull" json:"email"`
}

type UserSecure struct {
	User               `xorm:"extends"`
	pwd.SecurePassword `xorm:"extends"`
}

func init() {
	// 1. Setup User ORM Model
	err := orm.RegisterModel(&UserSecure{})
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

func (r CreateUserRequest) Validate() error {
	var errs []error

	var empty []string
	if strings.TrimSpace(r.Name) == "" {
		empty = append(empty, "name")
	}
	if strings.TrimSpace(r.Email) == "" {
		empty = append(empty, "email")
	}
	if strings.TrimSpace(r.Password) == "" {
		empty = append(empty, "password")
	}
	if len(empty) > 0 {
		err := fmt.Errorf("empty fields (%s) provided", strings.Join(empty, ", "))
		errs = append(errs, err)
	}

	emailRegexp := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	if !emailRegexp.MatchString(r.Email) {
		errs = append(errs, fmt.Errorf("email address has an invalid format"))
	}

	// Step 2a: Return nil if no issues found
	if len(errs) < 1 {
		return nil
	}
	// Step 2b: Combine the errors & return
	errMessage := fmt.Sprintf("request has %d issue(s):", len(errs))
	for i, e := range errs {
		errMessage = errMessage + fmt.Sprintf("\n[%d] %v", i+1, e)
	}
	return fmt.Errorf(errMessage)

}

// CreateUser creates a new user
func CreateUser(req CreateUserRequest) (*User, error) {
	var err error

	// Validate the request is good?
	err = req.Validate()
	if err != nil {
		return nil, err
	}

	var u UserSecure
	u.Name = req.Name
	u.Email = req.Email

	// Get the password hash
	u.SecurePassword, err = pwd.NewSecurePassword(req.Password)
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

// GetUser provides the single user with the ID id
func GetUser(id orm.ID) (*User, error) {
	var su UserSecure
	exists, err := orm.GetByID(id, &su)
	if err != nil {
		return nil, err
	}
	if !exists {
		clog.Warnf("user: no user found with id %v", id)
		return nil, nil
	}
	if su.ID != id {
		panic(fmt.Sprintf("user fetched by id (%v) has a different id (%v)", id, su.ID))
	}
	return &su.User, nil
}

// GetUsers returns an slice of users given the userIDs passed
func GetUsers(ids ...orm.ID) ([]*User, error) {
	var users []*User
	for _, id := range ids {
		u, err := GetUser(id)
		if err != nil {
			return nil, fmt.Errorf("userID %v: %v", id, err)
		}
		users = append(users, u)
	}
	return users, nil
}

// GetSecureUserByEmail returns the full User object, including the password info (iteration count, hash, salt etc.)
// This function is exported only because the auth service needs this info to validate the password when a user is logging in.
// DISCUSS: Ideally, this info shouldn't travel between services - so should we do the password validation here? Or should we
// actually store these secure credentials as part of the auth service/database?
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
