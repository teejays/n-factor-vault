package handler_test

import (
	"net/http"
	"testing"

	"github.com/teejays/clog"
	"github.com/teejays/n-factor-vault/backend/library/go-api/apitest"
	"github.com/teejays/n-factor-vault/backend/src/auth"
	"github.com/teejays/n-factor-vault/backend/src/orm"
	"github.com/teejays/n-factor-vault/backend/src/server/handler"
)

func init() {
	clog.LogLevel = 0
}

func TestHandleCreateVault(t *testing.T) {

	// Make sure that we empty any table that these tests might populate once the test is over
	var relevantOrmTables = []string{"UserSecure", "Vault", "VaultUser"}
	defer orm.EmptyTestTables(t, relevantOrmTables)

	// Setup Test
	// 1. Create some users
	helperCreateTestUsersT(t)
	// 2. Login a test user and get the JWT token
	token, _ := helperLoginTestUsersT(t)
	// 3. Create a func that returns the token, so we can use that function as a param to the TestSuite
	var getAuthTokenFunc = func(t *testing.T) string { return token }

	// Define the Test Suite
	ts := apitest.TestSuite{
		Route:                 "/v1/vault",
		Method:                http.MethodPost,
		HandlerFunc:           handler.HandleCreateVault,
		AuthBearerTokenFunc:   getAuthTokenFunc,
		AuthMiddlewareHandler: auth.AuthenticateRequestMiddleware,
		BeforeTestFunc:        nil,
		AfterTestFunc:         func(t *testing.T) { orm.EmptyTestTables(t, []string{"Vault", "VaultUser"}) },
		// ^AfterTestFunc: we should empty the vault table after each test to start the next run on a fresh slate
	}

	_ = `{"name":"Facebook", "description":"Shared account for our org"}`

	// Define the individual tests
	tests := []apitest.HandlerTest{
		{
			Name:           "status OK if request has valid content",
			Content:        `{"name":"Facebook", "description":"Shared account for our org"}`,
			WantStatusCode: http.StatusCreated,
			WantContent:    "",
			WantErrMessage: "",
			AssertContentFields: map[string]apitest.AssertFunc{
				"id":          apitest.AssertNotEmptyFunc,
				"name":        apitest.AssertIsEqual("Facebook"),
				"description": apitest.AssertIsEqual("Shared account for our org"),
				"created_at":  apitest.AssertNotEmptyFunc,
				"updated_at":  apitest.AssertNotEmptyFunc,
				"users":       apitest.AssertNotEmptyFunc,
			},
			SkipAfterTestFunc: true,
		},
		{
			// In the last test above, we set teh flag to skip AfterRunFunc, which means that the DB will not cleared
			Name:           "status BadRequest if a vault with same name already exists",
			Content:        `{"name":"Facebook", "description":"a different desc than before"}`,
			WantStatusCode: http.StatusBadRequest,
			WantErrMessage: "duplicate key value violates unique constraint",
		},
		{
			Name:           "status Unauthorized if request has no auth token",
			Content:        `{"name":"Facebook", "description":"Shared account for our org"}`,
			SkipAuthToken:  true,
			WantStatusCode: http.StatusUnauthorized,
		},
		{
			Name:                 "status Unauthorized if request has a bad auth token",
			Content:              `{"name":"Facebook", "description":"Shared account for our org"}`,
			AuthBeaererTokenFunc: func(t *testing.T) string { return "jkkjhkjasdkjh.oijowqieoij.12lkjadlkj" }, // Bad Token
			WantStatusCode:       http.StatusUnauthorized,
		},
		{
			Name:           "status BadRequest if request with empty content",
			Content:        "",
			WantStatusCode: http.StatusBadRequest,
			WantErrMessage: "no content provided with the HTTP request",
		},
		{
			Name:           "status BadRequest if request is not a valid JSON",
			Content:        "I am a non-JSON content",
			WantStatusCode: http.StatusBadRequest,
			WantErrMessage: "content is not a valid JSON",
		},
		{
			Name:           "status BadRequest if request does not include required fields",
			Content:        `{}`,
			WantStatusCode: http.StatusBadRequest,
			WantContent:    "",
			WantErrMessage: "empty",
		},
		{
			Name:           "status BadRequest if request if name is empty",
			Content:        `{"name":"", "description":"a different desc than before"}`,
			WantStatusCode: http.StatusBadRequest,
			WantErrMessage: "name is empty",
		},
		{
			Name:           "status BadRequest if request if description is empty",
			Content:        `{"name":"Facebook"}`,
			WantStatusCode: http.StatusBadRequest,
			WantErrMessage: "description is empty",
		},
	}

	ts.RunHandlerTests(t, tests)

}

func TestHandleGetVaults(t *testing.T) {

	// Make sure that we empty any table that these tests might populate once the test is over
	var relevantOrmTables = []string{"UserSecure", "Vault", "VaultUser"}
	orm.EmptyTestTables(t, relevantOrmTables)
	defer orm.EmptyTestTables(t, relevantOrmTables)

	// Setup Test
	// 1. Create some users
	helperCreateTestUsersT(t)
	// 2. Login a test user and get the JWT token
	token1, token2 := helperLoginTestUsersT(t)
	// 3. Create tests vaulst for user
	helperCreateTestVaultsT(t, token1)

	// 4. Create a func that returns the token, so we can use that function as a param to the TestSuite
	var getAuthTokenFunc = func(t *testing.T) string { return token1 }

	// Define the Test Suite
	ts := apitest.TestSuite{
		Route:                 "/v1/vault",
		Method:                http.MethodGet,
		HandlerFunc:           handler.HandleGetVaults,
		AuthBearerTokenFunc:   getAuthTokenFunc,
		AuthMiddlewareHandler: auth.AuthenticateRequestMiddleware,
	}

	// Define the individual tests
	// TODO: we're only checking for http status code since we have no good way of asserting that the content (array) is what we need
	tests := []apitest.HandlerTest{
		{
			Name:           "status Unauthorized if request has no auth token",
			Content:        `{"name":"Facebook", "description":"Shared account for our org"}`,
			SkipAuthToken:  true,
			WantStatusCode: http.StatusUnauthorized,
		},
		{
			Name:                 "status Unauthorized if request has a bad auth token",
			Content:              `{"name":"Facebook", "description":"Shared account for our org"}`,
			AuthBeaererTokenFunc: func(t *testing.T) string { return "jkkjhkjasdkjh.oijowqieoij.12lkjadlkj" }, // Bad Token
			WantStatusCode:       http.StatusUnauthorized,
		},
		{
			Name:               "status OK if user has vaults",
			WantStatusCode:     http.StatusOK,
			AssertContentFuncs: []apitest.AssertFunc{apitest.AssertIsSlice, apitest.AssertSliceOfLen(2)},
		},
		{
			Name:                 "status Ok but empty response if user has no vaults",
			WantStatusCode:       http.StatusOK,
			AuthBeaererTokenFunc: func(t *testing.T) string { return token2 }, // this is a token for user with no vaults
			AssertContentFuncs:   []apitest.AssertFunc{apitest.AssertIsSlice, apitest.AssertSliceOfLen(0)},
		},
	}

	ts.RunHandlerTests(t, tests)

}
