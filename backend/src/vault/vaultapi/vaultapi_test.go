package vaultapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/teejays/n-factor-vault/backend/library/go-api/apitest"
	"github.com/teejays/n-factor-vault/backend/src/auth"
	"github.com/teejays/n-factor-vault/backend/src/orm"
	"github.com/teejays/n-factor-vault/backend/src/user/userapi"
	"github.com/teejays/n-factor-vault/backend/src/vault/vaultapi"
)

func TestHandleCreateVault(t *testing.T) {

	// Make sure that we empty any table that these tests might populate once the test is over
	var relevantOrmTables = []string{"UserSecure", "Vault"}
	defer orm.EmptyTestTables(t, relevantOrmTables)

	// Setup Test
	// 1. Create some users
	helperCreateTestUsersT(t)
	// 2. Login a test user and get the JWT token
	token := helperLoginTestUserT(t)
	// 3. Create a func that returns the token, so we can use that function as a param to the TestSuite
	var getAuthTokenFunc = func(t *testing.T) string { return token }

	// Define the Test Suite
	ts := apitest.TestSuite{
		Route:                 "/v1/vault",
		Method:                http.MethodPost,
		HandlerFunc:           vaultapi.HandleCreateVault,
		AuthBearerTokenFunc:   getAuthTokenFunc,
		AuthMiddlewareHandler: auth.AuthenticateRequestMiddleware,
		BeforeTestFunc:        nil,
		AfterTestFunc:         func(t *testing.T) { orm.EmptyTestTables(t, []string{"Vault"}) },
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

func helperCreateTestUsersT(t *testing.T) {
	err := helperCreateJonJane()
	if err != nil {
		t.Error(err)
	}
}

func helperLoginTestUserT(t *testing.T) string {
	token, err := helperLoginJon()
	if err != nil {
		t.Error(err)
	}
	return token
}

// A function to create test users
func helperCreateJonJane() error {

	// Define the Handler Request to signup a user
	p := apitest.HandlerReqParams{
		Route:       "/v1/signup",
		Method:      http.MethodPost,
		HandlerFunc: userapi.HandleSignup,
	}

	// Create Jon
	if _, _, err := p.MakeHandlerRequest(
		`{"name":"Jon Doe", "email":"jon.doe@email.com","password":"jons_secret"}`,
		[]int{http.StatusCreated, http.StatusOK},
	); err != nil {
		return err
	}

	// Create Jane
	if _, _, err := p.MakeHandlerRequest(
		`{"name":"Jane Does", "email":"jane.does@email.com","password":"janes_secret"}`,
		[]int{http.StatusCreated, http.StatusOK},
	); err != nil {
		return err
	}

	return nil

}

func helperLoginJon() (string, error) {
	// Define the Handler Request to signup a user
	p := apitest.HandlerReqParams{
		Route:       "/v1/login",
		Method:      http.MethodPost,
		HandlerFunc: userapi.HandleLogin,
	}

	// Make the Login request
	_, body, err := p.MakeHandlerRequest(
		`{"email":"jon.doe@email.com","password":"jons_secret"}`,
		[]int{http.StatusOK},
	)
	if err != nil {
		return "", err
	}

	// Get JWT Token from response
	var m = make(map[string]interface{})
	if err := json.Unmarshal(body, &m); err != nil {
		return "", err
	}
	tokenX := m["JWT"]
	if tokenX == "" {
		return "", fmt.Errorf("couldn't get JWT token in response")
	}
	token, ok := tokenX.(string)
	if !ok {
		return "", fmt.Errorf("JWT token in response is not of type string")
	}

	return token, nil

}
