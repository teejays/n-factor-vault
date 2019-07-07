package server_test

import (
	"net/http"
	"testing"

	"github.com/teejays/clog"
	"github.com/teejays/n-factor-vault/backend/library/go-api/apitest"
	"github.com/teejays/n-factor-vault/backend/src/orm"
	"github.com/teejays/n-factor-vault/backend/src/server"
)

func EmptyTestTables(t *testing.T, tables []string) {
	if err := orm.EmptyTables(tables); err != nil {
		t.Fatalf("error emptying tables: %v", err)
	}
}

func init() {
	clog.LogLevel = 0
}

func TestHandleSignup(t *testing.T) {

	// There are two ways to make sure that while we run tests, the data in the database is not actually
	// persisted, otherwise it can affect the remaining tests.
	// 1. After each test, or run of a test, explicitly empty any relevant SQL table that the test might've written to
	// 2. Use orm.TestSession (usage shown below), which is like a transaction but makes sure that the transaction is
	// not committed. The problem with this approach is that some fields such as 'created_at', 'updated_at' etc. are
	// only populated in the Go instance of a struct once the struct is inserted/committed into the DB.
	// Exmample user of method (2):
	//
	// if err := orm.StartTestSession(); err != nil {
	// 	t.Errorf("could not start orm session: %v", err)
	// }
	// defer func() {
	// 	if err := orm.EndTestSession(); err != nil {
	// 		t.Errorf("could not end orm session: %v", err)
	// 	}
	// }()

	var relevantOrmTables = []string{"user_secure"}
	defer EmptyTestTables(t, relevantOrmTables)

	ts := apitest.TestSuite{
		Route:         "/v1/signup",
		Method:        http.MethodPost,
		HandlerFunc:   server.HandleSignup,
		AfterTestFunc: func(t *testing.T) { EmptyTestTables(t, relevantOrmTables) },
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
			Name:           "status BadRequest if request does not include email and password",
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
			WantStatusCode: http.StatusOK,
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
