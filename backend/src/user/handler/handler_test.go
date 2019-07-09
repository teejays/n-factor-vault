package handler_test

import (
	"net/http"
	"testing"

	"github.com/teejays/clog"
	"github.com/teejays/n-factor-vault/backend/library/go-api/apitest"
	"github.com/teejays/n-factor-vault/backend/src/orm"
	userhandler "github.com/teejays/n-factor-vault/backend/src/user/handler"
)

func init() {
	clog.LogLevel = 8
}

// There are two ways to make sure that while we run tests, the data in the database is not actually
// persisted, otherwise it can affect the remaining tests.
// 1. After each test, or run of a test, explicitly empty any relevant SQL table that the test might've written to
// 2. Use orm.TestSession (usage shown below), which is like a transaction but makes sure that the transaction is
// not committed. The problem with this approach is that some fields such as 'created_at', 'updated_at' etc. are
// only populated in the Go instance of a struct once the struct is inserted/committed into the DB.
//
// Example (Method 1):
//
//  var relevantOrmTables = []string{"user_secure"}
// 	defer orm.EmptyTestTables(t, relevantOrmTables)
//
// Exmample (Method 2):
//
// if err := orm.StartTestSession(); err != nil {
// 	t.Errorf("could not start orm session: %v", err)
// }
// defer func() {
// 	if err := orm.EndTestSession(); err != nil {
// 		t.Errorf("could not end orm session: %v", err)
// 	}
// }()
//

func TestHandleSignup(t *testing.T) {

	var relevantOrmObjs = []string{"UserSecure"}
	defer orm.EmptyTestTables(t, relevantOrmObjs)

	ts := apitest.TestSuite{
		Route:         "/v1/signup",
		Method:        http.MethodPost,
		HandlerFunc:   userhandler.HandleSignup,
		AfterTestFunc: func(t *testing.T) { orm.EmptyTestTables(t, relevantOrmObjs) },
	}

	tests := []apitest.HandlerTest{
		{
			Name:           "status BadRequest if request with empty content",
			Content:        "",
			WantStatusCode: http.StatusBadRequest,
			WantContent:    "",
			WantErrMessage: "no content provided with the HTTP request",
		},
		{
			Name:           "status BadRequest if request is not a valid JSON",
			Content:        "I am a non-JSON content",
			WantStatusCode: http.StatusBadRequest,
			WantContent:    "",
			WantErrMessage: "content is not a valid JSON",
		},
		{
			Name:           "status BadRequest if request does not include name, email and password",
			Content:        `{}`,
			WantStatusCode: http.StatusBadRequest,
			WantContent:    "",
			WantErrMessage: "empty fields (name, email, password) provided",
		},
		{
			Name:           "status BadRequest if request if name is missing",
			Content:        `{"email":"email@email.com", "password":"secret"}`,
			WantStatusCode: http.StatusBadRequest,
			WantContent:    "",
			WantErrMessage: "empty fields (name) provided",
		},
		{
			Name:           "status BadRequest if request if email is missing",
			Content:        `{"name":"Jon Doe", "password":"secret"}`,
			WantStatusCode: http.StatusBadRequest,
			WantContent:    "",
			WantErrMessage: "empty fields (email) provided",
		},
		{
			Name:           "status BadRequest if request if password is missing",
			Content:        `{"name":"Jon Doe", "email":"email@email.com"}`,
			WantStatusCode: http.StatusBadRequest,
			WantContent:    "",
			WantErrMessage: "empty fields (password) provided",
		},
		{
			Name:           "status BadRequest if email is not a valid email",
			Content:        `{"name":"Jon Doe", "email":"email.email.com", "password":"secret"}`,
			WantStatusCode: http.StatusBadRequest,
			WantContent:    "",
			WantErrMessage: "email address has an invalid format",
		},
		{
			Name:           "status OK if request has valid name, email, and password",
			Content:        `{"name":"Jon Doe", "email":"email@email.com", "password":"secret"}`,
			WantStatusCode: http.StatusCreated,
			WantContent:    "",
			WantErrMessage: "",
			AssertContentFields: map[string]apitest.AssertFunc{
				"id":         apitest.AssertNotEmptyFunc,
				"name":       apitest.AssertIsEqual("Jon Doe"),
				"email":      apitest.AssertIsEqual("email@email.com"),
				"created_at": apitest.AssertNotEmptyFunc,
				"updated_at": apitest.AssertNotEmptyFunc,
			},
		},
	}

	ts.RunHandlerTests(t, tests)

}

func TestHandleLogin(t *testing.T) {

	// Make sure that we empty any table that these tests might populate too
	var relevantOrmObjs = []string{"UserSecure"}
	defer orm.EmptyTestTables(t, relevantOrmObjs)

	// Define the Test Suite
	ts := apitest.TestSuite{
		Route:          "/v1/login",
		Method:         http.MethodPost,
		HandlerFunc:    userhandler.HandleLogin,
		BeforeTestFunc: helperCreateTestUsersT,
		AfterTestFunc:  func(t *testing.T) { orm.EmptyTestTables(t, relevantOrmObjs) },
	}

	// Define the individual tests
	tests := []apitest.HandlerTest{
		{
			Name:           "status OK if request has valid credentials - 1",
			Content:        `{"email":"jon.doe@email.com", "password":"jons_secret"}`,
			WantStatusCode: http.StatusOK,
			WantContent:    "",
			WantErrMessage: "",
			AssertContentFields: map[string]apitest.AssertFunc{
				"JWT": apitest.AssertNotEmptyFunc,
			},
		},
		{
			Name:           "status OK if request has valid credentials - 2",
			Content:        `{"email":"jane.does@email.com", "password":"janes_secret"}`,
			WantStatusCode: http.StatusOK,
			WantContent:    "",
			WantErrMessage: "",
			AssertContentFields: map[string]apitest.AssertFunc{
				"JWT": apitest.AssertNotEmptyFunc,
			},
		},
		{
			Name:           "status BadRequest if request with empty content",
			Content:        "",
			WantStatusCode: http.StatusBadRequest,
			WantContent:    "",
			WantErrMessage: "no content provided with the HTTP request",
		},
		{
			Name:           "status BadRequest if request is not a valid JSON",
			Content:        "I am a non-JSON content",
			WantStatusCode: http.StatusBadRequest,
			WantContent:    "",
			WantErrMessage: "content is not a valid JSON",
		},
		// TODO: an empty body should produce a more revealing error message
		{
			Name:           "status BadRequest if request does not include email and password",
			Content:        `{}`,
			WantStatusCode: http.StatusBadRequest,
			WantContent:    "",
			WantErrMessage: "no email provided",
		},
		{
			Name:           "status BadRequest if request if email is missing",
			Content:        `{"password":"secret"}`,
			WantStatusCode: http.StatusBadRequest,
			WantContent:    "",
			WantErrMessage: "no email provided",
		},
		{
			Name:           "status BadRequest if request if password is missing",
			Content:        `{"email":"email@email.com"}`,
			WantStatusCode: http.StatusBadRequest,
			WantContent:    "",
			WantErrMessage: "no password provided",
		},
		{
			Name:           "status BadRequest if email is of an invalid format",
			Content:        `{"email":"email.email.com", "password":"secret"}`,
			WantStatusCode: http.StatusBadRequest,
			WantContent:    "",
			WantErrMessage: "login credentials are invalid",
		},
		{
			Name:           "status BadRequest if email is not of a signed up user",
			Content:        `{"email":"jack.die@email.com", "password":"secret"}`,
			WantStatusCode: http.StatusBadRequest,
			WantContent:    "",
			WantErrMessage: "login credentials are invalid",
		},
		{
			Name:           "status BadRequest if password is wrong for a valid email",
			Content:        `{"email":"jon.doe@email.com", "password":"wrong password"}`,
			WantStatusCode: http.StatusBadRequest,
			WantContent:    "",
			WantErrMessage: "login credentials are invalid",
		},
	}

	ts.RunHandlerTests(t, tests)

}

func helperCreateTestUsersT(t *testing.T) {
	err := helperCreateTestUsers()
	if err != nil {
		t.Error(err)
	}
}

// A function to create test users
func helperCreateTestUsers() error {

	// Define the Handler Request to signup a user
	p := apitest.HandlerReqParams{
		Route:       "/v1/signup",
		Method:      http.MethodPost,
		HandlerFunc: userhandler.HandleSignup,
	}

	// Create User 1
	if _, _, err := p.MakeHandlerRequest(
		`{"name":"Jon Doe", "email":"jon.doe@email.com","password":"jons_secret"}`,
		[]int{http.StatusCreated, http.StatusOK},
	); err != nil {
		return err
	}

	// Create User 2
	if _, _, err := p.MakeHandlerRequest(
		`{"name":"Jane Does", "email":"jane.does@email.com","password":"janes_secret"}`,
		[]int{http.StatusCreated, http.StatusOK},
	); err != nil {
		return err
	}

	return nil

}
